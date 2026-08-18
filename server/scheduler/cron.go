package scheduler

import (
	"bbs-go/pkg/sitemap"
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

	// Generate RSS
	addCronFunc(c, "@every 30m", func() {
		services.ArticleService.GenerateRss()
		services.TopicService.GenerateRss()
	})

	// Generate sitemap
	addCronFunc(c, "0 0 4 ? * *", func() {
		sitemap.Generate()
	})

	c.Start()
}

func addCronFunc(c *cron.Cron, sepc string, cmd func()) {
	err := c.AddFunc(sepc, cmd)
	if err != nil {
		logrus.Error(err)
	}
}
