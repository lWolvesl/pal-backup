package main

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
)

func main() {
	// MinIO 连接信息
	endpoint := "minio.example.com"
	accessKeyID := "your-access-key-id"
	secretAccessKey := "your-secret-access-key"
	bucketName := "your-bucket-name"

	// 待压缩的路径
	sourceDir := "/path/to/source"

	// 压缩后的文件名
	currentTime := time.Now()
	zipFileName := currentTime.Format("200601021504") + ".zip"

	// 创建一个缓冲区，用于保存压缩后的文件内容
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// 遍历源目录中的文件，并将它们添加到 zipWriter 中
	err := filepath.Walk(sourceDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 排除目录本身
		if filePath == sourceDir {
			return nil
		}

		// 创建一个新的文件头
		fileHeader, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		// 设置文件名为相对路径
		fileHeader.Name, err = filepath.Rel(sourceDir, filePath)
		if err != nil {
			return err
		}

		// 如果是目录，只创建目录头
		if info.IsDir() {
			fileHeader.Name += "/"
		} else {
			// 将文件内容写入压缩文件中
			fileWriter, err := zipWriter.CreateHeader(fileHeader)
			if err != nil {
				return err
			}
			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(fileWriter, file)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// 关闭 zipWriter
	err = zipWriter.Close()
	if err != nil {
		log.Fatal(err)
	}

	// 初始化 MinIO 客户端
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	// 将压缩后的文件上传到 MinIO
	_, err = minioClient.PutObject(context.Background(), bucketName, zipFileName, &buf, int64(buf.Len()), minio.PutObjectOptions{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("File uploaded successfully to MinIO!")
}
