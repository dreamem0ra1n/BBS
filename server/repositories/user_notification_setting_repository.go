package repositories

import (
	"bbs-go/model"

	"gorm.io/gorm"
)

var UserNotificationSettingRepository = newUserNotificationSettingRepository()

type userNotificationSettingRepository struct{}

func newUserNotificationSettingRepository() *userNotificationSettingRepository {
	return &userNotificationSettingRepository{}
}

func (r *userNotificationSettingRepository) GetByUserId(db *gorm.DB, userId int64) *model.UserNotificationSetting {
	ret := &model.UserNotificationSetting{}
	if err := db.Take(ret, "user_id = ?", userId).Error; err != nil {
		return nil
	}
	return ret
}

func (r *userNotificationSettingRepository) Create(db *gorm.DB, setting *model.UserNotificationSetting) error {
	return db.Create(setting).Error
}

func (r *userNotificationSettingRepository) Update(db *gorm.DB, setting *model.UserNotificationSetting) error {
	return db.Save(setting).Error
}
