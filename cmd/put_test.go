package cmd

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

type mockedPutObjectOutput struct {
	s3iface.S3API
	Output s3.PutObjectOutput
}

func (m mockedPutObjectOutput) PutObject(in *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	return &m.Output, nil
}
func TestPut_putS3Cse(t *testing.T) {

	cases := []struct {
		Output   s3.PutObjectOutput
		Expected []byte
	}{
		{ // Case 1, expect output with versionId
			Output: s3.PutObjectOutput{
				VersionId: aws.String(".FLQEZscLIcfxSq.jsFJ.szUkmng2Yw6"),
			},
			Expected: []byte("\".FLQEZscLIcfxSq.jsFJ.szUkmng2Yw6\""),
		},
		{ // Case 2, no versionId returned
			Output: s3.PutObjectOutput{
				VersionId: new(string),
			},
			Expected: []byte("\"\""),
		},
	}

	for i, tt := range cases {
		p := Put{
			Client: mockedPutObjectOutput{Output: tt.Output},
			Bucket: fmt.Sprintf("mockBUCKET_%d", i),
			Key:    fmt.Sprintf("mockKEY_%d", i),
			Source: fmt.Sprintf("mockSOURCE_%d", i),
		}
		versionID, err := p.putS3Cse()
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
