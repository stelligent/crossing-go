package cmd

import "testing"

func Test_parseS3Url(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name       string
		args       args
		wantBucket string
		wantKey    string
		wantErr    bool
	}{
		{"Fail non-S3 URL", args{"notaurl"}, "", "", true},
		{"Fail malformed S3 URL", args{"s3:/foo/bar"}, "", "", true},
		{"Succeed well-formed S3 URL", args{"s3://foo/bar"}, "foo", "bar", false},
		{"Succeed S3 URL with key prefix", args{"s3://foo/bar/baz"}, "foo", "bar/baz", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBucket, gotKey, err := parseS3Url(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseS3Url() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotBucket != tt.wantBucket {
				t.Errorf("parseS3Url() gotBucket = %v, want %v", gotBucket, tt.wantBucket)
			}
			if gotKey != tt.wantKey {
				t.Errorf("parseS3Url() gotKey = %v, want %v", gotKey, tt.wantKey)
			}
		})
	}
}
