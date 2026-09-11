package services

import (
	"github.com/mlogclub/simple/sqls"
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

type oldPostMeta struct {
	PostId     int64 `gorm:"column:tid"`
	Permission int   `gorm:"column:readperm"`
}

type oldPostCount struct {
	PostId int64 `gorm:"column:tid"`
	Count  int64 `gorm:"column:cnt"`
}

type oldForum struct {
	Id      int64  `gorm:"column:fid"`
	SuperId int64  `gorm:"column:fup"`
	Name    string `gorm:"column:name"`
}

func (r *oldBBSService) post2topic(post oldPost) model.Topic {
	var cnt int64
	r.DB.Table("qsc_bbs_forum_post").Where("tid = ?", post.PostId).Where("first = 0").Count(&cnt)

	// 加入权限控制
	post_meta := oldPostMeta{}
	r.DB.Table("qsc_bbs_forum_thread").Where("tid = ?", post.PostId).Take(&post_meta)

	return r.buildTopic(post, cnt, post_meta.Permission)
}

// buildTopic 组装单条旧站主题，权限与回复数由调用方传入，便于批量场景复用。
func (r *oldBBSService) buildTopic(post oldPost, commentCount int64, permission int) model.Topic {
	return model.Topic{
		IsOldBBS:     true,
		Model:        model.Model{Id: post.PostId},
		Title:        post.Title,
		UserId:       -1,
		Author:       post.Author,
		Content:      post.Content,
		NodeId:       oldBBSNodeId,
		CommentCount: commentCount,
		CreateTime:   post.Timestamp,
		Forum:        r.getForumName(post.ForumId),
		AccessLv:     permission,
	}
}

// buildTopics 批量组装旧站主题。
// 原实现是每篇主题各查一次回复数和权限（N+1），这里改为两次 IN 查询后内存组装。
func (r *oldBBSService) buildTopics(posts []oldPost) []model.Topic {
	if len(posts) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(posts))
	for _, post := range posts {
		ids = append(ids, post.PostId)
	}

	// 回复数
	counts := make(map[int64]int64, len(posts))
	countRows := []oldPostCount{}
	r.DB.Table("qsc_bbs_forum_post").
		Select("tid, count(*) as cnt").
		Where("tid in ?", ids).
		Where("first = 0").
		Group("tid").
		Find(&countRows)
	for _, row := range countRows {
		counts[row.PostId] = row.Count
	}

	// 阅读权限
	permissions := make(map[int64]int, len(posts))
	metas := []oldPostMeta{}
	r.DB.Table("qsc_bbs_forum_thread").Where("tid in ?", ids).Find(&metas)
	for _, meta := range metas {
		permissions[meta.PostId] = meta.Permission
	}

	topics := make([]model.Topic, 0, len(posts))
	for _, post := range posts {
		topics = append(topics, r.buildTopic(post, counts[post.PostId], permissions[post.PostId]))
	}
	return topics
}

// GetTopicsByIds 按编号批量获取旧站主题，用于搜索引擎命中后回表。
func (r *oldBBSService) GetTopicsByIds(ids []int64) map[int64]model.Topic {
	if len(ids) == 0 {
		return nil
	}
	posts := []oldPost{}
	if r.DB.Table("qsc_bbs_forum_post").Where("tid in ?", ids).Where("first = 1").Find(&posts).Error != nil {
		return nil
	}
	result := make(map[int64]model.Topic, len(posts))
	for _, topic := range r.buildTopics(posts) {
		result[topic.Id] = topic
	}
	return result
}

// ScanTopics 按 tid 顺序分批扫描全部旧站主题，用于全量建索引。
func (r *oldBBSService) ScanTopics(batchSize int, callback func(topics []model.Topic) error) error {
	if batchSize <= 0 {
		batchSize = 500
	}
	var cursor int64 = 0
	for {
		posts := []oldPost{}
		if err := r.DB.Table("qsc_bbs_forum_post").
			Where("first = 1").
			Where("tid > ?", cursor).
			Order("tid asc").
			Limit(batchSize).
			Find(&posts).Error; err != nil {
			return err
		}
		if len(posts) == 0 {
			return nil
		}
		cursor = posts[len(posts)-1].PostId
		if err := callback(r.buildTopics(posts)); err != nil {
			return err
		}
	}
}

func (r *oldBBSService) GetTopic(id int64) *model.Topic {
	post := oldPost{}
	if r.DB.Table("qsc_bbs_forum_post").Where("tid = ?", id).Where("first = 1").Take(&post).Error != nil {
		return nil
	}
	topic := r.post2topic(post)
	return &topic
}

