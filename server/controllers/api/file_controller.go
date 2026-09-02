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
	"mime"
	"path/filepath"
	"strconv"
	"strings"
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
	user := services.UserTokenService.GetCurrent(c.Ctx)
	if err := services.UserService.CheckPostStatus(user); err != nil {
		return web.JsonError(err)
	}

	file, info, err := c.Ctx.FormFile("file")

	if err != nil {
		logrus.Error("error happen when get multipart file: ", err)
		return web.JsonError(err)
	}
	defer file.Close()

	fileSize := info.Size

	if fileSize > constants.UploadMaxBytes {
		return web.JsonErrorMsg("文件不能超过" + strconv.Itoa(constants.UploadMaxM) + "M")
	}

	newFile, err := putFile(file, info.Filename, fileSize, user.Id)

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

	sourceType := strings.TrimSpace(c.Ctx.FormValue("source"))
	record, err := putFile(file, header.Filename, header.Size, user.Id, sourceType)
	if err != nil {
		return web.JsonError(err)
	}

	host := config.Instance.BaseUrl
	url := fmt.Sprintf("%s/api/file/preview/%s#", host, record.FileUUID)

	return web.NewEmptyRspBuilder().Put("url", url).JsonResult()
}

func (c *FileController) GetDownloadBy(fileId string) {
	c.serveFile(fileId, false)
}

// GetPreviewBy serves a file inline so browsers and the MinIO-style admin
// file list can preview images and other browser-supported formats.
func (c *FileController) GetPreviewBy(fileId string) {
	c.serveFile(fileId, true)
}

func (c *FileController) serveFile(fileId string, inline bool) {

	if fileId == "" {
		logrus.Error("empty fileId!")
		c.Ctx.StatusCode(400)
		return
	}

	object, fileRecord, err := getFile(fileId)
	if err != nil {
		c.Ctx.StatusCode(400)
		return
	}

	if inline {
		c.Ctx.Header("Content-Type", previewContentType(fileRecord))
		c.Ctx.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", fileRecord.FileName))
	} else {
		c.Ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileRecord.FileName))
	}
	c.Ctx.ServeContent(object, fileRecord.FileName, time.Now())
	object.Close()
}

func putFile(file io.Reader, fineName string, fileSize int64, userId int64, sourceTypes ...string) (*model.FileRecord, error) {
	sourceType := "unattached"
	if len(sourceTypes) > 0 {
		switch sourceTypes[0] {
		case "avatar", "background", "node_logo", "link_logo":
			sourceType = sourceTypes[0]
		}
	}
	fileUUID := uuid.New().String()
	ext := strings.ToLower(filepath.Ext(fineName))
	if len(ext) > 16 {
		ext = ""
	}
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	prefix := "file"
	if strings.HasPrefix(contentType, "image/") {
		prefix = "image"
	}
	objectName := prefix + "/" + time.Now().Format("2006/0102") + "/" + fileUUID + ext

	bytes, err := globalMinioClient.PutObject(bucketName, objectName, file, fileSize, minio.PutObjectOptions{ContentType: contentType})

	if err != nil {
		logrus.Errorf("error happen when put object to minio: %s", err)
		return nil, err
	}

	logrus.Infof("finish put object with %d bytes to minio", bytes)

	newFile := &model.FileRecord{
		FileName:    fineName,
		FileUUID:    fileUUID,
		FileSize:    fileSize,
		BucketName:  bucketName,
		ObjectName:  objectName,
		ContentType: contentType,
		UserId:      userId,
		SourceType:  sourceType,
		Managed:     true,
		CreateTime:  time.Now().UnixMilli(),
	}

	err = services.FileService.CreateRecord(newFile)
	if err != nil {
		logrus.Errorf("error happen when recording the file: %s", err)
		return nil, err
	}
	return newFile, err
}

func getFile(fileId string) (*minio.Object, *model.FileRecord, error) {
	fileRecord := repositories.FileRepository.GetByUUID(sqls.DB(), fileId)

	if fileRecord == nil {
		logrus.Error("no such file")
		return nil, nil, errors.New("no such file")
	}

	object, err := globalMinioClient.GetObject(
		fileRecord.BucketName,
		objectName(fileRecord),
		minio.GetObjectOptions{},
	)

	if err != nil {
		logrus.Error(fmt.Sprintf("error happen when get object from minio: %s", err))
		return nil, nil, errors.New("error happen when get object from minio")
	}

	return object, fileRecord, nil
}

func objectName(fileRecord *model.FileRecord) string {
	if fileRecord.ObjectName != "" {
		return fileRecord.ObjectName
	}
	// Records created before object names were introduced used the UUID as key.
	return fileRecord.FileUUID
}

func previewContentType(fileRecord *model.FileRecord) string {
	if fileRecord.ContentType != "" {
		return fileRecord.ContentType
	}
	contentType := mime.TypeByExtension(filepath.Ext(fileRecord.FileName))
	if contentType == "" {
		return "application/octet-stream"
	}
	return contentType
}

func deleteObject(fileRecord *model.FileRecord) error {
	return globalMinioClient.RemoveObject(fileRecord.BucketName, objectName(fileRecord))
}

func RemoveObject(bucket, object, fallback string) error {
	if object == "" {
		object = fallback
	}
	return globalMinioClient.RemoveObject(bucket, object)
}
