package services

import (
	"bbs-go/model/constants"
	"bbs-go/pkg/event"
	"errors"
	"strings"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/common/jsons"
	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"bbs-go/model"
	"bbs-go/repositories"
)

var CommentService = newCommentService()

func newCommentService() *commentService {
	return &commentService{}
}

type commentService struct {
}

func (s *commentService) Get(id int64) *model.Comment {
	return repositories.CommentRepository.Get(sqls.DB(), id)
}

func (s *commentService) Take(where ...interface{}) *model.Comment {
	return repositories.CommentRepository.Take(sqls.DB(), where...)
}

func (s *commentService) Find(cnd *sqls.Cnd) []model.Comment {
	return repositories.CommentRepository.Find(sqls.DB(), cnd)
}

func (s *commentService) FindOne(cnd *sqls.Cnd) *model.Comment {
	return repositories.CommentRepository.FindOne(sqls.DB(), cnd)
}

func (s *commentService) FindPageByParams(params *params.QueryParams) (list []model.Comment, paging *sqls.Paging) {
	return repositories.CommentRepository.FindPageByParams(sqls.DB(), params)
}

func (s *commentService) FindPageByCnd(cnd *sqls.Cnd) (list []model.Comment, paging *sqls.Paging) {
	return repositories.CommentRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *commentService) FindUserCommentsPage(userId int64, page int, ascOrder bool) ([]model.Comment, *sqls.Paging) {
	cnd := sqls.NewCnd().Eq("user_id", userId).
		Where("status = ? or status = ?", constants.StatusOk, constants.StatusDeleted).
		Page(page, 20)
	if ascOrder {
		cnd.Asc("id")
	} else {
		cnd.Desc("id")
	}
	return s.FindPageByCnd(cnd)
}

func (s *commentService) Count(cnd *sqls.Cnd) int64 {
	return repositories.CommentRepository.Count(sqls.DB(), cnd)
}

func (s *commentService) Create(t *model.Comment) error {
	return repositories.CommentRepository.Create(sqls.DB(), t)
}

func (s *commentService) Update(t *model.Comment) error {
	return repositories.CommentRepository.Update(sqls.DB(), t)
}

func (s *commentService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.CommentRepository.Updates(sqls.DB(), id, columns)
}

func (s *commentService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.CommentRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *commentService) Delete(id int64) error {
	return repositories.CommentRepository.UpdateColumn(sqls.DB(), id, "status", constants.StatusDeleted)
}

func (s *commentService) CanManage(user *model.User, comment *model.Comment) bool {
	if user == nil || comment == nil || comment.IsOldBBS {
		return false
	}
	if user.Id == comment.UserId {
		return true
	}
	return s.canModerate(user, comment)
}

func (s *commentService) CanDelete(user *model.User, comment *model.Comment) bool {
	if user == nil || comment == nil || comment.IsOldBBS {
		return false
	}
	return s.canModerate(user, comment)
}

func (s *commentService) canModerate(user *model.User, comment *model.Comment) bool {
	if !user.IsAdminUserOrHigher() {
		return false
	}

	entityType := comment.EntityType
	entityId := comment.EntityId
	for depth := 0; entityType == constants.EntityComment && depth < 10; depth++ {
		parent := s.Get(entityId)
		if parent == nil {
			return false
		}
		entityType = parent.EntityType
		entityId = parent.EntityId
	}

	if entityType == constants.EntityTopic {
		topic := TopicService.Get(entityId)
		return topic != nil && user.CanManageTopic(topic)
	}
	if entityType == constants.EntityArticle {
		return user.IsAdminUserOrHigher()
	}
	return false
}

