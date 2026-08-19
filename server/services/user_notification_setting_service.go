package services

import (
	"errors"
	"strings"
	"unicode/utf8"

	"bbs-go/model"
	"bbs-go/pkg/dingtalk"
	"bbs-go/repositories"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/sqls"
)

var UserNotificationSettingService = newUserNotificationSettingService()

type userNotificationSettingService struct{}

type DingTalkSettingsInput struct {
	Enabled     bool
	Webhook     string
	Secret      string
	Keyword     string
	ClearSecret bool
}

func newUserNotificationSettingService() *userNotificationSettingService {
	return &userNotificationSettingService{}
}

func (s *userNotificationSettingService) GetByUserId(userId int64) *model.UserNotificationSetting {
	return repositories.UserNotificationSettingRepository.GetByUserId(sqls.DB(), userId)
}

func (s *userNotificationSettingService) GetDingTalkSettings(userId int64) *model.DingTalkSettingsResponse {
	setting := s.GetByUserId(userId)
	if setting == nil {
		return &model.DingTalkSettingsResponse{}
	}
	return &model.DingTalkSettingsResponse{
		Enabled:           setting.DingTalkEnabled,
		WebhookConfigured: strings.TrimSpace(setting.DingTalkWebhook) != "",
		SecretConfigured:  strings.TrimSpace(setting.DingTalkSecret) != "",
		Keyword:           setting.DingTalkKeyword,
	}
}

func (s *userNotificationSettingService) UpdateDingTalkSettings(userId int64, input DingTalkSettingsInput) error {
	input.Webhook = strings.TrimSpace(input.Webhook)
	input.Secret = strings.TrimSpace(input.Secret)
	input.Keyword = strings.TrimSpace(input.Keyword)
	if len(input.Webhook) > 2048 {
		return errors.New("钉钉 Webhook 地址过长")
	}
	if len(input.Secret) > 512 {
		return errors.New("钉钉加签密钥过长")
	}
	if utf8.RuneCountInString(input.Keyword) > 64 {
		return errors.New("钉钉关键词不能超过64个字")
	}
	if input.Webhook != "" {
		if err := dingtalk.ValidateWebhook(input.Webhook); err != nil {
			return err
		}
	}

	setting := s.GetByUserId(userId)
	now := dates.NowTimestamp()
	if setting == nil {
		setting = &model.UserNotificationSetting{UserId: userId, CreateTime: now}
	}
	if input.Webhook != "" {
		setting.DingTalkWebhook = input.Webhook
	}
	if input.Secret != "" {
		setting.DingTalkSecret = input.Secret
	} else if input.ClearSecret {
		setting.DingTalkSecret = ""
	}
	setting.DingTalkEnabled = input.Enabled
	setting.DingTalkKeyword = input.Keyword
	setting.UpdateTime = now
	if setting.DingTalkEnabled {
		if setting.DingTalkWebhook == "" {
			return errors.New("开启钉钉通知前请填写 Webhook")
		}
		if err := dingtalk.ValidateWebhook(setting.DingTalkWebhook); err != nil {
			return err
		}
	}
	if setting.Id == 0 {
		return repositories.UserNotificationSettingRepository.Create(sqls.DB(), setting)
	}
	return repositories.UserNotificationSettingRepository.Update(sqls.DB(), setting)
}
