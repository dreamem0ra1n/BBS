package services

import (
	"fmt"
	"time"

	"bbs-go/model"
	"bbs-go/model/constants"
	"bbs-go/pkg/msg"
	"bbs-go/repositories"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/common/jsons"
	"github.com/mlogclub/simple/sqls"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const birthdayLayout = "2006-01-02"

var BirthdayService = newBirthdayService()

type birthdayService struct{}

type birthdayExtraData struct {
	Year int `json:"birthdayYear"`
}

func newBirthdayService() *birthdayService {
	return &birthdayService{}
}

// SendNotices 每个自然年为生日用户发送一次祝福。
func (s *birthdayService) SendNotices(now time.Time) {
	UserService.Scan(func(users []model.User) {
		for index := range users {
			s.sendNoticeAndLog(&users[index], now)
		}
	})
}

// SendNotice 在用户更新资料时检查并发送当天的生日祝福。
func (s *birthdayService) SendNotice(userId int64, now time.Time) {
	user := UserService.Get(userId)
	if user != nil {
		s.sendNoticeAndLog(user, now)
	}
}

func (s *birthdayService) sendNoticeAndLog(user *model.User, now time.Time) {
	if err := s.sendNotice(user, now); err != nil {
		logrus.WithField("userId", user.Id).Error(err)
	}
}

func (s *birthdayService) sendNotice(user *model.User, now time.Time) error {
	if user == nil || user.Status != constants.StatusOk {
		return nil
	}
	birthday, err := time.ParseInLocation(birthdayLayout, user.Birthday, now.Location())
	if err != nil || birthday.After(now) || birthday.Month() != now.Month() || birthday.Day() != now.Day() {
		return nil
	}
	age := now.Year() - birthday.Year()

	notification := &model.Message{
		FromId:     0,
		UserId:     user.Id,
		Title:      "生日快乐！",
		Content:    fmt.Sprintf("亲爱的潮人 %s ：今天是你%d岁的生日，求是潮BBS祝你生日快乐！愿你永远有大步向前的勇气，永远有一颗真诚的心，也祝你学习进步，工作顺利。但更重要的是，我们希望你身体健康，无忧无虑。浪潮不息，求是潮BBS永远是你的港湾，每朵浪花我们都记念于心^_^", user.Nickname, age),
		Type:       int(msg.TypeBirthday),
		ExtraData:  jsons.ToJsonStr(birthdayExtraData{Year: now.Year()}),
		Status:     msg.StatusUnread,
		CreateTime: dates.NowTimestamp(),
	}

	created := false
	err = sqls.DB().Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.User{}).
			Where("id = ? AND birthday_year_sent < ?", user.Id, now.Year()).
			UpdateColumn("birthday_year_sent", now.Year())
		if result.Error != nil || result.RowsAffected == 0 {
			return result.Error
		}
		if err = repositories.MessageRepository.Create(tx, notification); err != nil {
			return err
		}
		created = true
		return nil
	})
	if err == nil && created {
		MessageService.SendDingTalkNoticeAsync(notification)
	}
	return err
}
