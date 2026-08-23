package api

import (
	"time"

	"bbs-go/model"
	"bbs-go/model/constants"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web"
)

type StatsController struct {
	Ctx iris.Context
}

func (c *StatsController) GetSite() *web.JsonResult {
	now := time.Now()
	monthStart := dates.Timestamp(time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()))
	monthEnd := dates.Timestamp(time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location()))
	db := sqls.DB()

	var totalTopics int64
	var monthlyTopics int64
	var totalUsers int64
	var monthlyActiveUsers int64

	db.Model(&model.Topic{}).Where("status = ?", constants.StatusOk).Count(&totalTopics)
	db.Model(&model.Topic{}).
		Where("status = ? AND create_time >= ? AND create_time < ?", constants.StatusOk, monthStart, monthEnd).
		Count(&monthlyTopics)
	db.Model(&model.User{}).Where("status = ?", constants.StatusOk).Count(&totalUsers)

	db.Raw(`
		SELECT COUNT(DISTINCT active_users.user_id)
		FROM (
			SELECT user_id FROM t_user_token WHERE create_time >= ? AND create_time < ?
			UNION ALL
			SELECT user_id FROM t_topic WHERE status = ? AND create_time >= ? AND create_time < ?
			UNION ALL
			SELECT user_id FROM t_comment WHERE status = ? AND create_time >= ? AND create_time < ?
			UNION ALL
			SELECT user_id FROM t_check_in WHERE update_time >= ? AND update_time < ?
		) AS active_users
		INNER JOIN t_user ON t_user.id = active_users.user_id
		WHERE t_user.status = ?
	`, monthStart, monthEnd,
		constants.StatusOk, monthStart, monthEnd,
		constants.StatusOk, monthStart, monthEnd,
		monthStart, monthEnd, constants.StatusOk).
		Scan(&monthlyActiveUsers)

	return web.JsonData(map[string]int64{
		"totalTopics":        totalTopics,
		"monthlyTopics":      monthlyTopics,
		"totalUsers":         totalUsers,
		"monthlyActiveUsers": monthlyActiveUsers,
	})
}
