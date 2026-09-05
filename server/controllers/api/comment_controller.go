package api

import (
	"bbs-go/model"
	"bbs-go/model/constants"
	"bbs-go/pkg/errs"
	"bbs-go/spam"
	"fmt"
	"strconv"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"

	"bbs-go/controllers/render"
	"bbs-go/services"
)

type CommentController struct {
	Ctx iris.Context
}

func (c *CommentController) GetComments() *web.JsonResult {
	var (
		err         error
		cursor      int64
		page        int
		entityType  string
		entityId    int64
		IsOldBBS    bool
		entityIdStr string
		ascOrder    bool
	)
	cursor = params.FormValueInt64Default(c.Ctx, "cursor", 0)
	page = params.FormValueIntDefault(c.Ctx, "page", 0)
	ascOrder = !(params.FormValueIntDefault(c.Ctx, "asc_order", 0) == 0)

	if entityType, err = params.FormValueRequired(c.Ctx, "entityType"); err != nil {
		return web.JsonError(err)
	}
	entityIdStr = params.FormValue(c.Ctx, "entityId")
	if entityId, IsOldBBS = parseIdStr(entityIdStr); entityId == -1 {
		return web.JsonError(fmt.Errorf("bad id"))
	}
	if page > 0 {
		if !IsOldBBS {
			currentUser := services.UserTokenService.GetCurrent(c.Ctx)
			comments, paging := services.CommentService.GetCommentsPage(entityType, entityId, page, ascOrder)
			return web.JsonPageData(render.BuildComments(comments, currentUser, true, false), paging)
		}
		comments, paging := services.OldBBSService.GetCommentsPage(entityId, page, ascOrder)
		return web.JsonPageData(render.BuildComments(comments, nil, true, false), paging)
	}
	if !IsOldBBS {
		currentUser := services.UserTokenService.GetCurrent(c.Ctx)
		comments, cursor, hasMore := services.CommentService.GetComments(entityType, entityId, cursor, ascOrder)
		return web.JsonCursorData(render.BuildComments(comments, currentUser, true, false), strconv.FormatInt(cursor, 10), hasMore)
	} else {
		comments, cursor, hasMore := services.OldBBSService.GetComments(entityType, entityId, cursor, ascOrder)
		return web.JsonCursorData(render.BuildComments(comments, nil, true, false), strconv.FormatInt(cursor, 10), hasMore)
	}
}

func (c *CommentController) GetReplies() *web.JsonResult {
	var (
		cursor    = params.FormValueInt64Default(c.Ctx, "cursor", 0)
		commentId = params.FormValueInt64Default(c.Ctx, "commentId", 0)
	)
	currentUser := services.UserTokenService.GetCurrent(c.Ctx)
	comments, cursor, hasMore := services.CommentService.GetReplies(commentId, cursor, 10)
	return web.JsonCursorData(render.BuildComments(comments, currentUser, false, true), strconv.FormatInt(cursor, 10), hasMore)
}

func (c *CommentController) GetUserComments() *web.JsonResult {
	userId, err := params.FormValueInt64(c.Ctx, "userId")
	if err != nil || userId <= 0 {
		return web.JsonErrorMsg("用户不存在")
	}
	page := params.FormValueIntDefault(c.Ctx, "page", 1)
	ascOrder := params.FormValueIntDefault(c.Ctx, "asc_order", 0) != 0
	comments, paging := services.CommentService.FindUserCommentsPage(userId, page, ascOrder)
	currentUser := services.UserTokenService.GetCurrent(c.Ctx)
	return web.JsonPageData(render.BuildUserComments(comments, currentUser), paging)
}

func (c *CommentController) PostCreate() *web.JsonResult {
	user := services.UserTokenService.GetCurrent(c.Ctx)
	if err := services.UserService.CheckPostStatus(user); err != nil {
		return web.JsonError(err)
	}
	form := model.GetCreateCommentForm(c.Ctx)
	if err := spam.CheckComment(user, form); err != nil {
		return web.JsonError(err)
	}

	comment, err := services.CommentService.Publish(user.Id, form)
	if err != nil {
		return web.JsonError(err)
	}

	return web.JsonData(render.BuildComment(comment, user))
}

func (c *CommentController) PostEditBy(commentId int64) *web.JsonResult {
	user := services.UserTokenService.GetCurrent(c.Ctx)
	if err := services.UserService.CheckPostStatus(user); err != nil {
		return web.JsonError(err)
	}
	comment := services.CommentService.Get(commentId)
	if comment == nil || comment.Status != constants.StatusOk {
		return web.JsonErrorMsg("评论不存在或已被删除")
	}
	if !services.CommentService.CanManage(user, comment) {
		return web.JsonErrorMsg("无权限")
	}
	form := model.GetCreateCommentForm(c.Ctx)
	if err := services.CommentService.Edit(commentId, user.Id, form.Content, form.ImageList); err != nil {
		return web.JsonError(err)
	}
	services.OperateLogService.AddOperateLog(user.Id, constants.OpTypeUpdate, constants.EntityComment, commentId,
		"", c.Ctx.Request())
	return web.JsonData(render.BuildComment(services.CommentService.Get(commentId), user))
}

func (c *CommentController) PostDeleteBy(commentId int64) *web.JsonResult {
	user := services.UserTokenService.GetCurrent(c.Ctx)
	if err := services.UserService.CheckPostStatus(user); err != nil {
		return web.JsonError(err)
	}
	comment := services.CommentService.Get(commentId)
	if comment == nil || comment.Status == constants.StatusDeleted {
		return web.JsonSuccess()
	}
	if !services.CommentService.CanDelete(user, comment) {
		return web.JsonErrorMsg("无权限")
	}
	if err := services.CommentService.DeleteWithCounts(comment); err != nil {
		return web.JsonError(err)
	}
	services.OperateLogService.AddOperateLog(user.Id, constants.OpTypeDelete, constants.EntityComment, commentId,
		"", c.Ctx.Request())
	return web.JsonSuccess()
}

func (c *CommentController) PostLikeBy(commentId int64) *web.JsonResult {
	user := services.UserTokenService.GetCurrent(c.Ctx)
	if user == nil {
		return web.JsonError(errs.NotLogin)
	}
	err := services.UserLikeService.CommentLike(user.Id, commentId)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}
