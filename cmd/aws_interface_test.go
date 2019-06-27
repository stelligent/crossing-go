package cmd

import "github.com/aws/aws-sdk-go/service/s3"

// MockS3ClientPutAPI represents S3EncryptionClient
type MockS3ClientPutAPI struct {
	PutObjectOutput *s3.PutObjectOutput
}

func (m *MockS3ClientPutAPI) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	return m.PutObjectOutput, nil
}

type MockS3DecryptionClientAPI struct {
	GetObjectOutput *s3.GetObjectOutput
}

func (m *MockS3DecryptionClientAPI) GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	return m.GetObjectOutput, nil
}
