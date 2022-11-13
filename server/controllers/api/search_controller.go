package api

import (
	"bbs-go/model"
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
			"SELECT * FROM t_topic WHERE ( title LIKE CONCAT('%',?,'%') ) ORDER BY create_time DESC LIMIT ?, ?;",
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
			"SELECT * FROM t_topic WHERE ( title LIKE CONCAT('%',?,'%') ) AND ( recommend = 1 ) ORDER BY create_time DESC LIMIT ?, ?;",
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
			"SELECT * FROM t_topic WHERE ( title LIKE CONCAT('%',?,'%') ) AND ( node_id = ? ) ORDER BY create_time DESC LIMIT ?, ?;",
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

func (c *SearchController) PostOldbbs() *web.JsonResult {

	var (
		page    = params.FormValueIntDefault(c.Ctx, "page", 1)
		keyword = params.FormValue(c.Ctx, "keyword")
		limit   = params.FormValueIntDefault(c.Ctx, "limit", 20)
	)

	var searchResults []model.Topic
	var totResult int64

	searchResults, totResult = services.OldBBSService.GetTopicsByKeyword(keyword, page, limit)

	paging := &sqls.Paging{
		Page:  page,
		Limit: limit,
		Total: totResult,
	}

	return web.JsonPageData(searchResults, paging)
}
