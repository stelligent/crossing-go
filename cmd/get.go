package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/stelligent/crossing-go/clientfactory"

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

		sess := viper.Get("ClientSess").(*session.Session)

		decyrptionclient := Get{
			Client:          clientfactory.NewDecryptionClient(sess).S3Client,
			Bucket:          s3bucket,
			Key:             s3object,
			FileDestination: filedest,
		}

		err = decyrptionclient.getS3Cse()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

// Get provides the ability to get objects
type Get struct {
	Client          s3iface.S3API
	Bucket          string
	Key             string
	FileDestination string
}

func (g *Get) getS3Cse() error {
	// fmt.Println("getS3 bucket:" + s3bucket + " object:" + s3object + " dest:" + filedest)
	// cmkID := "_unused_get_kms_key_"

	result, err := g.Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(g.Bucket),
		Key:    aws.String(g.Key),
	})

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
	f, err := os.Create(g.FileDestination)
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}
	io.Copy(f, result.Body)
	result.Body.Close()
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
