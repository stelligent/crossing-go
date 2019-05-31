package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3crypto"
	yaml "gopkg.in/yaml.v2"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/clunaslunas/crossing-go/crypto"

	"github.com/aws/aws-sdk-go/service/s3"
)

// envCmd represents the env command
var envCmd = &cobra.Command{
	Use:   "env S3URL",
	Short: "Retrieve an object from S3 and export as environment variables",
	Long: `Downloads an S3 object, using Client Side Encryption (CSE)
to decrypt it securely. Then parses as YAML and prints out export statements
suitable for use in a shell script.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		_, _, err := parseS3Url(args[0])
		if err != nil {
			return fmt.Errorf("invalid S3 URL: %s", args[0])
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {

		s3bucket, s3object, err := parseS3Url(args[0])

		if err != nil {
			cmd.UsageFunc()(cmd)
			os.Exit(1)
		}

		yamlByteArray, err := sGetS3Cse(s3bucket, s3object)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		m := make(map[interface{}]interface{})
		err = yaml.Unmarshal(yamlByteArray, &m)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		walkMap(m, viper.GetString("prefix"))
	},
}

func sGetS3Cse(s3bucket, s3object string) ([]byte, error) {
	// fmt.Println("getS3 bucket:" + s3bucket + " object:" + s3object + " dest:" + filedest)
	// cmkID := "_unused_get_kms_key_"
	params := &s3.GetObjectInput{
		Bucket: &s3bucket,
		Key:    &s3object}
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	// Create the KeyProvider
	// handler := s3crypto.NewKMSKeyGenerator(kms.New(sess), cmkID)
	// HeaderV2LoadStrategy
	svc := s3crypto.NewDecryptionClient(sess)
	svc.CEKRegistry[crosscrypto.AESCBCPKCS5Padding] = crosscrypto.NewAESCBCContentCipher

	// resp, err := svc.S3Client.GetObject(params)
	resp, err := svc.GetObject(params)
	if err != nil {
		fmt.Println("Error in fetch!")
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err)
		return nil, err
	}

	objectByteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return objectByteArray, nil
}

func init() {
	RootCmd.AddCommand(envCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// envCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// envCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	envCmd.Flags().StringP("prefix", "p", "", "String to prefix environment variables with")
	viper.BindPFlag("prefix", envCmd.Flags().Lookup("prefix"))
}

func walkInterfaceArray(in []interface{}, path string) []interface{} {
	res := make([]interface{}, len(in))
	if path != "" {
		path = path + "__"
	}
	for i, v := range in {
		res[i] = walkMap(v, path+strconv.Itoa(i))
	}
	return res
}

func walkInterfaceMap(in map[interface{}]interface{}, path string) map[string]interface{} {
	res := make(map[string]interface{})
	if path != "" {
		path = path + "__"
	}
	for k, v := range in {
		res[fmt.Sprintf("%v", k)] = walkMap(v, path+fmt.Sprintf("%v", k))
	}
	return res
}

func walkMap(v interface{}, path string) interface{} {
	switch v := v.(type) {
	case []interface{}:
		return walkInterfaceArray(v, path)
	case map[interface{}]interface{}:
		return walkInterfaceMap(v, path)
	case string:
		fmt.Printf("export %s=\"%s\"\n", path, v)
		return v
	default:
		fmt.Printf("export %s=\"%v\"\n", path, v)
		return fmt.Sprintf("%v", v)
	}
}
