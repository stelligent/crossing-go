package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/stelligent/crossing-go/clientfactory"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3crypto"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	crosscrypto "github.com/stelligent/crossing-go/crypto"
)

// putCmd represents the put command
var putCmd = &cobra.Command{
	Use:   "put SOURCE S3URL",
	Short: "Upload a file to S3",
	Long: `Using Client Side Encryption (CSE), encrypt and upload
a file to S3.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(2)(cmd, args); err != nil {
			return err
		}
		_, _, err := parseS3Url(args[1])
		if err != nil {
			return fmt.Errorf("invalid S3 URL: %s", args[1])
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		sourceFile := args[0]
		s3bucket, s3object, err := parseS3Url(args[1])

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
			os.Exit(1)
		}
		// If destination is "" or /
		// assume passed bare bucket and generate key
		// from source filename
		if s3object == "" || s3object == "/" {
			s3object = sourceFile
		}
		// If object key ends with /, assume key is a
		// key prefix and append on the source file
		if s3object[len(s3object)-1:] == "/" {
			s3object = s3object + sourceFile
		}

		//Return an encryption client from the global session
		cmkID := viper.GetString("kmskeyid")
		newSess := viper.Get("ClientSess").(*session.Session)
		//Create the KeyProvider
		handler := s3crypto.NewKMSKeyGenerator(kms.New(newSess), cmkID)

		encryptionclient := Put{
			Client: clientfactory.NewEncryptyionClient(newSess, s3crypto.AESCBCContentCipherBuilder(handler, crosscrypto.NewPKCS7Padder(16))).S3Client,
			Bucket: s3bucket,
			Key:    s3object,
			Source: sourceFile,
		}

		versionID, err := encryptionclient.putS3Cse()
		flagBool := viper.GetBool("verboseoutput")

		if err != nil {
			fmt.Fprintf(os.Stderr, "err uploading file: %s\n", err)
			os.Exit(1)
		}
		if flagBool {
			fmt.Fprintf(os.Stdout, "{ \"VersionId\": %s }\n", string(versionID))
		}

	},
}

// Put provides the ability to put objects
type Put struct {
	Client s3iface.S3API
	Bucket string
	Key    string
	Source string
}

func (p *Put) putS3Cse() ([]byte, error) {

	result, err := p.Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(p.Bucket),
		Key:    aws.String(p.Key),
		Body:   aws.ReadSeekCloser(strings.NewReader(p.Source)),
	})

	if err != nil {
		fmtErr := fmt.Errorf("bad response: %s", err)

		return nil, fmtErr
	}

	versionID, err := json.Marshal(result.VersionId)

	if err != nil {
		fmtErr := fmt.Errorf("Issue with json.Marshal %s", err)
		return nil, fmtErr
	}

	return versionID, nil

}

func init() {
	RootCmd.AddCommand(putCmd)

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	putCmd.Flags().StringP("kms-key-id", "k", "", "KMS CMK ID to use for encryption")
	putCmd.Flags().BoolP("verbose-output", "V", false, "Set to output the version id of the uploaded object")
	putCmd.MarkFlagRequired("kms-key-id")
	viper.BindPFlag("kmskeyid", putCmd.Flags().Lookup("kms-key-id"))
	viper.BindPFlag("verboseoutput", putCmd.Flags().Lookup("verbose-output"))

}
