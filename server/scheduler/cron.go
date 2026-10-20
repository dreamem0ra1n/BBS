package scheduler

import (
	"bbs-go/controllers/api"
	"bbs-go/model"
	"bbs-go/pkg/config"
	"bbs-go/pkg/sitemap"
	"context"
	"time"

	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"

	"bbs-go/services"
)

func Start() {
	c := cron.New()

	// 启动时补发当天祝福，之后每天定时发送。
	go services.BirthdayService.SendNotices(time.Now())
	addCronFunc(c, "0 5 0 * * *", func() {
		services.BirthdayService.SendNotices(time.Now())
	})
	addCronFunc(c, "0 0 3 * * *", func() {
		result, err := services.FileService.CleanupPreviousDay(context.Background(), time.Now(), func(file *model.FileRecord) error {
			return api.RemoveObject(file.BucketName, file.ObjectName, file.FileUUID)
		})
		entry := logrus.WithFields(logrus.Fields{
			"scanned": result.Scanned, "deleted": result.Deleted,
			"referenced": result.Referenced, "failed": result.Failed, "freedBytes": result.FreedBytes,
		})
		if err != nil {
			entry.WithError(err).Error("daily file cleanup interrupted")
		} else {
			entry.Info("daily file cleanup completed")
		}
	})

	// Generate RSS
	addCronFunc(c, "@every 30m", func() {
		services.ArticleService.GenerateRss()
		services.TopicService.GenerateRss()
	})

	// Generate sitemap
	addCronFunc(c, "0 0 4 ? * *", func() {
		sitemap.Generate()
	})

	// 搜索索引对账重建：修复漏同步的文档。
	// 旧站数据量大，仅在开启 SyncOldBBS 时一起重建。
	addCronFunc(c, "0 30 4 * * *", func() {
		services.SearchService.ReindexAll()
		if config.Instance != nil && config.Instance.Search.SyncOldBBS {
			services.SearchService.ReindexOldBBS()
		}
	})

	c.Start()
}

func addCronFunc(c *cron.Cron, sepc string, cmd func()) {
	err := c.AddFunc(sepc, cmd)
	if err != nil {
		logrus.Error(err)
	}
}
