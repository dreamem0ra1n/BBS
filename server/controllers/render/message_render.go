package render

import (
	"bbs-go/model"
	"bbs-go/model/constants"
	"bbs-go/pkg/bbsurls"
	"bbs-go/pkg/msg"
	"bbs-go/repositories"
	"strconv"

	"github.com/mlogclub/simple/sqls"
	"github.com/tidwall/gjson"
)

func BuildMessage(msg *model.Message) *model.MessageResponse {
	if msg == nil {
		return nil
	}

	from := BuildUserInfoDefaultIfNull(msg.FromId)
	if msg.FromId <= 0 {
		from.Nickname = "系统通知"
	}
	detailUrl := getMessageDetailUrl(msg)
	// logrus.Info("get URL : ", detailUrl)
	resp := &model.MessageResponse{
		MessageId:    msg.Id,
		From:         from,
		UserId:       msg.UserId,
		Title:        msg.Title,
		Content:      msg.Content,
		QuoteContent: msg.QuoteContent,
		Type:         msg.Type,
		DetailUrl:    detailUrl,
		ExtraData:    msg.ExtraData,
		Status:       msg.Status,
		CreateTime:   msg.CreateTime,
	}
	return resp
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
	if msgType == msg.TypeTopicComment || msgType == msg.TypeArticleComment || msgType == msg.TypeCommentReply {
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
