package services

import (
	"bbs-go/model"
	"bbs-go/model/constants"
	"bbs-go/pkg/mention"
	"bbs-go/pkg/msg"

	"github.com/mlogclub/simple/sqls"
)

func prepareMentions(content, format string, topic *model.Topic) (string, []int64) {
	users := map[string]int64{}
	return mention.Replace(content, format, func(nickname string) int64 {
		if id, found := users[nickname]; found {
			return id
		}
		users[nickname] = 0
		matches := UserService.Find(sqls.NewCnd().Eq("nickname", nickname).Eq("status", constants.StatusOk).Limit(2))
		if len(matches) != 1 || topic != nil && !model.UserCanAccessTopic(&matches[0], topic) && matches[0].Id != topic.UserId {
			return 0
		}
		users[nickname] = matches[0].Id
		return matches[0].Id
	})
}

func sendMentionMessages(from int64, userIds []int64, entityType string, entityId, commentId int64, title string) {
	for _, userId := range userIds {
		if userId == from {
			continue
		}
		user := UserService.Get(userId)
		if user == nil || user.Status != constants.StatusOk {
			continue
		}
		var extra interface{}
		if commentId > 0 {
			extra = &msg.CommentExtraData{EntityType: entityType, EntityId: entityId, CommentId: commentId}
		} else {
			extra = &msg.TopicLikeExtraData{TopicId: entityId}
		}
		MessageService.SendMsg(from, userId, msg.TypeMention, "提到了你", "", "《"+title+"》", extra)
	}
}

func mentionCommentTarget(comment *model.Comment) (string, int64, string, *model.Topic, bool) {
	entityType, entityId := comment.EntityType, comment.EntityId
	for depth := 0; entityType == constants.EntityComment && depth < 10; depth++ {
		parent := CommentService.Get(entityId)
		if parent == nil || parent.Status != constants.StatusOk {
			return "", 0, "", nil, false
		}
		entityType, entityId = parent.EntityType, parent.EntityId
	}
	switch entityType {
	case constants.EntityTopic:
		topic := TopicService.Get(entityId)
		if topic != nil && topic.Status == constants.StatusOk {
			return entityType, entityId, topic.GetTitle(), topic, true
		}
	case constants.EntityArticle:
		article := ArticleService.Get(entityId)
		if article != nil && article.Status == constants.StatusOk {
			return entityType, entityId, article.Title, nil, true
		}
	}
	return "", 0, "", nil, false
}
