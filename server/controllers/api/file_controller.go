package api

import (
	"bbs-go/model"
	"bbs-go/pkg/config"
	"bbs-go/services"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"github.com/minio/minio-go/v6"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
	"github.com/sirupsen/logrus"
)

var (
	minioClient *minio.Client
)

// Actually a random number is better.
const bucketName = "qscbbsbucket"

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
		logrus.Fatal("Fail to create MinIO Client")
		return
	}
	exists, err := minioClient.BucketExists(bucketName)
	if err == nil && exists {
		logrus.Info("We already own a bucket called %s\n", bucketName)
	} else {
		if err != nil {
			logrus.Fatal("Fail to find exist bucket: ", err)
			return
		}
		err = minioClient.MakeBucket(bucketName, conf.MinIO.BucketLocation)
		if err != nil {
			logrus.Fatal("Fail to create bucket:", err)
			return
		}
	}
}

func PostUpload(c *FileController) *web.JsonResult {
	file, info, err := c.Ctx.FormFile("uploadfile")

	if err != nil {
		logrus.Info("error happen when get multipart file: ", err)
		return web.JsonError(err)
	}

	fileNameOri := info.Filename
	fileSize := info.Size
	fileUUID := uuid.New().String()

	bytes, err := minioClient.PutObject(bucketName, fileUUID, file, fileSize, minio.PutObjectOptions{})
	if err != nil {
		logrus.Error("error happen when put object to minio: %s", err)
		return web.JsonError(errors.New("error happen when put object to minio"))
	}

	logrus.Info("finish put object with %d bytes to minio", bytes)

	newFile := &model.FileRecord{
		FileName:   fileNameOri,
		FileUUID:   fileUUID,
		FileSize:   fileSize,
		BucketName: bucketName,
	}

	err = services.FileService.CreateRecord(newFile)
	if err != nil {
		logrus.Error("error happen when recording the file: %s", err)
		return web.JsonError(errors.New("error happen when recording the file"))
	}

	return web.JsonData(newFile)
}

func GetDownload(c *FileController) {
	fileId := params.FormValueInt64Default(c.Ctx, "file_id", -1)

	if fileId == -1 {
		logrus.Error("empty fileId!")
		c.Ctx.StatusCode(400)
		return
	}

	fileRecord := services.FileService.Get(fileId)

	if fileRecord == nil {
		logrus.Error("no such file")
		c.Ctx.StatusCode(400)
		return
	}

	bucket := fileRecord.BucketName
	fileUUID := fileRecord.FileUUID

	object, err := minioClient.GetObject(bucket, fileUUID, minio.GetObjectOptions{})
	if err != nil {
		logrus.Error("error happen when get object from minio: %s", err)
		c.Ctx.StatusCode(400)
		return
	}

	if err != nil {
		logrus.Error("error happen when change file reader to string: %s", err)
		c.Ctx.StatusCode(400)
		return
	}

	c.Ctx.ServeContent(object, fileRecord.FileName, time.Now())
	return
}
