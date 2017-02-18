package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPutS3Cse tests functionality of putS3Cse
func TestPutS3Cse(t *testing.T) {
	err := putS3Cse("nosuchbucket", "nokey", "badkmsid", "invalidsource")
	assert.Error(t, err, "Calling with bad data returns error")
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
			if err := putS3Cse(tt.args.bucket, tt.args.key, tt.args.kmskeyid, tt.args.source); (err != nil) != tt.wantErr {
				t.Errorf("putS3Cse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
