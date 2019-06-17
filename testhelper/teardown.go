package testhelper

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
)

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

//CleanUp cleans all provisioned resources from unit testing
func CleanUp(bucketname string, kmskey string, key string) {
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
