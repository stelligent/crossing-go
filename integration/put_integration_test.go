package integration

import (
	"bufio"
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
	"github.com/aws/aws-sdk-go/service/s3/s3crypto"
	crosscrypto "github.com/stelligent/crossing-go/crypto"
	"github.com/stelligent/crossing-go/cmd/Put"

)

var (
	prefix     = "crossinggo"
	bucketName = ""
	key        = ""
	source     = ""
	kmsKey     = ""
)

//TestPutIntegration will test putS3Cse for a return value of a valid UTF-8 encoded version id
func TestPutIntegration(t *testing.T) {
	//Open file
	file, err := os.Open(source)

	if err != nil {
		fmt.Fprintf(os.Stderr, "err opening file: %s", err)
	}

	defer file.Close()

	fileInfo, _ := file.Stat()
	size := fileInfo.Size()

	//Return an encryption client from the global session
	cmkID := kmsKey
	newSess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)
	//Create the KeyProvider
	handler := s3crypto.NewKMSKeyGenerator(kms.New(newSess), cmkID)
	cipher := s3crypto.AESCBCContentCipherBuilder(handler, crosscrypto.NewPKCS7Padder(16))
	svc := s3crypto.NewEncryptionClient(newSess, cipher)
	encryptionclient := &cmd.Put{
		Client:   svc.S3Client,
		Bucket:   bucketName,
		Key:      key,
		Source:   source,
		Reader:   bufio.NewReader(file),
		ByteSize: int(size),
	}

	args := []struct {
		bucket     string
		file       string
		kmskeyid   string
		filesource string
		isEncoded  bool
	}{
		{bucketName, key, kmsKey, source, true},
	}

	for _, arg := range args {
		vstring, err := putS3Cse()
		isvalid := utf8.Valid(vstring)
		if err != nil {
			t.Errorf("Error occured: %v", err)
		} else if len(vstring) < 0 {
			t.Errorf("No version string was returned")
		} else if isvalid == true {
			t.Logf("Success! %v", vstring)
		}
	}
}

//SetupPut sets up an S3 Bucket, KMS key, and writes a file for Put integration testing
func SetupPut() {

	rando := strings.ToLower(randStringBytesMaskImprSrcUnsafe(9))
	buffer := bytes.NewBufferString(prefix)
	buffer.WriteString(rando)
	bucketName = buffer.String()
	key = "dat1"
	source = "/tmp/dat1"
	createWriteFile()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)

	setUpBucket(sess, bucketName)
	kmsKey = setupKmsKey(sess)

	if err != nil {
		exitErrorf("Unable to create session", err)
	}

}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
}

func setUpBucket(sess *session.Session, bucketName string) {
	svc := s3.New(sess)

	_, err := svc.CreateBucket(&s3.CreateBucketInput{
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

//Write a file to disk
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

//Generate random postfix
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func randStringBytesMaskImprSrcUnsafe(n int) string {
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

//Setup KMS key
func setupKmsKey(sess *session.Session) string {
	rando := strings.ToLower(randStringBytesMaskImprSrcUnsafe(9))
	buffer := bytes.NewBufferString(prefix)
	buffer.WriteString(rando)
	aliasname := buffer.String()

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

	if reqerr != nil {
		exitErrorf("Empty!", reqerr)
	} else {
		fmt.Printf("Returning key: %q", returnkey)
	}
	newalias := "alias/" + aliasname
	aliasreq, aliasresp := svc.CreateAliasRequest(&kms.CreateAliasInput{
		AliasName:   aws.String(newalias),
		TargetKeyId: aws.String(string(returnkey)),
	})

	aliaserr := aliasreq.Send()
	if aliaserr != nil {
		exitErrorf("Error occured creating alias!", aliaserr)
	} else {
		fmt.Println(aliasresp)
	}
	return newalias
}

//emptyBucket empties the Amazon S3 bucket
func emptyBucket() {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)

	svc := s3.New(sess)

	objectversions, err := svc.ListObjectVersions(&s3.ListObjectVersionsInput{
		Bucket:    aws.String(bucketName),
		KeyMarker: aws.String(key),
	})

	if err != nil {
		exitErrorf("Listing error occurred: ", err)
	}

	versions := objectversions.Versions

	for _, version := range versions {
		req, resp := svc.DeleteObjectRequest(&s3.DeleteObjectInput{
			Bucket:    aws.String(bucketName),
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

func deleteBucket() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)

	if err != nil {
		exitErrorf("Unable to create session", err)
	}

	s3svc := s3.New(sess)

	// Delete test bucket
	s3buckreq, s3buckresp := s3svc.DeleteBucketRequest(&s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})

	s3buckerr := s3buckreq.Send()

	if s3buckerr != nil {
		fmt.Println("Error occurred deleting bucket: ", s3buckerr)
		emptyBucket()
	} else {
		fmt.Println("Delete was successful", s3buckresp)
	}
}

//PutCleanUp destroys all resources created for integration testing
func PutCleanUp() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2")},
	)
	// Empty bucket
	emptyBucket()

	// Delete bucket
	deleteBucket()

	// Delete kms key
	kmssvc := kms.New(sess)

	keyoutput, err := kmssvc.DescribeKey(&kms.DescribeKeyInput{
		KeyId: aws.String(kmsKey),
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
