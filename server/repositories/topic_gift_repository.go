package repositories

import (
	"bbs-go/model"

	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

var TopicGiftRepository = newTopicGiftRepository()

type topicGiftRepository struct{}

func newTopicGiftRepository() *topicGiftRepository {
	return &topicGiftRepository{}
}

func (r *topicGiftRepository) Create(db *gorm.DB, gift *model.TopicGift) error {
	return db.Create(gift).Error
}

func (r *topicGiftRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []model.TopicGift) {
	cnd.Find(db, &list)
	return
}
