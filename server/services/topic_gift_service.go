package services

import (
	"bbs-go/cache"
	"bbs-go/model"
	"bbs-go/model/constants"
	"bbs-go/pkg/msg"
	"bbs-go/repositories"
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/common/jsons"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	defaultGiftScoreMax = 50
	maxGiftScoreMax     = 50
)

var TopicGiftService = newTopicGiftService()

type topicGiftService struct{}

func newTopicGiftService() *topicGiftService {
	return &topicGiftService{}
}

func (s *topicGiftService) FindByTopicId(topicId int64) []model.TopicGift {
	return repositories.TopicGiftRepository.Find(sqls.DB(), sqls.NewCnd().Eq("topic_id", topicId).Asc("id"))
}

func (s *topicGiftService) Gift(userId, topicId int64, score int, reason string) (*model.TopicGift, error) {
	reason = strings.TrimSpace(reason)
	if err := validateTopicGift(score, reason, SysConfigService.GetGiftScoreMax()); err != nil {
		return nil, err
	}

	topic := repositories.TopicRepository.Get(sqls.DB(), topicId)
	if topic == nil || topic.Status != constants.StatusOk {
		return nil, errors.New("话题不存在")
	}
	if topic.UserId == userId {
		return nil, errors.New("不能给自己创建的话题赠米")
	}

	var gift *model.TopicGift
	var notification *model.Message
	err := sqls.DB().Transaction(func(tx *gorm.DB) error {
		var users []model.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id in ?", []int64{userId, topic.UserId}).Order("id").Find(&users).Error; err != nil {
			return err
		}
		if len(users) != 2 {
			return errors.New("用户不存在")
		}

		var giver *model.User
		for index := range users {
			if users[index].Id == userId {
				giver = &users[index]
				break
			}
		}
		if giver == nil {
			return errors.New("用户不存在")
		}
		if giver.Score < score {
			return errors.New("积分不足")
		}

		now := dates.NowTimestamp()
		if err := repositories.UserRepository.Updates(tx, userId, map[string]interface{}{
			"score":       gorm.Expr("score - ?", score),
			"update_time": now,
		}); err != nil {
			return err
		}
		if err := repositories.UserRepository.Updates(tx, topic.UserId, map[string]interface{}{
			"score":       gorm.Expr("score + ?", score),
			"update_time": now,
		}); err != nil {
			return err
		}

		gift = &model.TopicGift{
			TopicId:    topic.Id,
			UserId:     userId,
			ReceiverId: topic.UserId,
			Score:      score,
			Reason:     reason,
			CreateTime: now,
		}
		if err := repositories.TopicGiftRepository.Create(tx, gift); err != nil {
			return err
		}

		sourceId := strconv.FormatInt(gift.Id, 10)
		topicTitle := "《" + topic.GetTitle() + "》"
		logs := []model.UserScoreLog{
			{
				UserId: userId, SourceType: constants.EntityTopicGift, SourceId: sourceId,
				Description: "给话题" + topicTitle + "赠米", Type: constants.ScoreTypeDecr,
				Score: -score, CreateTime: now,
			},
			{
				UserId: topic.UserId, SourceType: constants.EntityTopicGift, SourceId: sourceId,
				Description: "话题" + topicTitle + "收到赠米", Type: constants.ScoreTypeIncr,
				Score: score, CreateTime: now,
			},
		}
		if err := tx.Create(&logs).Error; err != nil {
			return err
		}

		notification = &model.Message{
			FromId: userId, UserId: topic.UserId,
			Title: "给你的话题赠米",
			Content: reason, QuoteContent: topicTitle, Type: int(msg.TypeTopicGift),
			ExtraData: jsons.ToJsonStr(&msg.TopicGiftExtraData{
				TopicId: topic.Id, GiftId: gift.Id, Score: gift.Score,
			}),
			Status: msg.StatusUnread, CreateTime: now,
		}
		return repositories.MessageRepository.Create(tx, notification)
	})
	if err != nil {
		return nil, err
	}

	cache.UserCache.Invalidate(userId)
	cache.UserCache.Invalidate(topic.UserId)
	MessageService.SendExternalNotices(notification)
	return gift, nil
}

func validateTopicGift(score int, reason string, maxScore int) error {
	if maxScore <= 0 {
		maxScore = defaultGiftScoreMax
	}
	if maxScore > maxGiftScoreMax {
		maxScore = maxGiftScoreMax
	}
	if score < 1 || score > maxScore {
		return errors.New("赠米数量必须在1-" + strconv.Itoa(maxScore) + "之间")
	}
	if strings.TrimSpace(reason) == "" {
		return errors.New("请输入赠米理由")
	}
	if utf8.RuneCountInString(reason) > 15 {
		return errors.New("赠米理由不能超过15个字")
	}
	return nil
}
