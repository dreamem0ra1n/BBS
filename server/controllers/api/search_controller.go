package api

import (
	"bbs-go/controllers/render"
	"bbs-go/model"
	"bbs-go/pkg/es"
	"bbs-go/repositories"
	"bbs-go/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type SearchController struct {
	Ctx iris.Context
}

func (c *SearchController) AnyReindex() *web.JsonResult {
	go services.TopicService.ScanDesc(func(topics []model.Topic) {
		for _, t := range topics {
			topic := services.TopicService.Get(t.Id)
			es.UpdateTopicIndex(topic)
		}
	})
	return web.JsonSuccess()
}

func (c *SearchController) PostTopics() *web.JsonResult {
	var (
		page      = params.FormValueIntDefault(c.Ctx, "page", 1)
		keyword   = params.FormValue(c.Ctx, "keyword")
		nodeId    = params.FormValueInt64Default(c.Ctx, "nodeId", 0)
		timeRange = params.FormValueIntDefault(c.Ctx, "timeRange", 0)
	)

	docs, paging, err := es.SearchTopic(keyword, nodeId, timeRange, page, 20)
	if err != nil {
		return web.JsonError(err)
	}

	items := render.BuildSearchTopics(docs)
	return web.JsonPageData(items, paging)
}

func (c *SearchController) PostTopic() *web.JsonResult {
	var (
		page    = params.FormValueIntDefault(c.Ctx, "page", 1)
		keyword = params.FormValue(c.Ctx, "keyword")
		nodeId  = params.FormValueInt64Default(c.Ctx, "nodeId", 0)
		limit   = params.FormValueIntDefault(c.Ctx, "limit", 20)
	)
	offset := (page - 1) * limit

	var searchResults []model.Topic
	var totResult int64
	if nodeId == 0 {
		searchResults = repositories.TopicRepository.FindBySql(sqls.DB(),
			"SELECT * FROM t_topic WHERE ( title LIKE CONCAT('%',?,'%') ) LIMIT ?, ?;",
			keyword,
			offset,
			limit,
		)
		totResult = repositories.TopicRepository.CountBySql(sqls.DB(),
			"SELECT COUNT(*) FROM t_topic WHERE ( title LIKE CONCAT('%',?,'%') );",
			keyword,
		)
	} else if nodeId == -1 {
		searchResults = repositories.TopicRepository.FindBySql(sqls.DB(),
			"SELECT * FROM t_topic WHERE ( title LIKE CONCAT('%',?,'%') ) AND ( recommend = 1 ) LIMIT ?, ?;",
			keyword,
			offset,
			limit,
		)
		totResult = repositories.TopicRepository.CountBySql(sqls.DB(),
			"SELECT COUNT(*) FROM t_topic WHERE ( title LIKE CONCAT('%',?,'%') ) AND ( recommend = 1 );",
			keyword,
		)
	} else {
		searchResults = repositories.TopicRepository.FindBySql(sqls.DB(),
			"SELECT * FROM t_topic WHERE ( title LIKE CONCAT('%',?,'%') ) AND ( node_id = ? ) LIMIT ?, ?;",
			keyword,
			nodeId,
			offset,
			limit,
		)
		totResult = repositories.TopicRepository.CountBySql(sqls.DB(),
			"SELECT COUNT(*) FROM t_topic WHERE ( title LIKE CONCAT('%',?,'%') ) AND ( node_id = ? );",
			keyword,
			nodeId,
		)
	}

	paging := &sqls.Paging{
		Page:  page,
		Limit: limit,
		Total: totResult,
	}

	return web.JsonPageData(searchResults, paging)
}
