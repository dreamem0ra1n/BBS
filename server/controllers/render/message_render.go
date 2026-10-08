package render

import (
	"bbs-go/model"
	"bbs-go/model/constants"
	"bbs-go/pkg/bbsurls"
	"bbs-go/pkg/msg"
	"bbs-go/repositories"
	"bbs-go/services"
	"strconv"

	"github.com/mlogclub/simple/sqls"
	"github.com/tidwall/gjson"
)

func BuildMessage(message *model.Message) *model.MessageResponse {
	if message == nil {
		return nil
	}

	from := BuildUserInfoDefaultIfNull(message.FromId)
	if message.FromId <= 0 {
		from.Nickname = "系统通知"
	}
	detailUrl := getMessageDetailUrl(message)
	// logrus.Info("get URL : ", detailUrl)
	resp := &model.MessageResponse{
		MessageId:              message.Id,
		From:                   from,
		UserId:                 message.UserId,
		Title:                  message.Title,
		Content:                message.Content,
		QuoteContent:           message.QuoteContent,
		Type:                   message.Type,
		DetailUrl:              detailUrl,
		ExtraData:              message.ExtraData,
		BirthdayBlessingAuthor: getBirthdayBlessingAuthor(message),
		Status:                 message.Status,
		CreateTime:             message.CreateTime,
	}
	if msg.Type(message.Type) == msg.TypeMention {
		if resp.QuoteContent != "" {
			resp.Content = ""
		}
		if !canViewMentionMessage(message) {
			resp.Content = "内容不可查看"
			resp.QuoteContent = ""
			resp.DetailUrl = ""
		}
	}
	return resp
}

func canViewMentionMessage(message *model.Message) bool {
	user := services.UserService.Get(message.UserId)
	if user == nil || user.Status != constants.StatusOk {
		return false
	}
	commentId := gjson.Get(message.ExtraData, "commentId").Int()
	if commentId > 0 {
		comment := services.CommentService.Get(commentId)
		if comment == nil || comment.Status != constants.StatusOk {
			return false
		}
		for depth := 0; comment.EntityType == constants.EntityComment && depth < 10; depth++ {
			comment = services.CommentService.Get(comment.EntityId)
			if comment == nil || comment.Status != constants.StatusOk {
				return false
			}
		}
		if comment.EntityType == constants.EntityComment {
			return false
		}
	}
	entityType := gjson.Get(message.ExtraData, "entityType").String()
	entityId := gjson.Get(message.ExtraData, "entityId").Int()
	if commentId == 0 {
		entityType = constants.EntityTopic
		entityId = gjson.Get(message.ExtraData, "topicId").Int()
	}
	switch entityType {
	case constants.EntityTopic:
		topic := services.TopicService.Get(entityId)
		return topic != nil && topic.Status == constants.StatusOk && (model.UserCanAccessTopic(user, topic) || topic.UserId == user.Id)
	case constants.EntityArticle:
		article := services.ArticleService.Get(entityId)
		return article != nil && article.Status == constants.StatusOk
	}
	return false
}

func getBirthdayBlessingAuthor(message *model.Message) *model.UserInfo {
	if msg.Type(message.Type) != msg.TypeBirthday {
		return nil
	}
	authorId := gjson.Get(message.ExtraData, "blessingAuthorId").Int()
	if authorId > 0 {
		return BuildUserInfoDefaultIfNull(authorId)
	}
	blessingId := gjson.Get(message.ExtraData, "blessingId").Int()
	if blessingId <= 0 {
		return nil
	}
	blessing := repositories.BirthdayBlessingRepository.Get(sqls.DB(), blessingId)
	if blessing == nil {
		return nil
	}
	author := repositories.UserRepository.FindOne(sqls.DB(), sqls.NewCnd().
		Eq("nickname", blessing.Nickname).
		Eq("status", constants.StatusOk).
		Asc("id"))
	return BuildUserInfo(author)
}

// BuildMessages 渲染消息列表
func BuildMessages(messages []model.Message) []model.MessageResponse {
	if len(messages) == 0 {
		return nil
	}
	var responses []model.MessageResponse
	for _, message := range messages {
		responses = append(responses, *BuildMessage(&message))
	}
	return responses
}

// getMessageDetailUrl 查看消息详情链接地址
func getMessageDetailUrl(t *model.Message) string {
	msgType := msg.Type(t.Type)
	if msgType == msg.TypeBirthday {
		return ""
	}
	// logrus.Info("debug: ", msgType)
	if msgType == msg.TypeTopicComment || msgType == msg.TypeArticleComment || msgType == msg.TypeCommentReply || msgType == msg.TypeMention && gjson.Get(t.ExtraData, "commentId").Int() > 0 {
		entityType := gjson.Get(t.ExtraData, "entityType")
		entityId := gjson.Get(t.ExtraData, "entityId")
		commentId := gjson.Get(t.ExtraData, "commentId").Int()
		// logrus.Info("debug: ", entityId, entityType.String())
		if entityType.String() == constants.EntityArticle {
			return appendCommentAnchor(bbsurls.ArticleUrl(entityId.Int()), commentId)
		} else if entityType.String() == constants.EntityTopic {
			return appendCommentAnchor(bbsurls.TopicUrl(entityId.Int()), commentId)
		} else if entityType.String() == constants.EntityComment {
			return getCommentDetailUrl(entityId.Int(), commentId)
		}
	} else if msgType == msg.TypeTopicLike ||
		msgType == msg.TypeTopicFavorite ||
		msgType == msg.TypeTopicRecommend ||
		msgType == msg.TypeTopicGift {
		topicId := gjson.Get(t.ExtraData, "topicId")
		if topicId.Exists() && topicId.Int() > 0 {
			return bbsurls.TopicUrl(topicId.Int())
		}
	} else if msgType == msg.TypeMention {
		topicId := gjson.Get(t.ExtraData, "topicId").Int()
		if topicId > 0 {
			return bbsurls.TopicUrl(topicId)
		}
	}
	return bbsurls.AbsUrl("/user/messages")
}

func appendCommentAnchor(detailUrl string, commentId int64) string {
	if commentId <= 0 {
		return detailUrl
	}
	return detailUrl + "#comment-" + strconv.FormatInt(commentId, 10)
}

func getCommentDetailUrl(commentId, targetCommentId int64) string {
	for depth := 0; depth < 10 && commentId > 0; depth++ {
		commentResults := repositories.CommentRepository.FindBySql(sqls.DB(),
			"SELECT * FROM t_comment WHERE id = ?",
			commentId,
		)
		if len(commentResults) == 0 {
			return bbsurls.AbsUrl("/user/messages")
		}

		comment := commentResults[0]
		switch comment.EntityType {
		case constants.EntityArticle:
			return appendCommentAnchor(bbsurls.ArticleUrl(comment.EntityId), targetCommentId)
		case constants.EntityTopic:
			return appendCommentAnchor(bbsurls.TopicUrl(comment.EntityId), targetCommentId)
		case constants.EntityComment:
			commentId = comment.EntityId
		default:
			return bbsurls.AbsUrl("/user/messages")
		}
	}
	return bbsurls.AbsUrl("/user/messages")
}

// func getFatherMsg(t *model.Comment) *model.Comment {
// commentResults := repositories.CommentRepository.FindBySql(sqls.DB(),
// 	"SELECT * FROM t_comment WHERE id = ?",
// 	t.EntityId,
// )
// 	message := messagesResults[0]
// 	if ftype == constants.EntityTopic {
// 		return &message
// 	} else {
// 		return getFatherMsg(&message)
// 	}
// }
