package services

import (
	"bbs-go/cache"
	"bbs-go/model"
	"bbs-go/model/constants"
	"bbs-go/pkg/bbsurls"
	"bbs-go/pkg/dingtalk"
	"bbs-go/pkg/email"
	"bbs-go/pkg/msg"
	"bbs-go/repositories"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/common/jsons"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"github.com/sirupsen/logrus"
)

var MessageService = newMessageService()

func newMessageService() *messageService {
	return &messageService{}
}

type messageService struct {
}

func (s *messageService) Get(id int64) *model.Message {
	return repositories.MessageRepository.Get(sqls.DB(), id)
}

func (s *messageService) Take(where ...interface{}) *model.Message {
	return repositories.MessageRepository.Take(sqls.DB(), where...)
}

func (s *messageService) Find(cnd *sqls.Cnd) []model.Message {
	return repositories.MessageRepository.Find(sqls.DB(), cnd)
}

func (s *messageService) FindOne(cnd *sqls.Cnd) *model.Message {
	return repositories.MessageRepository.FindOne(sqls.DB(), cnd)
}

func (s *messageService) FindPageByParams(params *params.QueryParams) (list []model.Message, paging *sqls.Paging) {
	return repositories.MessageRepository.FindPageByParams(sqls.DB(), params)
}

func (s *messageService) FindPageByCnd(cnd *sqls.Cnd) (list []model.Message, paging *sqls.Paging) {
	return repositories.MessageRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *messageService) Create(t *model.Message) error {
	return repositories.MessageRepository.Create(sqls.DB(), t)
}

func (s *messageService) Update(t *model.Message) error {
	return repositories.MessageRepository.Update(sqls.DB(), t)
}

func (s *messageService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.MessageRepository.Updates(sqls.DB(), id, columns)
}

func (s *messageService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.MessageRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *messageService) Delete(id int64) {
	repositories.MessageRepository.Delete(sqls.DB(), id)
}

// GetUnReadCount 获取未读消息数量
func (s *messageService) GetUnReadCount(userId int64) (count int64) {
	sqls.DB().Where("user_id = ? and status = ?", userId, msg.StatusUnread).Model(&model.Message{}).Count(&count)
	return
}

// MarkRead 将所有消息标记为已读
func (s *messageService) MarkRead(userId int64) {
	sqls.DB().Exec("update t_message set status = ? where user_id = ? and status = ?", msg.StatusHaveRead,
		userId, msg.StatusUnread)
}

// SendMsg 发送消息
func (s *messageService) SendMsg(from, to int64, msgType msg.Type,
	title, content, quoteContent string, extraData interface{}) {

	t := &model.Message{
		FromId:       from,
		UserId:       to,
		Title:        title,
		Content:      content,
		QuoteContent: quoteContent,
		Type:         int(msgType),
		ExtraData:    jsons.ToJsonStr(extraData),
		Status:       msg.StatusUnread,
		CreateTime:   dates.NowTimestamp(),
	}
	if err := s.Create(t); err != nil {
		logrus.Error(err)
	} else {
		s.SendExternalNotices(t)
	}
}

func (s *messageService) SendExternalNotices(t *model.Message) {
	s.SendEmailNotice(t)
	s.SendDingTalkNoticeAsync(t)
}

func (s *messageService) SendDingTalkNoticeAsync(t *model.Message) {
	go func() {
		if err := s.SendDingTalkNotice(t); err != nil {
			logrus.WithField("userId", t.UserId).Error(err)
		}
	}()
}

func (s *messageService) SendDingTalkNotice(t *model.Message) error {
	if t == nil {
		return nil
	}
	setting := UserNotificationSettingService.GetByUserId(t.UserId)
	if setting == nil || !setting.DingTalkEnabled {
		return nil
	}
	title := strings.TrimSpace(t.Title)
	if title == "" {
		title = "新消息提醒"
	}
	title = strings.NewReplacer("\r", " ", "\n", " ").Replace(title)
	displayTitle, markdownTitle := buildDingTalkTitle(t.FromId, title)
	if setting.DingTalkKeyword != "" {
		displayTitle = setting.DingTalkKeyword + " " + displayTitle
		markdownTitle = setting.DingTalkKeyword + " " + markdownTitle
	}
	siteTitle := cache.SysConfigCache.GetValue(constants.SysConfigSiteTitle)
	markdown := "### " + markdownTitle
	if strings.TrimSpace(t.Content) != "" {
		markdown += "\n\n" + t.Content
	}
	if strings.TrimSpace(t.QuoteContent) != "" {
		markdown += "\n\n> " + strings.ReplaceAll(t.QuoteContent, "\n", "\n> ")
	}
	markdown += "\n\n[点击前往 " + siteTitle + " 查看详情](" + bbsurls.AbsUrl("/user/messages") + ")"
	return dingtalk.Send(setting.DingTalkWebhook, setting.DingTalkSecret, dingtalk.Message{
		Title: displayTitle,
		Text:  truncateDingTalkText(markdown, 12000),
	})
}