func (s *commentService) Edit(commentId, editorUserId int64, content string, imageList []model.ImageDTO) error {
	content = strings.TrimSpace(content)
	if strs.IsBlank(content) {
		return errors.New("请输入评论内容")
	}
	if strs.RuneLen(content) > constants.ContentMaxLen {
		return errors.New("评论内容长度不能超过5000个字符")
	}

	imageListStr := ""
	if len(imageList) > 0 {
		var err error
		imageListStr, err = jsons.ToStr(imageList)
		if err != nil {
			return err
		}
	}

	if err := repositories.CommentRepository.Updates(sqls.DB(), commentId, map[string]interface{}{
		"content":           content,
		"image_list":        imageListStr,
		"last_edit_user_id": editorUserId,
		"last_edit_time":    dates.NowTimestamp(),
	}); err != nil {
		return err
	}
	comment := s.Get(commentId)
	if comment != nil {
		if err := FileService.BindCommentFiles(commentId, comment.UserId, content, imageList); err != nil {
			logrus.Error("error associating uploaded files with edited comment: ", err)
		}
	}
	return nil
}

func (s *commentService) DeleteWithCounts(comment *model.Comment) error {
	if comment == nil || comment.Status != constants.StatusOk {
		return nil
	}

	err := sqls.DB().Transaction(func(tx *gorm.DB) error {
		if err := repositories.CommentRepository.UpdateColumn(tx, comment.Id, "status", constants.StatusDeleted); err != nil {
			return err
		}
		if comment.EntityType == constants.EntityTopic {
			return tx.Model(&model.Topic{}).Where("id = ?", comment.EntityId).
				UpdateColumn("comment_count", gorm.Expr("CASE WHEN comment_count > 0 THEN comment_count - 1 ELSE 0 END")).Error
		}
		if comment.EntityType == constants.EntityComment {
			return tx.Model(&model.Comment{}).Where("id = ?", comment.EntityId).
				UpdateColumn("comment_count", gorm.Expr("CASE WHEN comment_count > 0 THEN comment_count - 1 ELSE 0 END")).Error
		}
		return nil
	})
	if err == nil {
		if unbindErr := FileService.UnbindCommentFiles(comment.Id); unbindErr != nil {
			logrus.Error("error unbinding comment files: ", unbindErr)
		}
		UserService.DecrCommentCount(comment.UserId)
	}
	return err
}

