package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"bbs-go/model"
	"bbs-go/repositories"
)

var FileService = newFileService()

var errFileReferenced = errors.New("文件仍被引用")

type FileCleanupResult struct {
	Scanned    int
	Deleted    int
	Referenced int
	Failed     int
	FreedBytes int64
}

func findFileReference(ctx context.Context, db *gorm.DB, file *model.FileRecord) (string, error) {
	if file.FileUUID == "" {
		return "", errors.New("文件标识为空，无法核验引用")
	}
	checks := []struct {
		model   interface{}
		columns []string
		label   string
	}{
		{&model.User{}, []string{"avatar"}, "用户头像"},
		{&model.User{}, []string{"background_image"}, "用户背景图"},
		{&model.User{}, []string{"description", "home_page"}, "用户简介或主页"},
		{&model.TopicNode{}, []string{"logo", "description"}, "节点图标或描述"},
		{&model.Link{}, []string{"logo", "url", "summary"}, "友情链接"},
		{&model.SysConfig{}, []string{"value"}, "系统配置"},
		{&model.Topic{}, []string{"content", "image_list", "hide_content", "extra_data"}, "帖子（含已删除帖子）"},
		{&model.Comment{}, []string{"content", "image_list"}, "评论（含已删除评论）"},
		{&model.Article{}, []string{"content", "summary", "source_url"}, "文章（含已删除文章）"},
		{&model.Message{}, []string{"content", "quote_content", "extra_data"}, "站内消息"},
		{&model.BirthdayBlessing{}, []string{"content"}, "生日祝福"},
	}
	identifier := strings.ToLower(file.FileUUID)
	for _, check := range checks {
		conditions := make([]string, 0, len(check.columns))
		arguments := make([]interface{}, 0, len(check.columns))
		for _, column := range check.columns {
			conditions = append(conditions, "INSTR(LOWER("+column+"), ?) > 0")
			arguments = append(arguments, identifier)
		}
		var reference struct{ Id int64 }
		err := db.WithContext(ctx).Unscoped().Model(check.model).Select("id").
			Where(strings.Join(conditions, " OR "), arguments...).Take(&reference).Error
		if err == nil {
			return fmt.Sprintf("%s #%d", check.label, reference.Id), nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return "", err
		}
	}
	return "", nil
}

func newFileService() *fileService {
	return &fileService{}
}

type fileService struct {
	cleanupMutex sync.Mutex
}

var fileUUIDPattern = regexp.MustCompile(`(?i)[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}`)

func (s *fileService) Get(id int64) *model.FileRecord {
	return repositories.FileRepository.Get(sqls.DB(), id)
}

func (s *fileService) FindPageByParams(query *params.QueryParams) ([]model.FileRecord, *sqls.Paging) {
	return repositories.FileRepository.FindPageByParams(sqls.DB(), query)
}

func (s *fileService) CreateRecord(t *model.FileRecord) error {
	return repositories.FileRepository.Create(sqls.DB(), t)
}

func (s *fileService) Delete(ctx context.Context, id int64, removeObject func(*model.FileRecord) error) error {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	return sqls.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var file model.FileRecord
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&file, id).Error; err != nil {
			return err
		}
		reference, err := findFileReference(ctx, tx, &file)
		if err != nil {
			return fmt.Errorf("无法完成引用核验，未删除文件：%w", err)
		}
		if reference != "" {
			return fmt.Errorf("%w：%s，不能删除", errFileReferenced, reference)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := removeObject(&file); err != nil {
			return err
		}
		return tx.Delete(&file).Error
	})
}

func (s *fileService) CleanupPreviousDay(ctx context.Context, now time.Time, removeObject func(*model.FileRecord) error) (FileCleanupResult, error) {
	result := FileCleanupResult{}
	if !s.cleanupMutex.TryLock() {
		return result, errors.New("文件清理任务仍在运行")
	}
	defer s.cleanupMutex.Unlock()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	start := today.AddDate(0, 0, -1).UnixMilli()
	end := today.UnixMilli()
	var lastID int64
	for {
		var files []model.FileRecord
		if err := sqls.DB().WithContext(ctx).Select("id", "file_size").
			Where("create_time >= ? AND create_time < ? AND id > ?", start, end, lastID).
			Order("id ASC").Limit(100).Find(&files).Error; err != nil {
			return result, err
		}
		if len(files) == 0 {
			return result, nil
		}
		for _, file := range files {
			if err := ctx.Err(); err != nil {
				return result, err
			}
			lastID = file.Id
			result.Scanned++
			err := s.Delete(ctx, file.Id, removeObject)
			switch {
			case err == nil:
				result.Deleted++
				result.FreedBytes += file.FileSize
			case errors.Is(err, errFileReferenced):
				result.Referenced++
			case errors.Is(err, gorm.ErrRecordNotFound):
				continue
			default:
				result.Failed++
				logrus.WithError(err).WithField("fileId", file.Id).Warn("scheduled file cleanup failed")
			}
		}
	}
}

