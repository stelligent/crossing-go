package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3crypto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/stelligent/crossing-go/crypto"
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
		err = putS3Cse(s3bucket, s3object, viper.GetString("kmskeyid"), sourceFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "err uploading file: %s\n", err)
			os.Exit(1)
		}
	},
}

func putS3Cse(bucket string, key string, kmskeyid string, source string) error {
	file, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("err opening file: %s", err)
	}
	defer file.Close()
	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)
	fileBytes := bytes.NewReader(buffer)
	fileType := http.DetectContentType(buffer)

	params := &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   fileBytes,
		// ContentLength: aws.Int64(size),
		ContentType: aws.String(fileType),
	}
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	cmkID := kmskeyid
	// Create the KeyProvider
	handler := s3crypto.NewKMSKeyGenerator(kms.New(sess), cmkID)

	// Create an encryption and decryption client
	svc := s3crypto.NewEncryptionClient(sess, s3crypto.AESCBCContentCipherBuilder(handler, crosscrypto.NewPKCS7Padder(16)))

	_, err = svc.PutObject(params)
	if err != nil {
		return fmt.Errorf("bad response: %s", err)
	}
	return nil
}

func init() {
	RootCmd.AddCommand(putCmd)

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	putCmd.Flags().StringP("kms-key-id", "k", "", "KMS CMK ID to use for encryption")
	viper.BindPFlag("kmskeyid", putCmd.Flags().Lookup("kms-key-id"))

}