func (r *oldBBSService) GetCommentsPage(topicId int64, page int, ascOrder bool) (comments []model.Comment, paging *sqls.Paging) {
	if page < 1 {
		page = 1
	}
	limit := 10
	posts := []oldPost{}
	comments = []model.Comment{}
	optionalDesc := ""
	if !ascOrder {
		optionalDesc = " DESC"
	}

	var total int64
	r.DB.Table("qsc_bbs_forum_post").Where("tid = ?", topicId).Where("first = 0").Count(&total)

	if r.DB.Table("qsc_bbs_forum_post").Where("tid = ?", topicId).Order("position"+optionalDesc).Where("first = 0").Limit(limit).Offset((page-1)*limit).Find(&posts).Error != nil {
		comments = nil
		paging = &sqls.Paging{Page: page, Limit: limit, Total: total}
		return
	}
	for _, post := range posts {
		var cnt int64
		r.DB.Table("qsc_bbs_forum_postcomment").Where("pid = ?", post.MsgId).Count(&cnt)
		comments = append(comments, model.Comment{
			Model:        model.Model{Id: post.MsgId},
			UserId:       -1,
			EntityType:   "topic",
			EntityId:     topicId,
			Author:       post.Author,
			Content:      post.Content,
			CommentCount: cnt,
			CreateTime:   post.Timestamp,
			IsOldBBS:     true,
		})
	}
	paging = &sqls.Paging{Page: page, Limit: limit, Total: total}
	return
}

func (r *oldBBSService) GetComments(_ string, TopicId int64, cursor int64, ascOrder bool) (comments []model.Comment, nextCursor int64, hasMore bool) {
	limit := 20
	posts := []oldPost{}
	comments = []model.Comment{}
	optional_desc := ""
	if !ascOrder {
		optional_desc = " DESC"
	}
	if r.DB.Table("qsc_bbs_forum_post").Where("tid = ?", TopicId).Order("position"+optional_desc).Where("first = 0").Limit(limit).Offset(int(cursor)).Find(&posts).Error != nil {
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
			Author:       post.Author,
			Content:      post.Content,
			CommentCount: cnt,
			CreateTime:   post.Timestamp,
			IsOldBBS:     true,
		})
	}
	nextCursor = cursor + int64(len(comments))
	hasMore = len(comments) != 0
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
	hasMore = len(comments) != 0
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

	for forum.SuperId != 0 {
		r.DB.Table("qsc_bbs_forum_forum").Where("fid = ?", forum.SuperId).Take(&forum)
		name = forum.Name + "-" + name
	}

	r.forum_name_map[fid] = name
	return name
}

func (r *oldBBSService) GetTopicsByForum(fid int64, cursor int64) (topics []model.Topic, nextCursor int64, hasMore bool) {
	limit := 20
	posts := []oldPost{}
	topics = []model.Topic{}

	if r.DB.Table("qsc_bbs_forum_post").Where("fid = ?", fid).Order("dateline DESC").Where("first = 1").Limit(limit).Offset(int(cursor)).Find(&posts).Error != nil {
		topics = nil
		return
	}
	for _, post := range posts {
		topics = append(topics, r.post2topic(post))
	}
	nextCursor = cursor + int64(len(topics))
	hasMore = len(topics) != 0
	return
}

func (r *oldBBSService) GetTopicsByKeyword(keyword string, page int, pagesize int) (topics []model.Topic, total int64) {
	if page < 1 {
		page = 1
	}
	if pagesize <= 0 {
		pagesize = 20
	}
	pattern := "%" + escapeLike(keyword) + "%"
	r.DB.Table("qsc_bbs_forum_post").Where("subject like ?", pattern).Where("first = 1").Count(&total)
	return r.GetTopicsByKeywordOffset(keyword, (page-1)*pagesize, pagesize), total
}

// GetTopicsByKeywordOffset 按偏移量获取旧站搜索结果，用于回退路径的逐批扫描。
func (r *oldBBSService) GetTopicsByKeywordOffset(keyword string, offset int, limit int) []model.Topic {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 20
	}
	pattern := "%" + escapeLike(keyword) + "%"
	posts := []oldPost{}
	if r.DB.Table("qsc_bbs_forum_post").
		Where("subject like ?", pattern).
		Where("first = 1").
		Order("dateline DESC").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error != nil {
		return nil
	}
	return r.buildTopics(posts)
}
