package api

import (
	"bbs-go/model/constants"
	"bbs-go/pkg/errs"
	"bbs-go/pkg/markdown"
	"bbs-go/spam"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"

	"bbs-go/cache"
	"bbs-go/controllers/render"
	"bbs-go/model"
	"bbs-go/services"
)

type TopicController struct {
	Ctx iris.Context
}

// 节点
func (c *TopicController) GetNodes() *web.JsonResult {
	nodes := services.TopicNodeService.GetNodes()
	return web.JsonData(render.BuildNodes(nodes))
}

// 节点信息
func (c *TopicController) GetNode() *web.JsonResult {
	nodeId := params.FormValueInt64Default(c.Ctx, "nodeId", 0)
	node := services.TopicNodeService.Get(nodeId)
	return web.JsonData(render.BuildNode(node))
}

// 发表帖子
func (c *TopicController) PostCreate() *web.JsonResult {
	user := services.UserTokenService.GetCurrent(c.Ctx)
	if err := services.UserService.CheckPostStatus(user); err != nil {
		return web.JsonError(err)
	}
	form := model.GetCreateTopicForm(c.Ctx)

	if err := spam.CheckTopic(user, form); err != nil {
		return web.JsonError(err)
	}

	topic, err := services.TopicService.Publish(user.Id, form)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(render.BuildSimpleTopic(user, topic))
}

// 编辑时获取详情
func (c *TopicController) GetEditBy(topicId int64) *web.JsonResult {
	user := services.UserTokenService.GetCurrent(c.Ctx)
	if err := services.UserService.CheckPostStatus(user); err != nil {
		return web.JsonError(err)
	}

	topic := services.TopicService.Get(topicId)
	if topic == nil || topic.Status != constants.StatusOk {
		return web.JsonErrorMsg("话题不存在或已被删除")
	}
	if topic.Type != constants.TopicTypeTopic {
		return web.JsonErrorMsg("当前类型帖子不支持修改")
	}

	// 不是作者且没有管理权限
	if topic.UserId != user.Id && !user.CanManageTopic(topic) {
		return web.JsonErrorMsg("无权限")
	}

	tags := services.TopicService.GetTopicTags(topicId)
	var tagNames []string
	if len(tags) > 0 {
		for _, tag := range tags {
			tagNames = append(tagNames, tag.Name)
		}
	}

	return web.NewEmptyRspBuilder().
		Put("topicId", topic.Id).
		Put("nodeId", topic.NodeId).
		Put("title", topic.Title).
		Put("content", topic.Content).
		Put("hideContent", topic.HideContent).
		Put("tags", tagNames).
		Put("access_lv", topic.AccessLv).
		JsonResult()
}

// 编辑帖子
func (c *TopicController) PostEditBy(topicId int64) *web.JsonResult {
	user := services.UserTokenService.GetCurrent(c.Ctx)
	if err := services.UserService.CheckPostStatus(user); err != nil {
		return web.JsonError(err)
	}

	topic := services.TopicService.Get(topicId)
	if topic == nil || topic.Status != constants.StatusOk {
		return web.JsonErrorMsg("话题不存在或已被删除")
	}

	// 不是作者且没有管理权限
	if topic.UserId != user.Id && !user.CanManageTopic(topic) {
		return web.JsonErrorMsg("无权限")
	}

	var (
		nodeId      = params.FormValueInt64Default(c.Ctx, "nodeId", 0)
		title       = strings.TrimSpace(params.FormValue(c.Ctx, "title"))
		content     = strings.TrimSpace(params.FormValue(c.Ctx, "content"))
		hideContent = strings.TrimSpace(params.FormValue(c.Ctx, "hideContent"))
		tags        = params.FormValueStringArray(c.Ctx, "tags")
		accessLv    = params.FormValueIntDefault(c.Ctx, "access_lv", -1)
	)

	if accessLv < -1 {
		return web.JsonError(errors.New("invalid access level"))
	}
	err := services.TopicService.Edit(topicId, user.Id, nodeId, tags, title, content, hideContent, accessLv)
	if err != nil {
		return web.JsonError(err)
	}
	// 操作日志
	services.OperateLogService.AddOperateLog(user.Id, constants.OpTypeUpdate, constants.EntityTopic, topicId,
		"", c.Ctx.Request())
	return web.JsonData(render.BuildSimpleTopic(user, topic))
}

