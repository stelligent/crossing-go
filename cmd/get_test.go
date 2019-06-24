package cmd

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

func TestGet_getS3Cse(t *testing.T) {
	type fields struct {
		Client          s3iface.S3API
		Bucket          string
		Key             string
		FileDestination string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Get{
				Client:          tt.fields.Client,
				Bucket:          tt.fields.Bucket,
				Key:             tt.fields.Key,
				FileDestination: tt.fields.FileDestination,
			}
			if err := g.getS3Cse(); (err != nil) != tt.wantErr {
				t.Errorf("Get.getS3Cse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
