package cmd

import (
	"fmt"
	"net/url"
)

func parseS3Url(s3urlstring string) (bucket string, key string, err error) {
	s3url, err := url.Parse(s3urlstring)
	if err != nil {
		return "", "", err
	}
	if s3url.Scheme != "s3" {
		return "", "", fmt.Errorf("invalid schema: %s not s3", s3url.Scheme)
	}
	bucket = s3url.Host
	key = s3url.Path
	err = nil
	fmt.Printf("Host = %s, Path = %s", bucket, key)

	return
}
