package cmd

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
	"unsafe"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
)

// TestPutS3Cse tests functionality of putS3Cse
func TestPutS3Cse(t *testing.T) {
	vString, err := putS3Cse("nosuchbucket", "nokey", "badkmsid", "invalidsource")
	assert.Error(t, err, "Calling with bad data returns error")
	t.Log(vString)
}

// Test that putS3Cse returns a valid UTF-8 encoded version id
func TestVersionIdOutput(t *testing.T) {
	// Run setup
	// Create new random bucket name for bucket setup
	prefix := "testbucket"
	rando := strings.ToLower(RandStringBytesMaskImprSrcUnsafe(9))
	buffer := bytes.NewBufferString(prefix)
	buffer.WriteString(rando)
	bucketName := buffer.String()
	filekey := "dat1"
	outputkmskeyid := setupKmsKey()
	sourcestring := "/tmp/dat1"

	createWriteFile()
	setUpBucket(bucketName)

	args := []struct {
		bucket    string
		key       string
		kmskeyid  string
		source    string
		isEncoded bool
	}{
		{bucketName, filekey, outputkmskeyid, sourcestring, true},
	}

	for _, arg := range args {
		vstring, err := putS3Cse(arg.bucket, arg.key, arg.kmskeyid, arg.source)
		isvalid := utf8.Valid(vstring)
		if err != nil {
			t.Errorf("Error occured: %v, wanted true", err)
			cleanUp(arg.bucket, arg.kmskeyid, arg.key)
		} else if len(vstring) < 0 {
			t.Errorf("No version string was returned")
			cleanUp(arg.bucket, arg.kmskeyid, arg.key)
		} else if isvalid == true {
			t.Logf("Success! %v", vstring)
		}
	}

	//run clean up
	cleanUp(bucketName, outputkmskeyid, sourcestring)

}
func Test_putS3Cse(t *testing.T) {
	type args struct {
		bucket   string
		key      string
		kmskeyid string
		source   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Bad source file",
			args: args{
				"nosuchbucket",
				"nosuchkey",
				"nosuchcmkid",
				"nosuchsource",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if vString, err := putS3Cse(tt.args.bucket, tt.args.key, tt.args.kmskeyid, tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("putS3Cse() error = %v, wantErr %v", err, tt.wantErr)
				t.Log(vString)
			}
		})
	}
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func setUpBucket(bucketName string) {
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

func setupKmsKey() string {
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

//emptyBucket empties the Amazon S3 bucket
func emptyBucket(bucket string, key string) {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)

	svc := s3.New(sess)

	objectversions, err := svc.ListObjectVersions(&s3.ListObjectVersionsInput{
		Bucket:    aws.String(bucket),
		KeyMarker: aws.String(key),
	})

	if err != nil {
		exitErrorf("Listing error occurred: ", err)
	}

	versions := objectversions.Versions

	for _, version := range versions {
		req, resp := svc.DeleteObjectRequest(&s3.DeleteObjectInput{
			Bucket:    aws.String(bucket),
			Key:       version.Key,
			VersionId: version.VersionId,
		})

		err := req.Send()
		if err != nil {
			exitErrorf("Issue deleting: ", err)
		} else {
			fmt.Println("Deleted: ", resp)
		}
	}
}

func deleteBucket(bucket string, key string) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)

	if err != nil {
		exitErrorf("Unable to create session", err)
	}

	s3svc := s3.New(sess)

	// Delete test bucket
	s3buckreq, s3buckresp := s3svc.DeleteBucketRequest(&s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	})

	s3buckerr := s3buckreq.Send()

	if s3buckerr != nil {
		fmt.Println("Error occurred deleting bucket: ", s3buckerr)
		emptyBucket(bucket, key)
	} else {
		fmt.Println("Delete was successful", s3buckresp)
	}
}

func cleanUp(bucketname string, kmskey string, key string) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)

	if err != nil {
		exitErrorf("Unable to create session", err)
	}

	// Empty bucket
	emptyBucket(bucketname, key)

	// Delete bucket
	deleteBucket(bucketname, key)

	// Delete kms key
	kmssvc := kms.New(sess)

	keyoutput, err := kmssvc.DescribeKey(&kms.DescribeKeyInput{
		KeyId: aws.String(kmskey),
	})

	if err != nil {
		exitErrorf("Describing key errored out: ", err)
	}
	keyid := *keyoutput.KeyMetadata.KeyId

	kmsreq, kmsresp := kmssvc.ScheduleKeyDeletionRequest(&kms.ScheduleKeyDeletionInput{
		KeyId:               aws.String(keyid),
		PendingWindowInDays: aws.Int64(7),
	})

	kmserr := kmsreq.Send()

	if kmserr != nil {
		exitErrorf("Deleting key error occurred: ", kmserr)
	} else {
		fmt.Println("Key deletion scheduled: ", kmsresp)
	}

	// Delete file created for test
	delerr := os.Remove("/tmp/dat1")

	if delerr != nil {
		exitErrorf("Deleting file error occurred: ", delerr)
	}
}
