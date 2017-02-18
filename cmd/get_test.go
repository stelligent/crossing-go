package cmd

import "testing"

func Test_getS3Cse(t *testing.T) {
	type args struct {
		s3bucket string
		s3object string
		filedest string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Fail on garbage input", args{"nosuchbucket", "nosuchobject", "notafile"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := getS3Cse(tt.args.s3bucket, tt.args.s3object, tt.args.filedest); (err != nil) != tt.wantErr {
				t.Errorf("getS3Cse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
