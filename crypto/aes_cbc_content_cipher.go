package crosscrypto

import (
	"io"

	"github.com/aws/aws-sdk-go/service/s3/s3crypto"
)

const (
	cbcKeySize   = 32
	cbcNonceSize = 16
)

type cbcContentCipherBuilder struct {
	generator s3crypto.CipherDataGenerator
}

// AESCBCContentCipherBuilder returns a new encryption only mode structure with a specific cipher
// for the master key
func AESCBCContentCipherBuilder(generator s3crypto.CipherDataGenerator) s3crypto.ContentCipherBuilder {
	return cbcContentCipherBuilder{generator}
}

func (builder cbcContentCipherBuilder) ContentCipher() (s3crypto.ContentCipher, error) {
	cd, err := builder.generator.GenerateCipherData(cbcKeySize, cbcNonceSize)
	if err != nil {
		return nil, err
	}

	return NewAESCBCContentCipher(cd)
}

// NewAESCBCContentCipher is AESCBCPKCS5Padding provider
func NewAESCBCContentCipher(cd s3crypto.CipherData) (s3crypto.ContentCipher, error) {
	cd.CEKAlgorithm = AESCBCPKCS5Padding
	cd.TagLength = "128"

	cipher, err := newAESCBC(cd)
	if err != nil {
		return nil, err
	}

	return &aesCBCContentCipher{
		CipherData: cd,
		Cipher:     cipher,
	}, nil
}

// AESCBCContentCipher will use AES CBC for the main cipher.
type aesCBCContentCipher struct {
	CipherData s3crypto.CipherData
	Cipher     s3crypto.Cipher
}

// EncryptContents will generate a random key and iv and encrypt the data using cbc
func (cc *aesCBCContentCipher) EncryptContents(src io.Reader) (io.Reader, error) {
	return cc.Cipher.Encrypt(src), nil
}

// DecryptContents will use the symmetric key provider to instantiate a new CBC cipher.
// We grab a decrypt reader from cbc and wrap it in a CryptoReadCloser. The only error
// expected here is when the key or iv is of invalid length.
func (cc *aesCBCContentCipher) DecryptContents(src io.ReadCloser) (io.ReadCloser, error) {
	reader := cc.Cipher.Decrypt(src)
	return &s3crypto.CryptoReadCloser{Body: src, Decrypter: reader}, nil
}

// GetCipherData returns cipher data
func (cc aesCBCContentCipher) GetCipherData() s3crypto.CipherData {
	return cc.CipherData
}
