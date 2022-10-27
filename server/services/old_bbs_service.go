package services

import (
	"gorm.io/gorm"

	"bbs-go/model"
)

var OldBBSService = newOldBBSService()

const oldBBSNodeId int64 = 114514

func newOldBBSService() *oldBBSService {
	return &oldBBSService{
		forum_name_map: make(map[int64]string),
	}
}

type oldBBSService struct {
	DB *gorm.DB

	// 缓存论坛名称
	forum_name_map map[int64]string
}

type oldPost struct {
	PostId    int64  `gorm:"column:tid"`
	ForumId   int64  `gorm:"column:fid"`
	MsgId     int64  `gorm:"column:pid"`
	Author    string `gorm:"column:author"`
	Timestamp int64  `gorm:"column:dateline"`
	Invisible int    `gorm:"column:invisible"`
	Title     string `gorm:"column:subject"`
	Content   string `gorm:"column:message"`
	Floor     int    `gorm:"column:position"`
	First     bool   `gorm:"column:first"`
}

type oldComment struct {
	Id        int64  `gorm:"column:id"`
	PostId    int64  `gorm:"column:tid"`
	MsgId     int64  `gorm:"column:pid"`
	Author    string `gorm:"column:author"`
	Timestamp int64  `gorm:"column:dateline"`
	Content   string `gorm:"column:comment"`
}

type oldForum struct {
	Id      string `gorm:"column:fid"`
	SuperId string `gorm:"column:fup"`
	Name    string `gorm:"column:name"`
}

func (r *oldBBSService) GetTopic(id int64) *model.Topic {
	post := oldPost{}
	if r.DB.Table("qsc_bbs_forum_post").Where("tid = ?", id).Where("first = 1").Take(&post).Error != nil {
		return nil
	}
	var cnt int64
	r.DB.Table("qsc_bbs_forum_post").Where("tid = ?", id).Where("first = 1").Count(&cnt)
	topic := model.Topic{
		IsOldBBS:     true,
		Model:        model.Model{Id: id},
		Title:        post.Title,
		UserId:       -1,
		Author:       post.Author,
		Content:      post.Content,
		NodeId:       oldBBSNodeId,
		CommentCount: cnt,
		CreateTime:   post.Timestamp,
		Forum:        r.getForumName(post.ForumId),
	}
	return &topic
}

func (r *oldBBSService) GetComments(TopicId int64, cursor int, limit int) (comments []model.Comment, nextCursor int, hasMore bool) {
	posts := []oldPost{}
	comments = []model.Comment{}
	if r.DB.Table("qsc_bbs_forum_post").Where("tid = ?", TopicId).Order("position").Where("first = 0").Limit(limit).Offset(cursor).Find(&posts).Error != nil {
		comments = nil
		return
	}
	for _, post := range posts {
		var cnt int64
		r.DB.Table("qsc_bbs_forum_postcomment").Where("pid = ?", post.MsgId).Count(&cnt)
		comments = append(comments, model.Comment{
			Model:        model.Model{Id: post.MsgId},
			UserId:       -1,
			EntityType:   "topic",
			EntityId:     TopicId,
			Content:      post.Content,
			CommentCount: cnt,
			CreateTime:   post.Timestamp,
			IsOldBBS:     true,
		})
	}
	nextCursor = cursor + len(comments)
	hasMore = len(comments) == 0
	return
}

func (r *oldBBSService) GetReplies(CommentId int64, cursor int, limit int) (comments []model.Comment, nextCursor int, hasMore bool) {
	post_cmts := []oldComment{}
	comments = []model.Comment{}
	if r.DB.Table("qsc_bbs_forum_postcomment").Where("pid = ?", CommentId).Order("dateline").Limit(limit).Offset(cursor).Find(&post_cmts).Error != nil {
		comments = nil
		return
	}
	for _, cmt := range post_cmts {
		comments = append(comments, model.Comment{
			IsOldBBS:     true,
			Model:        model.Model{Id: cmt.Id},
			UserId:       -1,
			Author:       cmt.Author,
			EntityType:   "comment",
			EntityId:     CommentId,
			QuoteId:      CommentId,
			Content:      cmt.Content,
			CommentCount: 0,
			CreateTime:   cmt.Timestamp,
		})
	}
	nextCursor = cursor + len(comments)
	hasMore = len(comments) == 0
	return
}

func (r *oldBBSService) getForumName(fid int64) string {
	if name, ok := r.forum_name_map[fid]; ok {
		return name
	}

	name := ""
	forum := oldForum{}
	r.DB.Table("qsc_bbs_forum_forum").Where("fid = ?", fid).Take(&forum)
	name = forum.Name

	if forum.SuperId != "" {
		r.DB.Table("qsc_bbs_forum_forum").Where("fid = ?", forum.SuperId).Take(&forum)
		name = forum.Name + "-" + name
	}

	r.forum_name_map[fid] = name
	return name
}
