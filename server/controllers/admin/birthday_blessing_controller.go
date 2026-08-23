package admin

import (
	"bbs-go/model"
	"bbs-go/services"
	"encoding/csv"
	"io"
	"strconv"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type BirthdayBlessingController struct{ Ctx iris.Context }

func (c *BirthdayBlessingController) AnyList() *web.JsonResult {
	p := params.NewQueryParams(c.Ctx).LikeByReq("nickname").LikeByReq("department").LikeByReq("content").PageByReq().Desc("id")
	list, paging := services.BirthdayBlessingService.FindPageByParams(p)
	return web.JsonData(&web.PageResult{Results: list, Page: paging})
}

func (c *BirthdayBlessingController) PostCreate() *web.JsonResult {
	item := &model.BirthdayBlessing{}
	if err := params.ReadForm(c.Ctx, item); err != nil {
		return web.JsonError(err)
	}
	if !services.BirthdayBlessingService.Normalize(item) {
		return web.JsonErrorMsg("祝福内容不能为空")
	}
	item.CreateTime = dates.NowTimestamp()
	if err := services.BirthdayBlessingService.Create(item); err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(item)
}

func (c *BirthdayBlessingController) PostImport() *web.JsonResult {
	item := &model.BirthdayBlessing{Nickname: c.Ctx.FormValue("nickname"), Department: c.Ctx.FormValue("department"), Content: c.Ctx.FormValue("content")}
	if !services.BirthdayBlessingService.Normalize(item) {
		return web.JsonErrorMsg("祝福内容不能为空")
	}
	item.CreateTime = dates.NowTimestamp()
	if err := services.BirthdayBlessingService.Create(item); err != nil {
		return web.JsonError(err)
	}
	return web.JsonData(item)
}

func (c *BirthdayBlessingController) PostBatchImport() *web.JsonResult {
	reader := csv.NewReader(strings.NewReader(c.Ctx.FormValue("data")))
	reader.FieldsPerRecord = -1
	created := 0
	for {
		parts, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return web.JsonErrorMsg("CSV 格式错误：" + err.Error())
		}
		if len(parts) < 3 {
			continue
		}
		first := strings.TrimSpace(strings.TrimPrefix(parts[0], "\ufeff"))
		if strings.EqualFold(first, "昵称") && strings.EqualFold(strings.TrimSpace(parts[1]), "部门") {
			continue
		}
		item := &model.BirthdayBlessing{Nickname: parts[0], Department: parts[1], Content: parts[2], CreateTime: dates.NowTimestamp()}
		if !services.BirthdayBlessingService.Normalize(item) {
			continue
		}
		if err := services.BirthdayBlessingService.Create(item); err != nil {
			return web.JsonError(err)
		}
		created++
	}
	return web.JsonData(map[string]int{"created": created})
}

func (c *BirthdayBlessingController) PostDelete() *web.JsonResult {
	ids := params.FormValueInt64Array(c.Ctx, "ids")
	if len(ids) == 0 {
		for _, value := range strings.Split(c.Ctx.FormValue("ids"), ",") {
			if id, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64); err == nil {
				ids = append(ids, id)
			}
		}
	}
	if len(ids) == 0 {
		if id, err := strconv.ParseInt(c.Ctx.FormValue("id"), 10, 64); err == nil {
			ids = []int64{id}
		}
	}
	if len(ids) == 0 {
		return web.JsonErrorMsg("请选择要删除的祝福")
	}
	for _, id := range ids {
		if err := services.BirthdayBlessingService.Delete(id); err != nil {
			return web.JsonError(err)
		}
	}
	return web.JsonSuccess()
}
