package cmd

import (
	"fmt"
	"net/url"
	"strings"
)

func parseS3Url(s3urlstring string) (bucket string, key string, err error) {
	s3url, err := url.Parse(s3urlstring)
	if err != nil {
		return "", "", err
	}

	if !strings.HasPrefix(s3url.String(), "s3://") || strings.HasPrefix(s3url.String(), "s3:///") {
		return "", "", fmt.Errorf("invalid schema: %s not s3", s3url.Scheme)
	}
	bucket = s3url.Host

	if s3url.Path[0] == '/' {
		key = s3url.Path[1:]
	} else {
		key = s3url.Path
	}

	return
}
