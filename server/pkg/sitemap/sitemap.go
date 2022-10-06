package sitemap

import (
	"bbs-go/model/constants"
	"bbs-go/pkg/bbsurls"
	"time"

	"github.com/ikeikeikeike/go-sitemap-generator/v2/stm"
	"github.com/mlogclub/simple/common/dates"
	"github.com/sirupsen/logrus"

	"bbs-go/model"
	"bbs-go/pkg/config"
	"bbs-go/services"
)

const (
	changefreqAlways  = "always"
	changefreqHourly  = "hourly"
	changefreqDaily   = "daily"
	changefreqWeekly  = "weekly"
	changefreqMonthly = "monthly"
	changefreqYearly  = "yearly"
	changefreqNever   = "never"
)

var building = false

// Generate
func Generate() {
	if config.Instance.Env != "prod" {
		return
	}
	if building {
		logrus.Info("Sitemap in building...")
		return
	}
	building = true
	defer func() {
		building = false
	}()

	sm := stm.NewSitemap(0)
	sm.SetDefaultHost(config.Instance.BaseUrl) // 网站host
	sm.SetSitemapsPath("sitemap2")             // sitemap存放目录
	sm.SetVerbose(false)
	sm.SetPretty(false)
	sm.SetCompress(true)
	sm.Create()

	sm.Add(stm.URL{
		{"loc", bbsurls.AbsUrl("/")},
		{"lastmod", time.Now()},
		{"changefreq", changefreqHourly},
	})

	sm.Add(stm.URL{
		{"loc", bbsurls.AbsUrl("/topics")},
		{"lastmod", time.Now()},
		{"changefreq", changefreqHourly},
	})

	sm.Add(stm.URL{
		{"loc", bbsurls.AbsUrl("/articles")},
		{"lastmod", time.Now()},
		{"changefreq", changefreqAlways},
	})

	// var (
	// 	dateFrom = dates.Timestamp(time.Now().AddDate(0, -1, 0))
	// 	dateTo   = dates.NowTimestamp()
	// )

	services.ArticleService.ScanDesc(func(articles []model.Article) {
		for _, article := range articles {
			if article.Status == constants.StatusOk {
				articleUrl := bbsurls.ArticleUrl(article.Id)
				sm.Add(stm.URL{
					{"loc", articleUrl},
					{"lastmod", dates.FromTimestamp(article.UpdateTime)},
					{"changefreq", changefreqMonthly},
					{"priority", 0.6},
				})
			}
		}
	})

	services.TopicService.ScanDesc(func(topics []model.Topic) {
		for _, topic := range topics {
			if topic.Status == constants.StatusOk {
				topicUrl := bbsurls.TopicUrl(topic.Id)
				sm.Add(stm.URL{
					{"loc", topicUrl},
					{"lastmod", dates.FromTimestamp(topic.LastCommentTime)},
					{"changefreq", changefreqDaily},
					{"priority", 0.6},
				})
			}
		}
	})

	services.TagService.Scan(func(tags []model.Tag) {
		for _, tag := range tags {
			tagUrl := bbsurls.TagArticlesUrl(tag.Id)
			sm.Add(stm.URL{
				{"loc", tagUrl},
				{"lastmod", time.Now()},
				{"changefreq", changefreqMonthly},
			})
		}
	})

	services.UserService.Scan(func(users []model.User) {
		for _, user := range users {
			sm.Add(stm.URL{
				{"loc", bbsurls.UserUrl(user.Id)},
				{"lastmod", time.Now()},
				{"changefreq", changefreqWeekly},
			})
		}
	})

	sm.Finalize().PingSearchEngines()
	// sm.Finalize().PingSearchEngines("http://www.google.cn/webmasters/tools/ping?sitemap=%s")
}

// My Adapter
type myAdapter struct {
}

// Bytes gets written content.
func (adp *myAdapter) Bytes() [][]byte {
	return nil
}
