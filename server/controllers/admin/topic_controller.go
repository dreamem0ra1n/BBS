package admin

import (
	"strconv"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"

	"bbs-go/controllers/render"
	"bbs-go/model/constants"
	"bbs-go/pkg/errs"
	"bbs-go/services"
)

type TopicController struct {
	Ctx iris.Context
}

func (c *TopicController) GetBy(id int64) *web.JsonResult {
	t := services.TopicService.Get(id)
	if t == nil {
		return web.JsonErrorMsg("Not found, id=" + strconv.FormatInt(id, 10))
	}
	return web.JsonData(t)
}

func (c *TopicController) AnyList() *web.JsonResult {
	list, paging := services.TopicService.FindPageByParams(params.NewQueryParams(c.Ctx).
		EqByReq("id").EqByReq("user_id").EqByReq("status").EqByReq("recommend").LikeByReq("title").PageByReq().Desc("id"))

	var results []map[string]interface{}
	for _, topic := range list {
		item := render.BuildSimpleTopic(services.UserTokenService.GetCurrent(c.Ctx), &topic)
		builder := web.NewRspBuilder(item)
		builder.Put("status", topic.Status)
		results = append(results, builder.Build())
	}

	return web.JsonData(&web.PageResult{Results: results, Page: paging})
}

// 推荐
func (c *TopicController) PostRecommend() *web.JsonResult {
	id, err := params.FormValueInt64(c.Ctx, "id")
	if err != nil {
		return web.JsonError(err)
	}
	err = services.TopicService.SetRecommend(id, true)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

// 取消推荐
func (c *TopicController) DeleteRecommend() *web.JsonResult {
	id, err := params.FormValueInt64(c.Ctx, "id")
	if err != nil {
		return web.JsonError(err)
	}
	err = services.TopicService.SetRecommend(id, false)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TopicController) PostDelete() *web.JsonResult {
	id, err := params.FormValueInt64(c.Ctx, "id")
	if err != nil {
		return web.JsonError(err)
	}
	user := services.UserTokenService.GetCurrent(c.Ctx)
	if user == nil {
		return web.JsonError(errs.NotLogin)
	}
	err = services.TopicService.Delete(id, user.Id, c.Ctx.Request())
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TopicController) PostUndelete() *web.JsonResult {
	id, err := params.FormValueInt64(c.Ctx, "id")
	if err != nil {
		return web.JsonError(err)
	}
	err = services.TopicService.Undelete(id)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

// 更新节点和标签
func (c *TopicController) PostUpdate() *web.JsonResult {
	topicId := params.FormValueInt64Default(c.Ctx, "id", 0)
	if topicId <= 0 {
		topicId = params.FormValueInt64Default(c.Ctx, "topicId", 0)
	}
	if topicId <= 0 {
		return web.JsonErrorMsg("id is required")
	}
	nodeId, err := params.FormValueInt64(c.Ctx, "nodeId")
	if err != nil {
		return web.JsonError(err)
	}
	tagIds := params.FormValueInt64Array(c.Ctx, "tagIds")
	user := services.UserTokenService.GetCurrent(c.Ctx)
	if user == nil || !user.IsAdminUserOrHigher() {
		return web.JsonErrorMsg("无权限")
	}
	if topic := services.TopicService.Get(topicId); topic == nil || topic.Status != constants.StatusOk {
		return web.JsonErrorMsg("话题不存在或已被删除")
	}
	if err := services.TopicService.UpdateNodeAndTags(topicId, user.Id, nodeId, tagIds); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
