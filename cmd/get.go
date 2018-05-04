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
			objectComponents := strings.Split(args[0], "/")
			filedest = objectComponents[len(objectComponents)-1]
		}
		err = getS3Cse(s3bucket, s3object, filedest)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func getS3Cse(s3bucket, s3object, filedest string) error {
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
		fmt.Println(err)
		return err
	}

	// Pretty-print the response data.
	// fmt.Println(resp)
	// n, err :=
	f, err := os.Create(filedest)
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}
	io.Copy(f, resp.Body)
	resp.Body.Close()
	return nil
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
