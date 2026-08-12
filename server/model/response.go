package model

import (
	"bbs-go/model/constants"

	"github.com/mlogclub/simple/web"
)

// UserInfo 用户简单信息
type UserInfo struct {
	Id           int64  `json:"id"`
	Nickname     string `json:"nickname"`
	Realname     string `json:"realname" form:"realname"`     // 真实姓名
	Major        string `json:"major" form:"major"`           // 专业
	Birthday     string `json:"birthday" form:"birthday"`     // 生日
	Department   string `json:"department" form:"department"` // 部门
	Mobile       string `json:"mobile" form:"mobile"`         // 电话
	Wechat       string `json:"wechat" form:"wechat"`         // 微信号
	Qq           string `json:"qq" form:"qq"`                 // QQ号
	Avatar       string `json:"avatar"`                       // 头像
	SmallAvatar  string `json:"smallAvatar"`                  // 小头像
	TopicCount   int    `json:"topicCount"`                   // 话题数量
	CommentCount int    `json:"commentCount"`                 // 跟帖数量
	FansCount    int    `json:"fansCount"`                    // 粉丝数量
	FollowCount  int    `json:"followCount"`                  // 关注数量
	Score        int    `json:"score"`                        // 积分
	Description  string `json:"description"`
	CreateTime   int64  `json:"createTime"`

	Followed bool `json:"followed"`
}

// UserDetail 用户详细信息
type UserDetail struct {
	UserInfo
	Username             string `json:"username"`
	BackgroundImage      string `json:"backgroundImage"`
	SmallBackgroundImage string `json:"smallBackgroundImage"`
	HomePage             string `json:"homePage"`
	Forbidden            bool   `json:"forbidden"` // 是否禁言
	Status               int    `json:"status"`
	Realname             string `json:"realname"`   // 真实姓名
	Major                string `json:"major"`      // 专业
	Birthday             string `json:"birthday"`   // 生日
	Department           string `json:"department"` // 部门
	Mobile               string `json:"mobile"`     // 电话
	Wechat               string `json:"wechat"`     // 微信号
	Qq                   string `json:"qq"`         // QQ号
}

// UserProfile 用户个人信息
type UserProfile struct {
	UserDetail
	Roles           []string `json:"roles"`
	PasswordSet     bool     `json:"passwordSet"` // 密码已设置
	BackgroundImage string   `json:"smallBackgroundImage"`
	Email           string   `json:"email"`
	EmailVerified   bool     `json:"emailVerified"`
	Realname        string   `json:"realname" form:"realname"`     // 真实姓名
	Major           string   `json:"major" form:"major"`           // 专业
	Birthday        string   `json:"birthday" form:"birthday"`     // 生日
	Department      string   `json:"department" form:"department"` // 部门
	Mobile          string   `json:"mobile" form:"mobile"`         // 电话
	Wechat          string   `json:"wechat" form:"wechat"`         // 微信号
	Qq              string   `json:"qq" form:"qq"`                 // QQ号
}

type TagResponse struct {
	NodeId  int    `json:"nodeId"`
	TagId   int64  `json:"tagId"`
	TagName string `json:"tagName"`
}

type ArticleSimpleResponse struct {
	ArticleId  int64          `json:"articleId"`
	User       *UserInfo      `json:"user"`
	Tags       *[]TagResponse `json:"tags"`
	Title      string         `json:"title"`
	Summary    string         `json:"summary"`
	SourceUrl  string         `json:"sourceUrl"`
	ViewCount  int64          `json:"viewCount"`
	CreateTime int64          `json:"createTime"`
	Status     int            `json:"status"`
}

type ArticleResponse struct {
	ArticleSimpleResponse
	Content string `json:"content"`
}

type NodeResponse struct {
	NodeId      int64  `json:"nodeId"`
	Name        string `json:"name"`
	Logo        string `json:"logo"`
	Description string `json:"description"`
}

