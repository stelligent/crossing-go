package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

		//Open file
		file, err := os.Open(sourceFile)

		if err != nil {
			fmt.Fprintf(os.Stderr, "err opening file: %s", err)
		}

		defer file.Close()

		fileInfo, _ := file.Stat()
		size := fileInfo.Size()

		putobj := &PutObject{
			Bucket:   s3bucket,
			Key:      s3object,
			Source:   sourceFile,
			Reader:   bufio.NewReader(file),
			ByteSize: int(size),
		}

		cmkID := viper.GetString("kmskeyid")
		encryptionclient := NewEncryptionClient(cmkID)

		versionID, err := PutS3Cse(putobj, encryptionclient)
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

// PutObject Represents params for object input
type PutObject struct {
	Bucket   string
	Key      string
	Source   string
	Reader   *bufio.Reader
	ByteSize int
}

//PutS3Cse puts encrypted objects into S3
func PutS3Cse(p *PutObject, encryptionclient S3ClientPutAPI) ([]byte, error) {

	content, _ := p.Reader.Peek(p.ByteSize)
	fileType := http.DetectContentType(content)
	result, err := encryptionclient.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(p.Bucket),
		Key:         aws.String(p.Key),
		Body:        aws.ReadSeekCloser(strings.NewReader(string(content))),
		ContentType: aws.String(fileType),
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
