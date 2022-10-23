package render

import (
	"bbs-go/model"
	"bbs-go/model/constants"
	"bbs-go/services"

	"github.com/mlogclub/simple/common/arrays"
)

func BuildTopic(user *model.User, topic *model.Topic) *model.TopicResponse {
	return _buildTopic(user, topic, true)
}

func BuildSimpleTopic(user *model.User, topic *model.Topic) *model.TopicResponse {
	buildContent := topic.Type == constants.TopicTypeTweet // 动态时渲染内容
	return _buildTopic(user, topic, buildContent)
}

func BuildSimpleTopics(topics []model.Topic, currentUser *model.User) []model.TopicResponse {
	if len(topics) == 0 {
		return nil
	}
	var likedTopicIds []int64
	if currentUser != nil {
		var topicIds []int64
		for _, topic := range topics {
			topicIds = append(topicIds, topic.Id)
		}
		likedTopicIds = services.UserLikeService.IsLiked(currentUser.Id, constants.EntityTopic, topicIds)
	}

	var responses []model.TopicResponse
	for _, topic := range topics {
		item := BuildSimpleTopic(currentUser, &topic)
		item.Liked = arrays.Contains(topic.Id, likedTopicIds)
		responses = append(responses, *item)
	}
	return responses
}

func _buildTopic(user *model.User, topic *model.Topic, buildContent bool) *model.TopicResponse {
	if topic == nil {
		return nil
	}
	rsp := &model.TopicResponse{}

	rsp.TopicId = topic.Id
	rsp.Type = topic.Type
	rsp.Title = topic.Title
	rsp.User = BuildUserInfoDefaultIfNull(topic.UserId)
	rsp.LastCommentTime = topic.LastCommentTime
	rsp.CreateTime = topic.CreateTime
	rsp.ViewCount = topic.ViewCount
	rsp.CommentCount = topic.CommentCount
	rsp.LikeCount = topic.LikeCount
	rsp.Recommend = topic.Recommend
	rsp.RecommendTime = topic.RecommendTime
	rsp.Sticky = topic.Sticky
	rsp.StickyTime = topic.StickyTime

	// 构建内容
	if !(model.UserCanAccessTopic(user, topic) || (user != nil && user.Id == topic.UserId)) {
		rsp.Title = "[ 🚫 权限不足 🚫 ]" + rsp.Title
	}

	if topic.NodeId > 0 {
		node := services.TopicNodeService.Get(topic.NodeId)
		rsp.Node = BuildNode(node)
	}

	tags := services.TopicService.GetTopicTags(topic.Id)
	rsp.Tags = BuildTags(tags)

	return rsp
}
