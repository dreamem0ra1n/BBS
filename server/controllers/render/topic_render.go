package render

import (
	"bbs-go/model"
	"bbs-go/model/constants"
	"bbs-go/pkg/markdown"
	"bbs-go/services"
	"html"

	"github.com/mlogclub/simple/common/arrays"
	"github.com/mlogclub/simple/common/strs"
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
	rsp.AccessLv = topic.AccessLv

	// 构建内容
	if !topic.IsOldBBS {
		if model.UserCanAccessTopic(user, topic) || (user != nil && user.Id == topic.UserId) {
			if buildContent {
				if topic.Type == constants.TopicTypeTopic {
					content := markdown.ToHTML(topic.Content)
					rsp.Content = handleHtmlContent(content)
				} else {
					rsp.Content = html.EscapeString(topic.Content)
				}
			} else {
				rsp.Summary = markdown.GetSummary(topic.Content, 128)
			}
		} else {
			if buildContent {
				rsp.Content = " 🚫 抱歉，您无权访问该帖子的内容！"
			} else {
				rsp.Title = " 🚫 抱歉，您无权访问该帖！"
			}
		}

		if topic.Type == constants.TopicTypeTweet {
			if strs.IsBlank(topic.Content) {
				rsp.Content = "分享图片"
			} else {
				rsp.Content = html.EscapeString(topic.Content)
			}
		}

		if topic.NodeId > 0 {
			node := services.TopicNodeService.Get(topic.NodeId)
			rsp.Node = BuildNode(node)
		}

		tags := services.TopicService.GetTopicTags(topic.Id)
		rsp.Tags = BuildTags(tags)
	} else {
		rsp.Content = topic.Content

		if topic.NodeId > 0 {
			node := services.TopicNodeService.Get(topic.NodeId)
			rsp.Node = BuildNode(node)
		}
		rsp.Tags = &[]model.TagResponse{{
			TagId:   -1,
			TagName: topic.Forum,
		}}
	}
	return rsp
}