// 删除帖子
func (c *TopicController) PostDeleteBy(topicId int64) *web.JsonResult {
	user := services.UserTokenService.GetCurrent(c.Ctx)
	if err := services.UserService.CheckPostStatus(user); err != nil {
		return web.JsonError(err)
	}

	topic := services.TopicService.Get(topicId)
	if topic == nil || topic.Status != constants.StatusOk {
		return web.JsonSuccess()
	}

	// 作者不可以删除自己的帖子
	if !user.CanManageTopic(topic) {
		return web.JsonErrorMsg("无权限")
	}

	if err := services.TopicService.Delete(topicId, user.Id, c.Ctx.Request()); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

// PostRecommendBy 设为推荐
func (c *TopicController) PostRecommendBy(topicId int64) *web.JsonResult {
	topic := services.TopicService.Get(topicId)
	recommend, err := params.FormValueBool(c.Ctx, "recommend")
	if err != nil {
		return web.JsonError(err)
	}
	user := services.UserTokenService.GetCurrent(c.Ctx)
	if user == nil {
		return web.JsonError(errs.NotLogin)
	}
	// 没有管理权限
	if !user.CanManageTopic(topic) {
		return web.JsonErrorMsg("无权限")
	}

	err = services.TopicService.SetRecommend(topicId, recommend)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

// 帖子详情
// 旧BBS带前缀OLD
func (c *TopicController) GetBy(topicIdStr string) *web.JsonResult {
	topicId, isOld := parseIdStr(topicIdStr)
	user := services.UserTokenService.GetCurrent(c.Ctx)
	var topic *model.Topic
	if !isOld {
		topic = services.TopicService.Get(topicId)
	} else {
		topic = services.OldBBSService.GetTopic(topicId)
	}
	if topic == nil || topic.Status != constants.StatusOk {
		return web.JsonErrorMsg("主题不存在")
	}

	if model.UserCanAccessTopic(user, topic) || (user != nil && topic.UserId == user.Id) {
		if !isOld {
			services.TopicService.IncrViewCount(topicId) // 增加浏览量
		}
		return web.JsonData(render.BuildTopic(user, topic))
	} else {
		return web.JsonErrorMsg("无权限")
	}
}

// 点赞
func (c *TopicController) PostLikeBy(topicId int64) *web.JsonResult {
	user := services.UserTokenService.GetCurrent(c.Ctx)
	if user == nil {
		return web.JsonError(errs.NotLogin)
	}
	err := services.UserLikeService.TopicLike(user.Id, topicId)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

// 赠米
func (c *TopicController) PostGiftBy(topicId int64) *web.JsonResult {
	user := services.UserTokenService.GetCurrent(c.Ctx)
	if user == nil {
		return web.JsonError(errs.NotLogin)
	}
	score := params.FormValueIntDefault(c.Ctx, "score", 0)
	reason := params.FormValue(c.Ctx, "reason")
	gift, err := services.TopicGiftService.Gift(user.Id, topicId, score, reason)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(render.BuildTopicGift(gift))
}

// 点赞用户
func (c *TopicController) GetRecentlikesBy(topicId int64) *web.JsonResult {
	likes := services.UserLikeService.Recent(constants.EntityTopic, topicId, 5)
	var users []model.UserInfo
	for _, like := range likes {
		userInfo := render.BuildUserInfoDefaultIfNull(like.UserId)
		if userInfo != nil {
			users = append(users, *userInfo)
		}
	}
	return web.JsonData(users)
}

// 最新帖子
func (c *TopicController) GetRecent() *web.JsonResult {
	user := services.UserTokenService.GetCurrent(c.Ctx)
	topics := services.TopicService.Find(sqls.NewCnd().Where("status = ?", constants.StatusOk).Desc("id").Limit(10))
	return web.JsonData(render.BuildSimpleTopics(topics, user))
}

// 用户帖子列表
func (c *TopicController) GetUserTopics() *web.JsonResult {
	userId, err := params.FormValueInt64(c.Ctx, "userId")
	if err != nil {
		return web.JsonError(err)
	}
	cursor := params.FormValueInt64Default(c.Ctx, "cursor", 0)
	user := services.UserTokenService.GetCurrent(c.Ctx)
	if page := params.FormValueIntDefault(c.Ctx, "page", 0); page > 0 {
		topics, paging := services.TopicService.FindPageByCnd(sqls.NewCnd().
			Eq("user_id", userId).Eq("status", constants.StatusOk).Page(page, 20).Desc("id"))
		return web.JsonPageData(render.BuildSimpleTopics(topics, user), paging)
	}
	topics, cursor, hasMore := services.TopicService.GetUserTopics(userId, cursor)
	return web.JsonCursorData(render.BuildSimpleTopics(topics, user), strconv.FormatInt(cursor, 10), hasMore)
}

// 帖子列表
func (c *TopicController) GetTopics() *web.JsonResult {
	var (
		cursor       = params.FormValueInt64Default(c.Ctx, "cursor", 0)
		nodeId       = params.FormValueInt64Default(c.Ctx, "nodeId", 0)
		recommend, _ = params.FormValueBool(c.Ctx, "recommend")
		sort          = params.FormValue(c.Ctx, "sort")
		user         = services.UserTokenService.GetCurrent(c.Ctx)
	)
	topics, cursor, hasMore := services.TopicService.GetTopics(nodeId, cursor, recommend, sort)
	return web.JsonCursorData(render.BuildSimpleTopics(topics, user), strconv.FormatInt(cursor, 10), hasMore)
}

// 根据 nodeId 和 tag 获取帖子列表
func (c *TopicController) PostTopicsnt() *web.JsonResult {
	var (
		cursor       = params.FormValueInt64Default(c.Ctx, "cursor", 0)
		nodeId, err1 = params.FormValueInt64(c.Ctx, "nodeId")
		tagId, err2  = params.FormValueInt64(c.Ctx, "tagId")
		sort          = params.FormValue(c.Ctx, "sort")
		user         = services.UserTokenService.GetCurrent(c.Ctx)
	)
	if err1 != nil {
		return web.JsonError(err1)
	}
	if err2 != nil {
		return web.JsonError(err2)
	}
	topics, cursor, hasMore := services.TopicService.GetTopicsByNodeIdAndTag(tagId, nodeId, cursor, sort)
	return web.JsonCursorData(render.BuildSimpleTopics(topics, user), strconv.FormatInt(cursor, 10), hasMore)
}

// 标签帖子列表
func (c *TopicController) GetTagTopics() *web.JsonResult {
	var (
		cursor   = params.FormValueInt64Default(c.Ctx, "cursor", 0)
		tagIdStr = params.FormValue(c.Ctx, "tagId")
		user     = services.UserTokenService.GetCurrent(c.Ctx)
	)
	tagId, isOld := parseIdStr(tagIdStr)
	if tagId == -1 {
		return web.JsonError(fmt.Errorf("bad tag id"))
	}
	var (
		topics  []model.Topic
		hasMore bool
	)
	if !isOld {
		topics, cursor, hasMore = services.TopicService.GetTagTopics(tagId, cursor)
	} else {
		topics, cursor, hasMore = services.OldBBSService.GetTopicsByForum(tagId, cursor)
	}
	return web.JsonCursorData(render.BuildSimpleTopics(topics, user), strconv.FormatInt(cursor, 10), hasMore)
}

// 收藏
func (c *TopicController) GetFavoriteBy(topicId int64) *web.JsonResult {
	user := services.UserTokenService.GetCurrent(c.Ctx)
	if user == nil {
		return web.JsonError(errs.NotLogin)
	}
	err := services.FavoriteService.AddTopicFavorite(user.Id, topicId)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

// 推荐话题列表（目前逻辑为取最近50条数据随机展示）
func (c *TopicController) GetRecommend() *web.JsonResult {
	topics := cache.TopicCache.GetRecommendTopics()
	if len(topics) == 0 {
		return web.JsonSuccess()
	} else {
		dest := make([]model.Topic, len(topics))
		perm := rand.Perm(len(topics))
		for i, v := range perm {
			dest[v] = topics[i]
		}
		end := 10
		if end > len(topics) {
			end = len(topics)
		}
		ret := dest[0:end]
		return web.JsonData(render.BuildSimpleTopics(ret, nil))
	}
}

// 最新话题
func (c *TopicController) GetNewest() *web.JsonResult {
	topics := services.TopicService.Find(sqls.NewCnd().Eq("status", constants.StatusOk).Desc("id").Limit(6))
	return web.JsonData(render.BuildSimpleTopics(topics, nil))
}

func (c *TopicController) GetSticky_topics() *web.JsonResult {
	user := services.UserTokenService.GetCurrent(c.Ctx)
	nodeId := params.FormValueInt64Default(c.Ctx, "nodeId", 0)
	topics := services.TopicService.GetStickyTopics(nodeId, 3)
	return web.JsonData(render.BuildSimpleTopics(topics, user))
}

// 设置指定
func (c *TopicController) PostStickyBy(topicId int64) *web.JsonResult {
	user := services.UserTokenService.GetCurrent(c.Ctx)
	if user == nil {
		return web.JsonError(errs.NotLogin)
	}
	if !user.IsAdminUserOrHigher() {
		return web.JsonErrorMsg("无权限")
	}

	var (
		sticky = params.FormValueBoolDefault(c.Ctx, "sticky", false) // 是否指定
	)
	if err := services.TopicService.SetSticky(topicId, sticky); err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

func (c *TopicController) GetHide_content() *web.JsonResult {
	topicId := params.FormValueInt64Default(c.Ctx, "topicId", 0)
	var (
		exists      = false // 是否有隐藏内容
		show        = false // 是否显示隐藏内容
		hideContent = ""    // 隐藏内容
	)
	topic := services.TopicService.Get(topicId)
	if topic != nil && topic.Status == constants.StatusOk && strs.IsNotBlank(topic.HideContent) {
		exists = true
		if user := services.UserTokenService.GetCurrent(c.Ctx); user != nil {
			if user.Id == topic.UserId || services.CommentService.IsCommented(user.Id, constants.EntityTopic, topic.Id) {
				show = true
				hideContent = markdown.ToHTML(topic.HideContent)
			}
		}
	}
	return web.JsonData(map[string]interface{}{
		"exists":  exists,
		"show":    show,
		"content": hideContent,
	})
}
