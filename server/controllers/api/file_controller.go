package api

import (
	"bbs-go/model"
	"bbs-go/pkg/config"
	"bbs-go/services"
	"errors"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"github.com/minio/minio-go/v6"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
	"github.com/sirupsen/logrus"
)

var (
	minioClient *minio.Client
	bucketName  string
)

type FileController struct {
	Ctx iris.Context
}

// Offical Docs
// http://docs.minio.org.cn/docs/master/golang-client-api-reference#PutObject

func InitMinio(conf *config.Config) {
	// 初使化minio client对象。
	minioClient, err := minio.New(
		conf.MinIO.Endpoint,
		conf.MinIO.AccessKeyID,
		conf.MinIO.SecretAccessKey,
		conf.MinIO.UseSSL,
	)
	if err != nil {
		logrus.Fatal("Fail to link to MinIO")
		return
	}
	exists, err := minioClient.BucketExists(bucketName)
	if err == nil && exists {
		logrus.Info("We already own a bucket called %s\n", bucketName)
	} else {
		if err != nil {
			logrus.Fatal("Fail to find exist bucket", err)
			return
		}
		err = minioClient.MakeBucket(bucketName, conf.MinIO.BucketLocation)
		if err != nil {
			logrus.Fatal("Fail to create bucket", err)
			return
		}
	}
}

func PostUpload(c *FileController) *web.JsonResult {
	fileString := params.FormValueDefault(c.Ctx, "file", "")
	fileNameOri := params.FormValueDefault(c.Ctx, "file_name", "")

	if fileString == "" {
		logrus.Error("No file are received!")
		return web.JsonError(errors.New("no file received"))
	}

	// to io.Reader
	file := strings.NewReader(fileString)

	fileUUID := uuid.New().String()

	bytes, err := minioClient.PutObject(bucketName, fileUUID, file, file.Size(), minio.PutObjectOptions{})
	if err != nil {
		logrus.Error("error happen when put object to minio: %s", err)
		return web.JsonError(errors.New("error happen when put object to minio"))
	}

	logrus.Info("finish put object with %d bytes to minio", bytes)

	newFile := &model.FileRecord{
		FileName:   fileNameOri,
		FileUUID:   fileUUID,
		BucketName: bucketName,
	}

	err = services.FileService.CreateRecord(newFile)
	if err != nil {
		logrus.Error("error happen when recording the file: %s", err)
		return web.JsonError(errors.New("error happen when recording the file"))
	}

	return web.JsonData(newFile)
}

func PostDownload(c *FileController) *web.JsonResult {
	fileId := params.FormValueInt64Default(c.Ctx, "file_id", -1)

	if fileId == -1 {
		logrus.Error("empty fileId!")
		return web.JsonError(errors.New("empty file_id"))
	}

	fileRecord := services.FileService.Get(fileId)

	if fileRecord == nil {
		logrus.Error("no such file")
		return web.JsonError(errors.New("no such file"))
	}

	bucket := fileRecord.BucketName
	fileUUID := fileRecord.FileUUID

	object, err := minioClient.GetObject(bucket, fileUUID, minio.GetObjectOptions{})
	if err != nil {
		logrus.Error("error happen when get object from minio: %s", err)
		return web.JsonError(errors.New("error happen when get object from minio"))
	}

	fileString, err := io.ReadAll(object)
	if err != nil {
		logrus.Error("error happen when change file reader to string: %s", err)
		return web.JsonError(errors.New("error happen when change file reader to string"))
	}

	resp := struct {
		File string `json:"file"`
	}{
		File: string(fileString),
	}

	return web.JsonData(resp)
}
