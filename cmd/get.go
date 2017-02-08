package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3crypto"

	"github.com/spf13/cobra"

	"github.com/stelligent/crossing-go/crypto"

	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [S3 URL] [destination]",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("get called")
		if len(args) < 1 || len(args) > 2 {
			cmd.UsageFunc()(cmd)
			return
		}
		if !strings.HasPrefix(args[0], "s3://") {
			cmd.UsageFunc()(cmd)
			return
		}
		// Parse S3 URL for bucket and object key
		s3url := strings.SplitN(args[0], "/", 4)
		s3bucket := s3url[2]
		s3object := s3url[3]
		filedest := ""
		// If destination file not explicitly given, determine from
		// last part of S3 object key
		if len(args) == 2 {
			filedest = args[1]
		} else {
			objectComponents := strings.Split(args[0], "/")
			filedest = objectComponents[len(objectComponents)-1]
		}
		getS3Cse(s3bucket, s3object, filedest)
	},
}

func getS3Cse(s3bucket, s3object, filedest string) {
	// fmt.Println("getS3 bucket:" + s3bucket + " object:" + s3object + " dest:" + filedest)
	// cmkID := "_unused_get_kms_key_"
	params := &s3.GetObjectInput{
		Bucket: &s3bucket,
		Key:    &s3object}
	sess := session.New()
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
		fmt.Println(err.Error())
		return
	}

	// Pretty-print the response data.
	// fmt.Println(resp)
	// n, err :=
	f, err := os.Create(filedest)
	defer f.Close()
	io.Copy(f, resp.Body)
	resp.Body.Close()
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

}