type SearchTopicResponse struct {
	TopicId    int64          `json:"topicId"`
	User       *UserInfo      `json:"user"`
	Node       *NodeResponse  `json:"node"`
	Tags       *[]TagResponse `json:"tags"`
	Title      string         `json:"title"`
	Summary    string         `json:"summary"`
	CreateTime int64          `json:"createTime"`
}

// 帖子列表返回实体
type TopicResponse struct {
	TopicId         int64               `json:"topicId"`
	Type            constants.TopicType `json:"type"`
	User            *UserInfo           `json:"user"`
	Node            *NodeResponse       `json:"node"`
	Tags            *[]TagResponse      `json:"tags"`
	AccessLv        int                 `json:"access_lv"`
	Title           string              `json:"title"`
	Summary         string              `json:"summary"`
	Content         string              `json:"content"`
	ImageList       []ImageInfo         `json:"imageList"`
	LastCommentTime int64               `json:"lastCommentTime"`
	ViewCount       int64               `json:"viewCount"`
	CommentCount    int64               `json:"commentCount"`
	LikeCount       int64               `json:"likeCount"`
	Liked           bool                `json:"liked"`
	CreateTime      int64               `json:"createTime"`
	LastEditUser    *UserInfo           `json:"lastEditUser,omitempty"`
	LastEditTime    int64               `json:"lastEditTime"`
	Recommend       bool                `json:"recommend"`
	RecommendTime   int64               `json:"recommendTime"`
	Sticky          bool                `json:"sticky"`
	StickyTime      int64               `json:"stickyTime"`
	IsOldBBS        bool                `gorm:"-" json:"isOldBBS" form:"-"`
}

// CommentResponse 评论返回数据
type CommentResponse struct {
	CommentId    int64             `json:"commentId"`
	User         *UserInfo         `json:"user"`
	EntityType   string            `json:"entityType"`
	EntityId     int64             `json:"entityId"`
	Content      string            `json:"content"`
	ImageList    []ImageInfo       `json:"imageList"`
	LikeCount    int64             `json:"likeCount"`
	CommentCount int64             `json:"commentCount"`
	Liked        bool              `json:"liked"`
	QuoteId      int64             `json:"quoteId"`
	Quote        *CommentResponse  `json:"quote"`
	Replies      *web.CursorResult `json:"replies"`
	Status       int               `json:"status"`
	CreateTime   int64             `json:"createTime"`
	RawContent   string            `json:"rawContent,omitempty"`
	LastEditUser *UserInfo         `json:"lastEditUser,omitempty"`
	LastEditTime int64             `json:"lastEditTime"`
	CanEdit      bool              `json:"canEdit"`
	CanDelete    bool              `json:"canDelete"`
	IsOldBBS     bool              `gorm:"-" json:"isOldBBS" form:"-"`
}

// 收藏返回数据
type FavoriteResponse struct {
	FavoriteId int64     `json:"favoriteId"`
	EntityType string    `json:"entityType"`
	EntityId   int64     `json:"entityId"`
	Deleted    bool      `json:"deleted"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	User       *UserInfo `json:"user"`
	Url        string    `json:"url"`
	CreateTime int64     `json:"createTime"`
}

// 消息
type MessageResponse struct {
	MessageId    int64     `json:"messageId"`
	From         *UserInfo `json:"from"`    // 消息发送人
	UserId       int64     `json:"userId"`  // 消息接收人编号
	Title        string    `json:"title"`   // 标题
	Content      string    `json:"content"` // 消息内容
	QuoteContent string    `json:"quoteContent"`
	Type         int       `json:"type"`
	DetailUrl    string    `json:"detailUrl"` // 消息详情url
	ExtraData    string    `json:"extraData"`
	Status       int       `json:"status"`
	CreateTime   int64     `json:"createTime"`
}

// 图片
type ImageInfo struct {
	Url     string `json:"url"`
	Preview string `json:"preview"`
}
