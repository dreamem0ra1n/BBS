package eventhandler

import (
	"reflect"

	"bbs-go/model/constants"
	"bbs-go/pkg/event"
	"bbs-go/services"
)

// 主题索引同步：新建、更新、推荐状态变化时刷新文档，删除时移除文档。
// 事件由 event 包异步派发，不阻塞发帖/编辑请求。
func init() {
	event.RegHandler(reflect.TypeOf(event.TopicCreateEvent{}), handleTopicIndexRefresh)
	event.RegHandler(reflect.TypeOf(event.TopicUpdateEvent{}), handleTopicIndexRefresh)
	event.RegHandler(reflect.TypeOf(event.TopicRecommendEvent{}), handleTopicIndexRefresh)
	event.RegHandler(reflect.TypeOf(event.TopicDeleteEvent{}), handleTopicIndexDelete)
}

func handleTopicIndexRefresh(i interface{}) {
	topicId := topicIdOf(i)
	if topicId <= 0 {
		return
	}
	topic := services.TopicService.Get(topicId)
	if topic == nil {
		return
	}
	if topic.Status != constants.StatusOk {
		services.SearchService.DeleteTopicIndex(topicId)
		return
	}
	services.SearchService.IndexTopic(topic)
}

func handleTopicIndexDelete(i interface{}) {
	e, ok := i.(event.TopicDeleteEvent)
	if !ok {
		return
	}
	services.SearchService.DeleteTopicIndex(e.TopicId)
}

func topicIdOf(i interface{}) int64 {
	switch e := i.(type) {
	case event.TopicCreateEvent:
		return e.TopicId
	case event.TopicUpdateEvent:
		return e.TopicId
	case event.TopicRecommendEvent:
		return e.TopicId
	default:
		return 0
	}
}
