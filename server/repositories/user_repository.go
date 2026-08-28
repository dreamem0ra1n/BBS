package repositories

import (
	"time"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"

	"bbs-go/model"
	"bbs-go/model/constants"
)

var UserRepository = newUserRepository()

func newUserRepository() *userRepository {
	return &userRepository{}
}

type userRepository struct {
}

func (r *userRepository) Get(db *gorm.DB, id int64) *model.User {
	ret := &model.User{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *userRepository) Take(db *gorm.DB, where ...interface{}) *model.User {
	ret := &model.User{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *userRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []model.User) {
	cnd.Find(db, &list)
	return
}

func (r *userRepository) FindNewbieScoreRank(db *gorm.DB, now time.Time) []model.User {
	return r.Find(db, sqls.NewCnd().
		Gte("create_time", dates.Timestamp(now.AddDate(-1, 0, 0))).
		Lt("create_time", dates.Timestamp(now.Add(time.Millisecond))).
		Desc("score").
		Limit(10))
}

func (r *userRepository) FindRecentYearScoreRank(db *gorm.DB, now time.Time) []model.User {
	return r.FindNewbieScoreRank(db, now)
}

func (r *userRepository) FindAnnualScoreRank(db *gorm.DB, now time.Time) []model.User {
	start := dates.Timestamp(now.AddDate(-1, 0, 0))
	end := dates.Timestamp(now.Add(time.Millisecond))

	type scoreTotal struct {
		UserId int64
		Score  int
	}
	var totals []scoreTotal
	if err := db.Model(&model.UserScoreLog{}).
		Select("user_id, SUM(score) AS score").
		Where("type = ? AND create_time >= ? AND create_time < ?", constants.ScoreTypeIncr, start, end).
		Group("user_id").
		Order("score DESC").
		Order("user_id ASC").
		Limit(10).
		Scan(&totals).Error; err != nil {
		return nil
	}
	if len(totals) == 0 {
		return []model.User{}
	}

	ids := make([]int64, 0, len(totals))
	for _, total := range totals {
		ids = append(ids, total.UserId)
	}
	var users []model.User
	if err := db.Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil
	}
	usersById := make(map[int64]model.User, len(users))
	for _, user := range users {
		usersById[user.Id] = user
	}

	ranked := make([]model.User, 0, len(totals))
	for _, total := range totals {
		if user, ok := usersById[total.UserId]; ok {
			user.Score = total.Score
			ranked = append(ranked, user)
		}
	}
	return ranked
}

func (r *userRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *model.User {
	ret := &model.User{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *userRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []model.User, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *userRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []model.User, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &model.User{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *userRepository) Create(db *gorm.DB, t *model.User) (err error) {
	err = db.Create(t).Error
	return
}

func (r *userRepository) Update(db *gorm.DB, t *model.User) (err error) {
	err = db.Save(t).Error
	return
}

func (r *userRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&model.User{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *userRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&model.User{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *userRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&model.User{}, "id = ?", id)
}

func (r *userRepository) GetByEmail(db *gorm.DB, email string) *model.User {
	return r.Take(db, "email = ?", email)
}

func (r *userRepository) GetByUsername(db *gorm.DB, username string) *model.User {
	return r.Take(db, "username = ?", username)
}
