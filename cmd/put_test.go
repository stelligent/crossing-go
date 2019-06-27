package cmd

import (
	"bufio"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

//TestPutS3Cse will test putS3Cse
func TestPutS3Cse(t *testing.T) {

	reader := strings.NewReader("Hello")
	cases := []struct {
		S3Encrypt *MockS3ClientPutAPI
		Expected  []byte
	}{
		{ // Case 1, expect output with versionId
			S3Encrypt: &MockS3ClientPutAPI{
				PutObjectOutput: &s3.PutObjectOutput{
					VersionId: aws.String(".FLQEZscLIcfxSq.jsFJ.szUkmng2Yw6"),
				},
			},
			Expected: []byte("\".FLQEZscLIcfxSq.jsFJ.szUkmng2Yw6\""),
		},
		{ // Case 2, no versionId returned
			S3Encrypt: &MockS3ClientPutAPI{
				PutObjectOutput: &s3.PutObjectOutput{
					VersionId: new(string),
				},
			},
			Expected: []byte("\"\""),
		},
	}

	for i, tt := range cases {
		p := &PutObject{
			Bucket:   fmt.Sprintf("mockBUCKET_%d", i),
			Key:      fmt.Sprintf("mockKEY_%d", i),
			Source:   fmt.Sprintf("mockSOURCE_%d", i),
			Reader:   bufio.NewReader(reader),
			ByteSize: 5,
		}
		versionID, err := PutS3Cse(p, tt.S3Encrypt)
		if err != nil {
			t.Fatalf("Unexpected error, %v", err)
		}
		if a, e := len(versionID), len(tt.Expected); a != e {
			t.Log("VersionId: ", string(versionID))
			t.Log("Expected: ", string(tt.Expected))
			t.Fatalf("Expected %d, length %d", a, e)

		}
	}
}
