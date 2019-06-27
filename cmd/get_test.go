package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
)

func TestGet_getS3Cse(t *testing.T) {

	cases := []struct {
		S3Decrypt *MockS3DecryptionClientAPI
		Expected  []byte
	}{
		{ // Case 1, expect successful write
			S3Decrypt: &MockS3DecryptionClientAPI{
				GetObjectOutput: &s3.GetObjectOutput{
					Body: ioutil.NopCloser(bytes.NewBuffer([]byte("Hello"))),
				},
			},
			Expected: []byte("Hello"),
		},
	}

	for i, tt := range cases {
		g := &GetObject{
			Bucket:          fmt.Sprintf("mockBUCKET_%d", i),
			Key:             fmt.Sprintf("mockKEY_%d", i),
			FileDestination: fmt.Sprintf("mockDESTINATION_%d", i),
		}
		content, err := GetS3Cse(g, tt.S3Decrypt)
		if err != nil {
			t.Fatalf("Unexpected error, %v", err)
		}
		bufOne := new(bytes.Buffer)
		bufOne.ReadFrom(content)
		actual := bufOne.String()
		expected := string(tt.Expected)
		if actual != expected {
			t.Log("Actual: ", actual)
			t.Log("Expected: ", expected)
			t.Fatalf("Expected: %v, Actual: %v", actual, expected)
		}
	}
}
