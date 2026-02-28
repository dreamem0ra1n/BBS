package api

import (
	"bbs-go/model"
	"bbs-go/model/constants"
	"bbs-go/pkg/config"
	"bbs-go/repositories"
	"bbs-go/services"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	_ "image/png"

	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"github.com/minio/minio-go/v6"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web"
	"github.com/sirupsen/logrus"
)

// Actually a random number is better.
const bucketName = "qscbbsbucket"

type FileController struct {
	Ctx iris.Context
}

// Global MinIO client reused across all file operations.
// Creating a new client per request leaks connections and exhausts resources.
var globalMinioClient *minio.Client

// Offical Docs
// http://docs.minio.org.cn/docs/master/golang-client-api-reference#PutObject

func InitMinio(conf *config.Config) {
	// 初使化minio client对象。
	var err error
	globalMinioClient, err = minio.New(
		conf.MinIO.Endpoint,
		conf.MinIO.AccessKeyID,
		conf.MinIO.SecretAccessKey,
		conf.MinIO.UseSSL,
	)

	if err != nil {
		logrus.Fatal("Fail to create MinIO Client: ", err)
		return
	}
	exists, err := globalMinioClient.BucketExists(bucketName)
	if err == nil && exists {
		logrus.Info(fmt.Sprintf("We already own a bucket called %s\n", bucketName))
	} else {
		if err != nil {
			logrus.Fatal("Fail to find exist bucket: ", err)
			return
		}
		err = globalMinioClient.MakeBucket(bucketName, conf.MinIO.BucketLocation)
		if err != nil {
			logrus.Fatal("Fail to create bucket:", err)
			return
		}
	}
}

func (c *FileController) PostUpload() *web.JsonResult {
	file, info, err := c.Ctx.FormFile("file")

	if err != nil {
		logrus.Error("error happen when get multipart file: ", err)
		return web.JsonError(err)
	}

	fileSize := info.Size

	if fileSize > constants.UploadMaxBytes {
		return web.JsonErrorMsg("文件不能超过" + strconv.Itoa(constants.UploadMaxM) + "M")
	}

	newFile, err := putFile(file, info.Filename, fileSize)

	if err != nil {
		logrus.Error("error happen when upload files: ", err)
		return web.JsonError(err)
	}

	return web.JsonData(newFile)
}

func (c *FileController) PostUploadImg() *web.JsonResult {
	user := services.UserTokenService.GetCurrent(c.Ctx)
	if err := services.UserService.CheckPostStatus(user); err != nil {
		return web.JsonError(err)
	}

	file, header, err := c.Ctx.FormFile("image")
	if err != nil {
		return web.JsonError(err)
	}
	defer file.Close()

	if header.Size > constants.UploadMaxBytes {
		return web.JsonErrorMsg("图片不能超过" + strconv.Itoa(constants.UploadMaxM) + "M")
	}

	record, err := putFile(file, header.Filename, header.Size)
	if err != nil {
		return web.JsonError(err)
	}

	host := config.Instance.BaseUrl
	url := fmt.Sprintf("%s/api/file/download/%s#", host, record.FileUUID)

	return web.NewEmptyRspBuilder().Put("url", url).JsonResult()
}

func (c *FileController) GetDownloadBy(fileId string) {

	if fileId == "" {
		logrus.Error("empty fileId!")
		c.Ctx.StatusCode(400)
		return
	}

	object, fileName, err := getFile(fileId)
	if err != nil {
		c.Ctx.StatusCode(400)
		return
	}

	c.Ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	c.Ctx.ServeContent(object, fileName, time.Now())
	object.Close()
}

func putFile(file io.Reader, fineName string, fileSize int64) (*model.FileRecord, error) {
	fileUUID := uuid.New().String()

	bytes, err := globalMinioClient.PutObject(bucketName, fileUUID, file, fileSize, minio.PutObjectOptions{ContentType: "application/octet-stream"})

	if err != nil {
		logrus.Error("error happen when put object to minio: %s", err)
		return nil, err
	}

	logrus.Info("finish put object with %d bytes to minio", bytes)

	newFile := &model.FileRecord{
		FileName:   fineName,
		FileUUID:   fileUUID,
		FileSize:   fileSize,
		BucketName: bucketName,
	}

	err = services.FileService.CreateRecord(newFile)
	if err != nil {
		logrus.Error("error happen when recording the file: %s", err)
		return nil, err
	}
	return newFile, err
}

func getFile(fileId string) (*minio.Object, string, error) {
	fileRecord := repositories.FileRepository.GetByUUID(sqls.DB(), fileId)

	if fileRecord == nil {
		logrus.Error("no such file")
		return nil, "", errors.New("no such file")
	}

	object, err := globalMinioClient.GetObject(
		fileRecord.BucketName,
		fileRecord.FileUUID,
		minio.GetObjectOptions{},
	)

	if err != nil {
		logrus.Error(fmt.Sprintf("error happen when get object from minio: %s", err))
		return nil, "", errors.New("error happen when get object from minio")
	}

	return object, fileRecord.FileName, nil
}
