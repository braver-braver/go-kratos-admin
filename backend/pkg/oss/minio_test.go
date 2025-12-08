package oss

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/tx7do/go-utils/trans"

	v1 "kratos-admin/api/gen/go/file/service/v1"

	conf "github.com/tx7do/kratos-bootstrap/api/gen/go/conf/v1"
)

func createTestClient() *MinIOClient {
	return NewMinIoClient(&conf.Bootstrap{
		Oss: &conf.OSS{
			Minio: &conf.OSS_MinIO{
				Endpoint:     "127.0.0.1:9000",
				UploadHost:   "127.0.0.1:9000",
				DownloadHost: "127.0.0.1:9000",
				AccessKey:    "root",
				SecretKey:    "*Abcd123456",
			},
		},
	}, log.DefaultLogger)
}

func TestMinIoClient(t *testing.T) {
	if conn, err := net.DialTimeout("tcp", "127.0.0.1:9000", time.Second); err != nil {
		t.Skip("minio not reachable, skipping integration test")
	} else {
		_ = conn.Close()
	}

	cli := createTestClient()
	assert.NotNil(t, cli)

	resp, err := cli.OssUploadUrl(context.Background(), &v1.OssUploadUrlRequest{
		Method:      v1.OssUploadUrlRequest_Put,
		ContentType: trans.String("image/jpeg"),
		BucketName:  trans.String("images"),
		FilePath:    trans.String("20221010"),
	})
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestListFile(t *testing.T) {
	if conn, err := net.DialTimeout("tcp", "127.0.0.1:9000", time.Second); err != nil {
		t.Skip("minio not reachable, skipping integration test")
	} else {
		_ = conn.Close()
	}

	cli := createTestClient()
	assert.NotNil(t, cli)

	req := &v1.ListOssFileRequest{
		BucketName: trans.Ptr("users"),
		Folder:     trans.Ptr("1"),
		Recursive:  trans.Ptr(true),
	}
	files, err := cli.ListFile(context.Background(), req)
	assert.Nil(t, err)
	fmt.Println(files)
}