// BindTopicFiles refreshes the topic association for files referenced by its
// content or image list. Files are uploaded before a topic receives an ID, so
// this also identifies uploads that remain unattached.
func (s *fileService) BindTopicFiles(topicId, userId int64, content string, images []model.ImageDTO) error {
	uuids := extractFileUUIDs(content, images)
	db := sqls.DB()
	if err := db.Model(&model.FileRecord{}).Where("topic_id = ?", topicId).
		Updates(map[string]interface{}{"topic_id": 0, "comment_id": 0, "source_type": "unattached"}).Error; err != nil {
		return err
	}
	for fileUUID := range uuids {
		query := db.Model(&model.FileRecord{}).Where("file_uuid = ?", fileUUID)
		if userId > 0 {
			query = query.Where("user_id = ? OR user_id = 0", userId)
		}
		if err := query.Updates(map[string]interface{}{"topic_id": topicId, "source_type": "topic"}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *fileService) UnbindTopicFiles(topicId int64) error {
	return sqls.DB().Model(&model.FileRecord{}).Where("topic_id = ?", topicId).
		Updates(map[string]interface{}{"topic_id": 0, "comment_id": 0, "source_type": "unattached"}).Error
}

func (s *fileService) BindCommentFiles(commentId, userId int64, content string, images []model.ImageDTO) error {
	uuids := extractFileUUIDs(content, images)
	db := sqls.DB()
	if err := db.Model(&model.FileRecord{}).Where("comment_id = ?", commentId).
		Updates(map[string]interface{}{"comment_id": 0, "source_type": "unattached"}).Error; err != nil {
		return err
	}
	for fileUUID := range uuids {
		query := db.Model(&model.FileRecord{}).Where("file_uuid = ?", fileUUID)
		if userId > 0 {
			query = query.Where("user_id = ? OR user_id = 0", userId)
		}
		if err := query.Updates(map[string]interface{}{"topic_id": 0, "comment_id": commentId, "source_type": "comment"}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *fileService) UnbindCommentFiles(commentId int64) error {
	return sqls.DB().Model(&model.FileRecord{}).Where("comment_id = ?", commentId).
		Updates(map[string]interface{}{"topic_id": 0, "comment_id": 0, "source_type": "unattached"}).Error
}

// ClassifyReferencedFiles backfills uploads made before source metadata was
// introduced by matching their UUIDs against known URL fields.
func (s *fileService) ClassifyReferencedFiles() {
	db := sqls.DB()
	queries := []string{
		"UPDATE t_file_record f JOIN t_user u ON u.avatar LIKE CONCAT('%', f.file_uuid, '%') SET f.source_type = 'avatar' WHERE f.source_type = 'unattached' AND f.topic_id = 0 AND f.comment_id = 0",
		"UPDATE t_file_record f JOIN t_user u ON u.background_image LIKE CONCAT('%', f.file_uuid, '%') SET f.source_type = 'background' WHERE f.source_type = 'unattached' AND f.topic_id = 0 AND f.comment_id = 0",
		"UPDATE t_file_record f JOIN t_link l ON l.logo LIKE CONCAT('%', f.file_uuid, '%') SET f.source_type = 'link_logo' WHERE f.source_type = 'unattached' AND f.topic_id = 0 AND f.comment_id = 0",
		"UPDATE t_file_record f JOIN t_topic_node n ON n.logo LIKE CONCAT('%', f.file_uuid, '%') SET f.source_type = 'node_logo' WHERE f.source_type = 'unattached' AND f.topic_id = 0 AND f.comment_id = 0",
		"UPDATE t_file_record f JOIN t_topic t ON (t.content LIKE CONCAT('%', f.file_uuid, '%') OR t.image_list LIKE CONCAT('%', f.file_uuid, '%')) SET f.topic_id = t.id, f.comment_id = 0, f.source_type = 'topic' WHERE f.source_type = 'unattached' AND f.topic_id = 0 AND f.comment_id = 0",
		"UPDATE t_file_record f JOIN t_comment c ON (c.content LIKE CONCAT('%', f.file_uuid, '%') OR c.image_list LIKE CONCAT('%', f.file_uuid, '%')) SET f.topic_id = 0, f.comment_id = c.id, f.source_type = 'comment' WHERE f.source_type = 'unattached' AND f.topic_id = 0 AND f.comment_id = 0",
	}
	for _, query := range queries {
		if err := db.Exec(query).Error; err != nil {
			// A missing optional table/column should not make file browsing fail.
			continue
		}
	}
}

func extractFileUUIDs(content string, images []model.ImageDTO) map[string]struct{} {
	uuids := make(map[string]struct{})
	for _, value := range fileUUIDPattern.FindAllString(content, -1) {
		uuids[strings.ToLower(value)] = struct{}{}
	}
	for _, image := range images {
		for _, value := range fileUUIDPattern.FindAllString(image.Url, -1) {
			uuids[strings.ToLower(value)] = struct{}{}
		}
	}
	return uuids
}
