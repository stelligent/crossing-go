package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get S3URL [destination]",
	Short: "Retrieve an object from S3",
	Long: `Downloads an S3 object, using Client Side Encryption (CSE)
to decrypt it securely.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
			return err
		}
		if err := cobra.MaximumNArgs(2)(cmd, args); err != nil {
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

		filedest := ""
		// If destination file not explicitly given, determine from
		// last part of S3 object key
		if len(args) == 2 {
			filedest = args[1]
		} else {
			filedest = "."
		}

		dest, err := os.Stat(filedest)

		if dest.IsDir() {
			objectComponents := strings.Split(s3object, "/")
			filedest = filedest + "/" + objectComponents[len(objectComponents)-1]
		}

		getobj := &GetObject{
			Bucket:          s3bucket,
			Key:             s3object,
			FileDestination: filedest,
			Version:         viper.GetString("versionid"),
		}

		decryptionclient := NewDecryptionClient()

		content, geterr := GetS3Cse(getobj, decryptionclient)
		// Pretty-print the response data.
		if geterr != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		f, err := os.Create(filedest)
		defer f.Close()
		if err != nil {
			fmt.Println(err)
		}
		io.Copy(f, content)
		content.Close()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

// GetObject represents params for get object input
type GetObject struct {
	Bucket          string
	Key             string
	FileDestination string
	Version         string
}

//GetS3Cse gets and decrypts objects from S3
func GetS3Cse(g *GetObject, decryptionclient S3DecryptionClientAPI) (io.ReadCloser, error) {
	// fmt.Println("getS3 bucket:" + s3bucket + " object:" + s3object + " dest:" + filedest)
	// cmkID := "_unused_get_kms_key_"t

	params := &s3.GetObjectInput{
		Bucket: aws.String(g.Bucket),
		Key:    aws.String(g.Key),
	}

	if g.Version != "" {
		params.VersionId = aws.String(g.Version)
	}

	result, err := decryptionclient.GetObject(params)

	if err != nil {
		fmt.Println("Error in fetch!")
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err)
		return nil, err
	}

	return result.Body, nil
}

func init() {
	RootCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	getCmd.Flags().StringP("version-id", "v", "", "Version ID of the object to download")
	viper.BindPFlag("versionid", getCmd.Flags().Lookup("version-id"))

}