// Publish 发表评论
func (s *commentService) Publish(userId int64, form model.CreateCommentForm) (*model.Comment, error) {
	form.Content = strings.TrimSpace(form.Content)
	if strs.IsBlank(form.EntityType) {
		return nil, errors.New("参数非法")
	}
	if form.EntityId <= 0 {
		return nil, errors.New("参数非法")
	}
	if strs.IsBlank(form.Content) {
		return nil, errors.New("请输入评论内容")
	}
	if strs.RuneLen(form.Content) > constants.ContentMaxLen {
		return nil, errors.New("评论内容长度不能超过5000个字符")
	}
	if form.EntityType == constants.EntityComment {
		parent := s.Get(form.EntityId)
		if parent == nil || parent.Status != constants.StatusOk {
			return nil, errors.New("被回复的评论不存在或已被删除")
		}
	}

	comment := &model.Comment{
		UserId:      userId,
		EntityType:  form.EntityType,
		EntityId:    form.EntityId,
		Content:     form.Content,
		ContentType: strs.DefaultIfBlank(form.ContentType, constants.ContentTypeMarkdown),
		QuoteId:     form.QuoteId,
		Status:      constants.StatusOk,
		UserAgent:   form.UserAgent,
		Ip:          form.Ip,
		CreateTime:  dates.NowTimestamp(),
	}

	logrus.Info("创建了这样一个评论：", comment)

	if len(form.ImageList) > 0 {
		imageListStr, err := jsons.ToStr(form.ImageList)
		if err == nil {
			comment.ImageList = imageListStr
		} else {
			logrus.Error(err)
		}
	}

	err := sqls.DB().Transaction(func(tx *gorm.DB) error {
		if err := repositories.CommentRepository.Create(tx, comment); err != nil {
			return err
		}

		if form.EntityType == constants.EntityTopic {
			if err := TopicService.onComment(tx, form.EntityId, comment); err != nil {
				return err
			}
		} else if form.EntityType == constants.EntityComment { // 二级评论
			if err := s.onComment(tx, comment); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	if bindErr := FileService.BindCommentFiles(comment.Id, userId, form.Content, form.ImageList); bindErr != nil {
		logrus.Error("error associating uploaded files with comment: ", bindErr)
	}

	// 用户跟帖计数
	UserService.IncrCommentCount(userId)
	// 获得积分
	UserService.IncrScoreForPostComment(comment)
	// 发送事件
	event.Send(event.CommentCreateEvent{
		UserId:    userId,
		CommentId: comment.Id,
	})

	return comment, nil
}

// onComment 评论被回复（二级评论）
func (s *commentService) onComment(tx *gorm.DB, comment *model.Comment) error {
	return repositories.CommentRepository.UpdateColumn(tx, comment.EntityId, "comment_count", gorm.Expr("comment_count + 1"))
}

// // 统计数量
// func (s *commentService) Count(entityType string, entityId int64) int64 {
// 	var count int64 = 0
// 	sqls.DB().Model(&model.Comment{}).Where("entity_type = ? and entity_id = ?", entityType, entityId).Count(&count)
// 	return count
// }

func (s *commentService) GetCommentsPage(entityType string, entityId int64, page int, ascOrder bool) ([]model.Comment, *sqls.Paging) {
	if page < 1 {
		page = 1
	}
	cnd := sqls.NewCnd().Eq("entity_type", entityType).Eq("entity_id", entityId).
		Where("(status = ? or (status = ? and comment_count > 0))", constants.StatusOk, constants.StatusDeleted).
		Page(page, 10)
	if ascOrder {
		cnd.Asc("id")
	} else {
		cnd.Desc("id")
	}
	return s.FindPageByCnd(cnd)
}

// GetComments 列表
func (s *commentService) GetComments(entityType string, entityId int64, cursor int64, ascOrder bool) (comments []model.Comment, nextCursor int64, hasMore bool) {
	limit := 20
	var cnd *sqls.Cnd
	if ascOrder {
		cnd = sqls.NewCnd().Eq("entity_type", entityType).Eq("entity_id", entityId).
			Where("(status = ? or (status = ? and comment_count > 0))", constants.StatusOk, constants.StatusDeleted).
			Asc("id").Limit(limit)
		if cursor > 0 {
			cnd.Gt("id", cursor)
		}
	} else {
		cnd = sqls.NewCnd().Eq("entity_type", entityType).Eq("entity_id", entityId).
			Where("(status = ? or (status = ? and comment_count > 0))", constants.StatusOk, constants.StatusDeleted).
			Desc("id").Limit(limit)
		if cursor > 0 {
			cnd.Lt("id", cursor)
		}
	}

	comments = repositories.CommentRepository.Find(sqls.DB(), cnd)
	if len(comments) > 0 {
		nextCursor = comments[len(comments)-1].Id
		hasMore = len(comments) >= limit
	} else {
		nextCursor = cursor
	}
	return
}

// GetReplies 二级回复列表
func (s *commentService) GetReplies(commentId int64, cursor int64, limit int) (comments []model.Comment, nextCursor int64, hasMore bool) {
	cnd := sqls.NewCnd().Eq("entity_type", constants.EntityComment).Eq("entity_id", commentId).Eq("status", constants.StatusOk).Asc("id").Limit(limit)
	if cursor > 0 {
		cnd.Gt("id", cursor)
	}
	comments = s.Find(cnd)
	if len(comments) > 0 {
		nextCursor = comments[len(comments)-1].Id
		hasMore = len(comments) >= limit
	} else {
		nextCursor = cursor
	}
	return
}

// ScanByUser 按照用户扫描数据
func (s *commentService) ScanByUser(userId int64, callback func(comments []model.Comment)) {
	var cursor int64 = 0
	for {
		list := repositories.CommentRepository.Find(sqls.DB(), sqls.NewCnd().
			Eq("user_id", userId).Gt("id", cursor).Asc("id").Limit(1000))
		if len(list) == 0 {
			break
		}
		cursor = list[len(list)-1].Id
		callback(list)
	}
}

func (s *commentService) IsCommented(userId int64, entityType string, entityId int64) bool {
	return s.FindOne(sqls.NewCnd().Where("user_id = ? and entity_id = ? and entity_type = ? and status = ?", userId, entityId, entityType, constants.StatusOk)) != nil
}
