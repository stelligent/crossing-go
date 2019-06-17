package testhelper

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
	"unsafe"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
)

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

//SetUpBucket sets up S3 bucket for unit testing
func SetUpBucket(bucketName string) {

	createWriteFile()
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)

	if err != nil {
		exitErrorf("Unable to create session", err)
	}

	svc := s3.New(sess)

	_, err = svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})

	if err != nil {
		exitErrorf("Unable to create bucket %q, %v", bucketName, err)
	}

	fmt.Printf("Waiting for bucket %q to be created...\n", bucketName)

	err = svc.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})

	if err != nil {
		exitErrorf("Error occurred while waiting for bucket to be created, %v", bucketName)
	}

	fmt.Printf("Bucket %q successfully created\n", bucketName)

	// Turn on versioning on the bucket
	input := &s3.PutBucketVersioningInput{
		Bucket: aws.String(bucketName),
		VersioningConfiguration: &s3.VersioningConfiguration{
			MFADelete: aws.String("Disabled"),
			Status:    aws.String("Enabled"),
		},
	}

	result, err := svc.PutBucketVersioning(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				exitErrorf(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			exitErrorf("Setting versioning failed: %v", err.Error())
		}
		exitErrorf("Catastrophe! %v", err)
	}

	fmt.Printf("Successfully configured versioning %q", result)

}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func createWriteFile() {
	f, err := os.Create("/tmp/dat1")
	check(err)

	defer f.Close()

	d1 := []byte{115, 111, 109, 101, 10}
	n2, err := f.Write(d1)
	check(err)
	fmt.Printf("wrote %d bytes\n", n2)

	f.Sync()
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

//RandStringBytesMaskImprSrcUnsafe returns random string
func RandStringBytesMaskImprSrcUnsafe(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

//SetupKmsKey sets up kms key for unit testing
func SetupKmsKey() string {
	prefix := "alias/crossingunittest"
	rando := strings.ToLower(RandStringBytesMaskImprSrcUnsafe(9))
	buffer := bytes.NewBufferString(prefix)
	buffer.WriteString(rando)
	aliasname := buffer.String()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)

	if err != nil {
		exitErrorf("Unable to create session", err)
	}

	svc := kms.New(sess)

	req, resp := svc.CreateKeyRequest(&kms.CreateKeyInput{
		Tags: []*kms.Tag{
			{
				TagKey:   aws.String("crossinggokey"),
				TagValue: aws.String("crossinggounittest"),
			},
		},
	})

	reqerr := req.Send()
	if reqerr == nil {
		fmt.Println(resp)
	} else {
		exitErrorf("Unable to create session", reqerr)
	}
	returnkey := *resp.KeyMetadata.KeyId

	if err != nil {
		exitErrorf("Empty!", err)
	} else {
		fmt.Printf("Returning key: %q", returnkey)
	}

	aliasreq, aliasresp := svc.CreateAliasRequest(&kms.CreateAliasInput{
		AliasName:   aws.String(aliasname),
		TargetKeyId: aws.String(string(returnkey)),
	})

	aliaserr := aliasreq.Send()
	if aliaserr != nil {
		exitErrorf("Error occured creating alias!", err)
	} else {
		fmt.Println(aliasresp)
	}
	return aliasname
}
