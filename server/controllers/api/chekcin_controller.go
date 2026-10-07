package api

import (
	"bbs-go/controllers/render"
	"bbs-go/pkg/errs"
	"bbs-go/services"
	"strconv"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
	"github.com/mlogclub/simple/web/params"
)

type CheckinController struct {
	Ctx iris.Context
}

// PostCheckin 签到
func (c *CheckinController) PostCheckin() *web.JsonResult {
	user := services.UserTokenService.GetCurrent(c.Ctx)
	if user == nil { return web.JsonError(errs.NotLogin) }
	if err := services.UserService.CheckPostStatus(user); err != nil {
		return web.JsonError(err)
	}
	err := services.CheckInService.CheckIn(user.Id)
	if err == nil {
		return web.JsonData(services.CheckInService.GetSummary(user.Id, time.Now()))
	} else {
		return web.JsonError(err)
	}
}

// GetCheckin 获取签到信息
func (c *CheckinController) GetCheckin() *web.JsonResult {
	user := services.UserTokenService.GetCurrent(c.Ctx)
	if user == nil {
		return web.JsonSuccess()
	}
	checkIn := services.CheckInService.GetByUserId(user.Id)
	if checkIn != nil {
		today := checkinDate(time.Now())
		return web.NewRspBuilder(checkIn).
			Put("checkIn", checkIn.LatestDayName == today). // 今日是否已签到
			JsonResult()
	}
	return web.JsonSuccess()
}

func (c *CheckinController) GetOverview() *web.JsonResult {
	user := services.UserTokenService.GetCurrent(c.Ctx)
	if user == nil { return web.JsonError(errs.NotLogin) }
	month := c.Ctx.FormValue("month")
	data, err := services.CheckInService.Overview(user.Id, month, time.Now())
	if err != nil { return web.JsonError(err) }
	return web.JsonData(data)
}

func (c *CheckinController) PostMakeup() *web.JsonResult {
	user := services.UserTokenService.GetCurrent(c.Ctx)
	if user == nil { return web.JsonError(errs.NotLogin) }
	if err := services.UserService.CheckPostStatus(user); err != nil { return web.JsonError(err) }
	day, summary, err := services.CheckInService.Makeup(user.Id)
	if err != nil { return web.JsonError(err) }
	return web.JsonData(map[string]interface{}{"date": formatCheckInDate(day.CheckInDate), "checkInType": day.CheckInType, "consecutiveDays": summary.ConsecutiveDays, "totalCheckInDays": summary.TotalCheckInDays, "makeupCards": summary.MakeupCards})
}

func (c *CheckinController) GetRankToday() *web.JsonResult {
	page, limit := rankPaging(c.Ctx)
	rows, total, err := services.CheckInService.TodayRank(time.Now(), page, limit)
	if err != nil { return web.JsonError(err) }
	results := make([]map[string]interface{}, 0, len(rows))
	for index, row := range rows { results = append(results, map[string]interface{}{"rank": (page-1)*limit + index + 1, "user": render.BuildUserInfoDefaultIfNull(row.UserId), "checkInTime": row.CheckInTime, "checkInType": row.CheckInType}) }
	return web.JsonData(map[string]interface{}{"results": results, "page": map[string]interface{}{"page": page, "limit": limit, "total": total}})
}

func (c *CheckinController) GetRankConsecutive() *web.JsonResult {
	page, limit := rankPaging(c.Ctx)
	rows, total, err := services.CheckInService.ConsecutiveRank(page, limit)
	if err != nil { return web.JsonError(err) }
	results := make([]map[string]interface{}, 0, len(rows))
	for index, row := range rows { results = append(results, map[string]interface{}{"rank": (page-1)*limit + index + 1, "user": render.BuildUserInfoDefaultIfNull(row.UserId), "consecutiveDays": row.ConsecutiveDays, "latestCheckInDate": formatCheckInDate(row.LatestCheckInDate)}) }
	return web.JsonData(map[string]interface{}{"results": results, "page": map[string]interface{}{"page": page, "limit": limit, "total": total}})
}

// GetRank 获取当天签到排行榜（最早签到的排在最前面）
func (c *CheckinController) GetRank() *web.JsonResult {
	rows, _, err := services.CheckInService.TodayRank(time.Now(), 1, 10)
	if err != nil { return web.JsonError(err) }
	itemList := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		itemList = append(itemList, map[string]interface{}{"id": row.Id, "userId": row.UserId, "updateTime": row.CheckInTime, "checkInType": row.CheckInType, "user": render.BuildUserInfoDefaultIfNull(row.UserId)})
	}
	return web.JsonData(itemList)
}

func rankPaging(ctx iris.Context) (int, int) {
	page := params.FormValueIntDefault(ctx, "page", 1)
	limit := params.FormValueIntDefault(ctx, "pageSize", 20)
	if page < 1 { page = 1 }
	if limit < 1 { limit = 20 }
	if limit > 100 { limit = 100 }
	return page, limit
}

func formatCheckInDate(value int) string {
	return strconv.Itoa(value/10000) + "-" + twoDigits((value/100)%100) + "-" + twoDigits(value%100)
}

func checkinDate(value time.Time) int {
	local := value.In(time.Local)
	return local.Year()*10000 + int(local.Month())*100 + local.Day()
}

func twoDigits(value int) string { if value < 10 { return "0" + strconv.Itoa(value) }; return strconv.Itoa(value) }
