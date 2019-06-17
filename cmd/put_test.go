package cmd

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stelligent/crossing-go/testhelper"
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
	rando := strings.ToLower(testhelper.RandStringBytesMaskImprSrcUnsafe(9))
	buffer := bytes.NewBufferString(prefix)
	buffer.WriteString(rando)
	bucketName := buffer.String()
	filekey := "dat1"
	outputkmskeyid := testhelper.SetupKmsKey()
	sourcestring := "/tmp/dat1"

	testhelper.SetUpBucket(bucketName)

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
			t.Errorf("Error occured: %v", err)
			testhelper.CleanUp(arg.bucket, arg.kmskeyid, arg.key)
		} else if len(vstring) < 0 {
			t.Errorf("No version string was returned")
			testhelper.CleanUp(arg.bucket, arg.kmskeyid, arg.key)
		} else if isvalid == true {
			t.Logf("Success! %v", vstring)
		}
	}

	//run clean up
	testhelper.CleanUp(bucketName, outputkmskeyid, sourcestring)

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
