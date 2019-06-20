package clientfactory

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3crypto"
)

//NewEncryptionClient returns an s3crypto encryption client
func NewEncryptionClient(sess *session.Session, cipher s3crypto.ContentCipherBuilder) *s3crypto.EncryptionClient {
	svc := s3crypto.NewEncryptionClient(sess, cipher)

	return svc
}
