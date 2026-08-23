package repositories

import (
	"bbs-go/model"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var BirthdayBlessingRepository = newBirthdayBlessingRepository()

type birthdayBlessingRepository struct{}

func newBirthdayBlessingRepository() *birthdayBlessingRepository {
	return &birthdayBlessingRepository{}
}

func (r *birthdayBlessingRepository) Get(db *gorm.DB, id int64) *model.BirthdayBlessing {
	ret := &model.BirthdayBlessing{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *birthdayBlessingRepository) ExistsByNickname(db *gorm.DB, nickname string) bool {
	var count int64
	db.Model(&model.BirthdayBlessing{}).Where("nickname = ?", nickname).Count(&count)
	return count > 0
}

func (r *birthdayBlessingRepository) FindPageByParams(db *gorm.DB, p *params.QueryParams) (list []model.BirthdayBlessing, paging *sqls.Paging) {
	p.Cnd.Find(db, &list)
	paging = &sqls.Paging{Page: p.Cnd.Paging.Page, Limit: p.Cnd.Paging.Limit, Total: p.Cnd.Count(db, &model.BirthdayBlessing{})}
	return
}

func (r *birthdayBlessingRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []model.BirthdayBlessing) {
	cnd.Find(db, &list)
	return
}
func (r *birthdayBlessingRepository) Create(db *gorm.DB, item *model.BirthdayBlessing) error {
	return db.Create(item).Error
}
func (r *birthdayBlessingRepository) Delete(db *gorm.DB, id int64) error {
	return db.Delete(&model.BirthdayBlessing{}, "id = ?", id).Error
}
