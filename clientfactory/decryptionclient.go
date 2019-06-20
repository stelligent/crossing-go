package clientfactory

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3crypto"

	crosscrypto "github.com/stelligent/crossing-go/crypto"
)

//NewDecryptionClient returns an s3crypto decryption client
func NewDecryptionClient(sess *session.Session) *s3crypto.DecryptionClient {
	svc := s3crypto.NewDecryptionClient(sess)

	svc.CEKRegistry[crosscrypto.AESCBCPKCS5Padding] = crosscrypto.NewAESCBCContentCipher

	return svc
}
