package repositories

import (
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"

	"bbs-go/model"
)

var FileRepository = newFileRepository()

func newFileRepository() *fileRepository {
	return &fileRepository{}
}

type fileRepository struct{}

func (r *fileRepository) Get(db *gorm.DB, id int64) *model.FileRecord {
	ret := &model.FileRecord{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *fileRepository) Take(db *gorm.DB, where ...interface{}) *model.FileRecord {
	ret := &model.FileRecord{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *fileRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []model.FileRecord) {
	cnd.Find(db, &list)
	return
}

func (r *fileRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *model.FileRecord {
	ret := &model.FileRecord{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *fileRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []model.FileRecord, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *fileRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []model.FileRecord, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &model.FileRecord{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *fileRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &model.FileRecord{})
}

func (r *fileRepository) Create(db *gorm.DB, t *model.FileRecord) (err error) {
	err = db.Create(t).Error
	return
}

func (r *fileRepository) Update(db *gorm.DB, t *model.FileRecord) (err error) {
	err = db.Save(t).Error
	return
}

func (r *fileRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&model.FileRecord{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *fileRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&model.FileRecord{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *fileRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&model.FileRecord{}, "id = ?", id)
}
