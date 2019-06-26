package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

type mockGetObjectOutput struct {
	s3iface.S3API
	Output s3.GetObjectOutput
}

func (m mockGetObjectOutput) GetObject(in *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	return &m.Output, nil
}

func TestGet_getS3Cse(t *testing.T) {

	cases := []struct {
		Output   s3.GetObjectOutput
		Expected []byte
	}{
		{ // Case 1, expect successful output
			Output: s3.GetObjectOutput{
				Body: ioutil.NopCloser(bytes.NewBuffer([]byte("Hello"))),
			},
			Expected: []byte("Hello"),
		},
	}

	for i, tt := range cases {
		g := Get{
			Client:          mockGetObjectOutput{Output: tt.Output},
			Bucket:          fmt.Sprintf("mockBUCKET_%d", i),
			Key:             fmt.Sprintf("mockKEY_%d", i),
			Version:         fmt.Sprintf("mockVERSION_%d", i),
			FileDestination: fmt.Sprintf("mockDESTINATION_%d", i),
		}
		content, err := g.getS3Cse()
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