func buildDingTalkTitle(fromId int64, title string) (displayTitle, markdownTitle string) {
	if fromId <= 0 {
		displayTitle = "系统通知：" + title
		return displayTitle, displayTitle
	}
	from := cache.UserCache.Get(fromId)
	if from == nil || strings.TrimSpace(from.Nickname) == "" {
		displayTitle = "用户通知：" + title
		return displayTitle, displayTitle
	}
	nickname := strings.NewReplacer("\r", " ", "\n", " ").Replace(strings.TrimSpace(from.Nickname))
	displayTitle = nickname + " " + title
	markdownTitle = "[" + escapeDingTalkMarkdownText(nickname) + "](" + bbsurls.UserUrl(fromId) + ") " + title
	return displayTitle, markdownTitle
}

func escapeDingTalkMarkdownText(value string) string {
	value = strings.NewReplacer(
		"\\", "\\\\",
		"[", "\\[",
		"]", "\\]",
	).Replace(value)
	return value
}

func (s *messageService) SendDingTalkTest(userId int64) error {
	setting := UserNotificationSettingService.GetByUserId(userId)
	if setting == nil || strings.TrimSpace(setting.DingTalkWebhook) == "" {
		return errors.New("请先保存钉钉 Webhook")
	}
	title := "钉钉机器人测试通知"
	if setting.DingTalkKeyword != "" {
		title = setting.DingTalkKeyword + " " + title
	}
	return dingtalk.Send(setting.DingTalkWebhook, setting.DingTalkSecret, dingtalk.Message{
		Title: title,
		Text:  "### " + title + "\n\n配置成功，之后的站内消息会发送到这个机器人。",
	})
}

func truncateDingTalkText(value string, maxRunes int) string {
	if utf8.RuneCountInString(value) <= maxRunes {
		return value
	}
	runes := []rune(value)
	return string(runes[:maxRunes-1]) + "…"
}

// SendEmailNotice 发送邮件通知
func (s *messageService) SendEmailNotice(t *model.Message) {
	msgType := msg.Type(t.Type)

	// 话题被删除不发送邮件提醒
	if msgType == msg.TypeTopicDelete {
		return
	}
	user := cache.UserCache.Get(t.UserId)
	if user == nil || len(user.Email.String) == 0 {
		return
	}
	var (
		siteTitle  = cache.SysConfigCache.GetValue(constants.SysConfigSiteTitle)
		emailTitle = siteTitle + " - 新消息提醒"
	)

	if msgType == msg.TypeTopicComment {
		emailTitle = siteTitle + " - 收到话题评论"
	} else if msgType == msg.TypeCommentReply {
		emailTitle = siteTitle + " - 收到他人回复"
	} else if msgType == msg.TypeTopicLike {
		emailTitle = siteTitle + " - 收到点赞"
	} else if msgType == msg.TypeTopicFavorite {
		emailTitle = siteTitle + " - 话题被收藏"
	} else if msgType == msg.TypeTopicRecommend {
		emailTitle = siteTitle + " - 话题被设为推荐"
	} else if msgType == msg.TypeTopicDelete {
		emailTitle = siteTitle + " - 话题被删除"
	} else if msgType == msg.TypeArticleComment {
		emailTitle = siteTitle + " - 收到文章评论"
	} else if msgType == msg.TypeTopicGift {
		emailTitle = siteTitle + " - 收到赠米"
	}

	var from *model.User
	if t.FromId > 0 {
		from = cache.UserCache.Get(t.FromId)
	}
	err := email.SendTemplateEmail(from, user.Email.String, emailTitle, emailTitle, t.Content,
		t.QuoteContent, &model.ActionLink{
			Title: "点击查看详情",
			Url:   bbsurls.AbsUrl("/user/messages"),
		})
	if err != nil {
		logrus.Error(err)
	}
}
