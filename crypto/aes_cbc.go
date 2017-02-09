package crosscrypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/service/s3/s3crypto"
)

// Appends padding.
func pkcs7Pad(data []byte, blocklen int) ([]byte, error) {
	if blocklen <= 0 {
		return nil, fmt.Errorf("invalid blocklen %d", blocklen)
	}
	padlen := 1
	for ((len(data) + padlen) % blocklen) != 0 {
		padlen = padlen + 1
	}

	pad := bytes.Repeat([]byte{byte(padlen)}, padlen)
	return append(data, pad...), nil
}

// Returns slice of the original data without padding.
func pkcs7Unpad(data []byte, blocklen int) ([]byte, error) {
	if blocklen <= 0 {
		return nil, fmt.Errorf("invalid blocklen %d", blocklen)
	}
	if len(data)%blocklen != 0 || len(data) == 0 {
		return nil, fmt.Errorf("invalid data len %d", len(data))
	}
	padlen := int(data[len(data)-1])
	if padlen > blocklen || padlen == 0 {
		return nil, fmt.Errorf("invalid padding")
	}
	// check padding
	pad := data[len(data)-padlen:]
	for i := 0; i < padlen; i++ {
		if pad[i] != byte(padlen) {
			return nil, fmt.Errorf("invalid padding")
		}
	}

	return data[:len(data)-padlen], nil
}

// AESCBC Symmetric encryption algorithm. Since Golang designed this
// with only TLS in mind. We have to load it all into memory meaning
// this isn't streamed.
type aesCBC struct {
	decrypter cipher.BlockMode
	encrypter cipher.BlockMode
	iv        []byte
}

// NewAESCBC creates a new AES CBC cipher. Expects keys to be of
// the correct size.
//
// Example:
//
//	cd := &s3crypto.CipherData{
//		Key: key,
//		"IV": iv,
//	}
//	cipher, err := crosscrypto.NewAESCBC(cd)
func newAESCBC(cd s3crypto.CipherData) (s3crypto.Cipher, error) {
	block, err := aes.NewCipher(cd.Key)
	if err != nil {
		return nil, err
	}

	aescbcDecrypter := cipher.NewCBCDecrypter(block, cd.IV)
	if err != nil {
		return nil, err
	}

	aescbcEncrypter := cipher.NewCBCEncrypter(block, cd.IV)
	if err != nil {
		return nil, err
	}

	return &aesCBC{aescbcDecrypter, aescbcEncrypter, cd.IV}, nil
}

// Encrypt will encrypt the data using AES CBC
// Tag will be included as the last 16 bytes of the slice
func (c *aesCBC) Encrypt(src io.Reader) io.Reader {
	reader := &cbcEncryptReader{
		encrypter: c.encrypter,
		iv:        c.iv,
		src:       src,
	}
	return reader
}

type cbcEncryptReader struct {
	encrypter cipher.BlockMode
	iv        []byte
	src       io.Reader
	buf       *bytes.Buffer
}

func (reader *cbcEncryptReader) Read(data []byte) (int, error) {
	if reader.buf == nil {
		b, err := ioutil.ReadAll(reader.src)
		if err != nil {
			return len(b), err
		}
		b, _ = pkcs7Pad(b, reader.encrypter.BlockSize())
		reader.encrypter.CryptBlocks(b, b)
		reader.buf = bytes.NewBuffer(b)
	}

	return reader.buf.Read(data)
}

// Decrypt will decrypt the data using AES GCM
func (c *aesCBC) Decrypt(src io.Reader) io.Reader {
	return &cbcDecryptReader{
		decrypter: c.decrypter,
		iv:        c.iv,
		src:       src,
	}
}

type cbcDecryptReader struct {
	decrypter cipher.BlockMode
	iv        []byte
	src       io.Reader
	buf       *bytes.Buffer
}

func (reader *cbcDecryptReader) Read(data []byte) (int, error) {
	if reader.buf == nil {
		b, err := ioutil.ReadAll(reader.src)
		if err != nil {
			return len(b), err
		}
		reader.decrypter.CryptBlocks(b, b)
		if err != nil {
			return len(b), err
		}
		b, _ = pkcs7Unpad(b, reader.decrypter.BlockSize())

		reader.buf = bytes.NewBuffer(b)
	}

	return reader.buf.Read(data)
}
