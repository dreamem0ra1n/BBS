package api

import (
	"time"

	"bbs-go/model"
	"bbs-go/model/constants"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web"
)

type StatsController struct {
	Ctx iris.Context
}

func (c *StatsController) GetSite() *web.JsonResult {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Unix()
	db := sqls.DB()

	var totalTopics int64
	var monthlyTopics int64
	var totalUsers int64
	var monthlyActiveUsers int64

	db.Model(&model.Topic{}).Where("status = ?", constants.StatusOk).Count(&totalTopics)
	db.Model(&model.Topic{}).
		Where("status = ? AND create_time >= ?", constants.StatusOk, monthStart).
		Count(&monthlyTopics)
	db.Model(&model.User{}).Where("status = ?", constants.StatusOk).Count(&totalUsers)

	db.Raw(`
		SELECT COUNT(DISTINCT user_id)
		FROM (
			SELECT user_id FROM t_user_token WHERE create_time >= ?
			UNION ALL
			SELECT user_id FROM t_topic WHERE status = ? AND create_time >= ?
			UNION ALL
			SELECT user_id FROM t_comment WHERE status = ? AND create_time >= ?
			UNION ALL
			SELECT user_id FROM t_check_in WHERE update_time >= ?
		) AS active_users
	`, monthStart, constants.StatusOk, monthStart, constants.StatusOk, monthStart, monthStart).
		Scan(&monthlyActiveUsers)

	return web.JsonData(map[string]int64{
		"totalTopics":        totalTopics,
		"monthlyTopics":      monthlyTopics,
		"totalUsers":         totalUsers,
		"monthlyActiveUsers": monthlyActiveUsers,
	})
}
