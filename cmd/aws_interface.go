package cmd

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3crypto"
	crosscrypto "github.com/stelligent/crossing-go/crypto"
)

// S3EncryptionClient creates a new client for encryption
var (
	// Initialize global session
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
)

//S3ClientPutAPI puts encrypted objects in S3
type S3ClientPutAPI interface {
	PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error)
}

//S3DecryptionClientAPI gets and decrypts objects in S3
type S3DecryptionClientAPI interface {
	GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error)
}

//NewEncryptionClient returns a new encryption client
func NewEncryptionClient(cmkID string) S3ClientPutAPI {
	handler := s3crypto.NewKMSKeyGenerator(kms.New(sess), cmkID)
	cipher := s3crypto.AESCBCContentCipherBuilder(handler, crosscrypto.NewPKCS7Padder(16))
	svc := s3crypto.NewEncryptionClient(sess, cipher)
	return svc
}

//NewDecryptionClient returns a new decryption client
func NewDecryptionClient() S3DecryptionClientAPI {
	svc := s3crypto.NewDecryptionClient(sess)
	svc.CEKRegistry[crosscrypto.AESCBCPKCS5Padding] = crosscrypto.NewAESCBCContentCipher
	return svc
}
