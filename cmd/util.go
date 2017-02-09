package cmd

import (
	"errors"
	"strings"
)

func parseS3Url(url string) (bucket string, key string, err error) {
	if !strings.HasPrefix(url, "s3://") {
		return "", "", errors.New("malformed S3 URL")
	}
	// Parse S3 URL for bucket and object key
	s3url := strings.SplitN(url, "/", 4)
	bucket = s3url[2]
	key = s3url[3]
	err = nil
	return
}
