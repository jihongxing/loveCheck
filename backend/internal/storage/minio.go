package storage

import (
	"context"
	"io"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"lovecheck/pkg/logger"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// MinioClient holds the global minio instance.
// MinioClient 暴露可以在任意包调用的持久化 MinIO 全局句柄。
var MinioClient *minio.Client

// BucketName where we securely store non-PII evidence records.
// BucketName 是用来存放未压缩全量证据的安全桶名称。
const BucketName = "lovecheck-evidence"

// InitMinio initializes connection to MinIO and ensures the target bucket is created.
// InitMinio 执行连接认证，并能在目标对象存储库未被创建时自动为您新建它。
func InitMinio() {
	endpoint := getEnv("MINIO_ENDPOINT", "localhost:9000")
	accessKeyID := getEnv("MINIO_ACCESS_KEY", "lovecheck_admin")
	secretAccessKey := getEnv("MINIO_SECRET_KEY", "lovecheck_minio_pwd")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	var err error
	MinioClient, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("MinIO initialization failed")
	}

	ctx := context.Background()

	// Attempt to create the bucket upon starting up.
	// 启动阶段安全地试图创建该桶池结构。
	err = MinioClient.MakeBucket(ctx, BucketName, minio.MakeBucketOptions{})
	if err != nil {
		// If exists, ignore the error and gracefully skip.
		// 若已存在，则静默其抛出的异常逻辑，放任应用流转即行。
		exists, errBucketExists := MinioClient.BucketExists(ctx, BucketName)
		if errBucketExists == nil && exists {
			logger.Log.Info().Str("bucket", BucketName).Msg("MinIO bucket verified")
		} else {
			logger.Log.Fatal().Err(err).Str("bucket", BucketName).Msg("Failed to create MinIO bucket")
		}
	} else {
		logger.Log.Info().Str("bucket", BucketName).Msg("MinIO bucket created")
	}
}

// UploadEvidence securely uploads a file stream into the bucket.
// UploadEvidence 将文件流以非公开形式安全固化进入 MinIO，返回它的资源索引。
func UploadEvidence(objectName string, reader io.Reader, objectSize int64, contentType string) (string, error) {
	ctx := context.Background()
	_, err := MinioClient.PutObject(ctx, BucketName, objectName, reader, objectSize, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", err
	}
	// Return the object name which acts as an evidence mask identifier.
	// 正常放回该挂载物的资源名指针（可落库，作为 evidence_mask_url 拼装源头）。
	return objectName, nil
}
