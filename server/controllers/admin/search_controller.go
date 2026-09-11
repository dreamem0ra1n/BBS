package admin

import (
	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"

	"bbs-go/pkg/config"
	"bbs-go/services"
)

type SearchController struct {
	Ctx iris.Context
}

// PostReindex 触发全量重建搜索索引（异步执行，不阻塞请求）。
func (c *SearchController) PostReindex() *web.JsonResult {
	go func() {
		services.SearchService.ReindexAll()
		if config.Instance != nil && config.Instance.Search.SyncOldBBS {
			services.SearchService.ReindexOldBBS()
		}
	}()
	return web.JsonSuccess()
}
