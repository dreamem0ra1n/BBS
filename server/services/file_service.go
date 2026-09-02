package services

import (
	"errors"
	"regexp"
	"strings"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"

	"bbs-go/model"
	"bbs-go/repositories"
)

var FileService = newFileService()

func newFileService() *fileService {
	return &fileService{}
}

type fileService struct{}

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

func (s *fileService) Delete(id int64, removeObject func(*model.FileRecord) error) error {
	file := s.Get(id)
	if file == nil {
		return errors.New("文件不存在")
	}
	if file.TopicId > 0 || file.CommentId > 0 || (file.SourceType != "unattached" && file.SourceType != "topic" && file.SourceType != "comment") {
		return errors.New("文件已关联，不能删除")
	}
	if err := removeObject(file); err != nil {
		return err
	}
	repositories.FileRepository.Delete(sqls.DB(), id)
	return nil
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
