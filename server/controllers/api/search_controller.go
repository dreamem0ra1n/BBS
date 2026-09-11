package api

import (
	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"

	"bbs-go/controllers/render"
	"bbs-go/services"
)

type SearchController struct {
	Ctx iris.Context
}

// PostTopic 站内搜索。
// 过滤删除态、page/limit 上限、权限过滤都在 SearchService 内统一处理。
func (c *SearchController) PostTopic() *web.JsonResult {
	user := services.UserTokenService.GetCurrent(c.Ctx)
	topics, paging := services.SearchService.SearchTopics(services.SearchParams{
		Keyword:  params.FormValue(c.Ctx, "keyword"),
		NodeId:   params.FormValueInt64Default(c.Ctx, "nodeId", 0),
		Page:     params.FormValueIntDefault(c.Ctx, "page", 1),
		Limit:    params.FormValueIntDefault(c.Ctx, "limit", 20),
		IsOldBBS: false,
		User:     user,
	})
	return web.JsonPageData(render.BuildSimpleTopics(topics, user), paging)
}

// PostOldbbs 旧 BBS 搜索。
// 旧站按 readperm（access_lv）控制可见性，权限判断与帖子详情页保持一致。
func (c *SearchController) PostOldbbs() *web.JsonResult {
	user := services.UserTokenService.GetCurrent(c.Ctx)
	topics, paging := services.SearchService.SearchTopics(services.SearchParams{
		Keyword:  params.FormValue(c.Ctx, "keyword"),
		Page:     params.FormValueIntDefault(c.Ctx, "page", 1),
		Limit:    params.FormValueIntDefault(c.Ctx, "limit", 20),
		IsOldBBS: true,
		User:     user,
	})
	return web.JsonPageData(render.BuildSimpleTopics(topics, user), paging)
}
