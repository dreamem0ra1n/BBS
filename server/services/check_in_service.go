package services

import (
	"bbs-go/cache"
	"bbs-go/model"
	"bbs-go/model/constants"
	"bbs-go/repositories"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var CheckInService = newCheckInService()

func newCheckInService() *checkInService {
	return &checkInService{}
}

type checkInService struct {
}

func (s *checkInService) Get(id int64) *model.CheckIn {
	return repositories.CheckInRepository.Get(sqls.DB(), id)
}

func (s *checkInService) Take(where ...interface{}) *model.CheckIn {
	return repositories.CheckInRepository.Take(sqls.DB(), where...)
}

func (s *checkInService) Find(cnd *sqls.Cnd) []model.CheckIn {
	return repositories.CheckInRepository.Find(sqls.DB(), cnd)
}

func (s *checkInService) FindOne(cnd *sqls.Cnd) *model.CheckIn {
	return repositories.CheckInRepository.FindOne(sqls.DB(), cnd)
}

func (s *checkInService) FindPageByParams(params *params.QueryParams) (list []model.CheckIn, paging *sqls.Paging) {
	return repositories.CheckInRepository.FindPageByParams(sqls.DB(), params)
}

func (s *checkInService) FindPageByCnd(cnd *sqls.Cnd) (list []model.CheckIn, paging *sqls.Paging) {
	return repositories.CheckInRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *checkInService) Count(cnd *sqls.Cnd) int64 {
	return repositories.CheckInRepository.Count(sqls.DB(), cnd)
}

func (s *checkInService) Create(t *model.CheckIn) error {
	return repositories.CheckInRepository.Create(sqls.DB(), t)
}

func (s *checkInService) Update(t *model.CheckIn) error {
	return repositories.CheckInRepository.Update(sqls.DB(), t)
}

func (s *checkInService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.CheckInRepository.Updates(sqls.DB(), id, columns)
}

func (s *checkInService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.CheckInRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *checkInService) Delete(id int64) {
	repositories.CheckInRepository.Delete(sqls.DB(), id)
}

func (s *checkInService) CheckIn(userId int64) error {
	today := checkInDay(time.Now())
	var summary model.CheckIn
	err := s.transactionWithRetry(func(tx *gorm.DB) error {
		stats, err := s.lockOrCreateStats(tx, userId)
		if err != nil { return err }
		var existing model.CheckInDay
		if err = tx.Where("user_id = ? AND check_in_date = ?", userId, today).First(&existing).Error; err == nil {
			return errors.New("你已签到")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) { return err }
		now := time.Now()
		if err = tx.Create(&model.CheckInDay{UserId: userId, CheckInDate: today, CheckInType: "normal", CheckInTime: now.UnixMilli(), CreateTime: now.UnixMilli(), UpdateTime: now.UnixMilli()}).Error; err != nil { return err }
		if err = s.refreshStats(tx, stats, userId, now); err != nil { return err }
		summary = *stats
		return nil
	})
	if err != nil { return err }
	s.refreshCachesAndScore(userId, summary.ConsecutiveDays, today)
	return nil
}

func (s *checkInService) transactionWithRetry(fn func(*gorm.DB) error) error {
	var err error
	for attempt := 0; attempt < 2; attempt++ {
		err = sqls.DB().Transaction(fn)
		if err == nil || strings.Contains(strings.ToLower(err.Error()), "已签到") { return err }
	}
	return err
}

func (s *checkInService) lockOrCreateStats(tx *gorm.DB, userId int64) (*model.CheckIn, error) {
	stats := &model.CheckIn{}
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userId).First(stats).Error
	if err == nil { return stats, nil }
	if !errors.Is(err, gorm.ErrRecordNotFound) { return nil, err }
	now := dates.NowTimestamp()
	stats = &model.CheckIn{UserId: userId, MakeupCards: 3, InitialCardsGranted: true, CreateTime: now, UpdateTime: now}
	if err = tx.Create(stats).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") { return s.lockOrCreateStats(tx, userId) }
		return nil, err
	}
	return stats, nil
}

func (s *checkInService) refreshStats(tx *gorm.DB, stats *model.CheckIn, userId int64, now time.Time) error {
	var datesList []int
	if err := tx.Model(&model.CheckInDay{}).Where("user_id = ?", userId).Order("check_in_date DESC").Pluck("check_in_date", &datesList).Error; err != nil { return err }
	stats.TotalCheckInDays = len(datesList)
	stats.ConsecutiveDays = consecutiveFromDates(datesList)
	if len(datesList) > 0 {
		stats.LatestDayName = datesList[0]
		var latest model.CheckInDay
		if err := tx.Where("user_id = ? AND check_in_date = ?", userId, datesList[0]).First(&latest).Error; err != nil { return err }
		stats.LatestCheckInTime = latest.CheckInTime
	}
	newRewarded := stats.TotalCheckInDays / 7
	if newRewarded > stats.RewardedCheckInDays { stats.MakeupCards += newRewarded - stats.RewardedCheckInDays }
	if stats.MakeupCards < 0 { stats.MakeupCards = 0 }
	stats.RewardedCheckInDays = newRewarded
	stats.UpdateTime = now.UnixMilli()
	return tx.Save(stats).Error
}

func consecutiveFromDates(values []int) int {
	if len(values) == 0 { return 0 }
	count := 1
	previous := dayToTime(values[0])
	for _, value := range values[1:] {
		current := dayToTime(value)
		if checkInDay(previous.AddDate(0, 0, -1)) != checkInDay(current) { break }
		count++
		previous = current
	}
	return count
}

func (s *checkInService) refreshCachesAndScore(userId int64, consecutiveDays, dayName int) {
	cache.UserCache.RefreshCheckInRank()
	config := SysConfigService.GetConfig()
	score := calculateCheckInScore(consecutiveDays, config.ScoreConfig)
	if score > 0 { _ = UserService.IncrScore(userId, score, constants.EntityCheckIn, strconv.FormatInt(userId, 10), "签到"+strconv.Itoa(dayName)) } else { logrus.Warn("签到积分未配置...") }
}

func (s *checkInService) InitializeUser(userId int64) error {
	return sqls.DB().Transaction(func(tx *gorm.DB) error {
		stats, err := s.lockOrCreateStats(tx, userId)
		if err != nil { return err }
		if stats.TotalCheckInDays == 0 && stats.LatestDayName > 0 {
			var legacyDay model.CheckInDay
			findErr := tx.Where("user_id = ? AND check_in_date = ?", userId, stats.LatestDayName).First(&legacyDay).Error
			if errors.Is(findErr, gorm.ErrRecordNotFound) {
				stamp := stats.UpdateTime
				if stamp == 0 { stamp = dates.NowTimestamp() }
				if err = tx.Create(&model.CheckInDay{UserId: userId, CheckInDate: stats.LatestDayName, CheckInType: "normal", CheckInTime: stamp, CreateTime: stamp, UpdateTime: stamp}).Error; err != nil { return err }
				stats.TotalCheckInDays = 1
				stats.LatestCheckInTime = stamp
				if stats.ConsecutiveDays <= 0 { stats.ConsecutiveDays = 1 }
			}
		}
		if !stats.InitialCardsGranted {
			stats.MakeupCards += 3
			stats.InitialCardsGranted = true
		}
		return tx.Save(stats).Error
	})
}

func (s *checkInService) InitializeAllUsers() error {
	var users []model.User
	if err := sqls.DB().Find(&users).Error; err != nil { return err }
	for _, user := range users { if err := s.InitializeUser(user.Id); err != nil { return err } }
	return nil
}

type CheckInSummary struct {
	ConsecutiveDays int `json:"consecutiveDays"`
	TotalCheckInDays int `json:"totalCheckInDays"`
	MakeupCards int `json:"makeupCards"`
	CheckedInToday bool `json:"checkedInToday"`
	LatestDayName int `json:"latestDayName"`
}

func (s *checkInService) GetSummary(userId int64, now time.Time) CheckInSummary {
	stats := s.GetByUserId(userId)
	if stats == nil { _ = s.InitializeUser(userId); stats = s.GetByUserId(userId) }
	if stats == nil { return CheckInSummary{} }
	return CheckInSummary{ConsecutiveDays: stats.ConsecutiveDays, TotalCheckInDays: stats.TotalCheckInDays, MakeupCards: stats.MakeupCards, CheckedInToday: stats.LatestDayName == checkInDay(now), LatestDayName: stats.LatestDayName}
}

func (s *checkInService) Makeup(userId int64) (model.CheckInDay, CheckInSummary, error) {
	var picked model.CheckInDay
	var summary CheckInSummary
	now := time.Now()
	err := s.transactionWithRetry(func(tx *gorm.DB) error {
		stats, err := s.lockOrCreateStats(tx, userId)
		if err != nil { return err }
		if stats.MakeupCards <= 0 { return errors.New("补签卡不足") }
		user := UserService.Get(userId)
		if user == nil { return errors.New("用户不存在") }
		registerDay := checkInDay(time.UnixMilli(user.CreateTime))
		for candidate := now.AddDate(0, 0, -1); checkInDay(candidate) >= registerDay; candidate = candidate.AddDate(0, 0, -1) {
			date := checkInDay(candidate)
			var existing model.CheckInDay
			findErr := tx.Where("user_id = ? AND check_in_date = ?", userId, date).First(&existing).Error
			if errors.Is(findErr, gorm.ErrRecordNotFound) {
				picked = model.CheckInDay{UserId: userId, CheckInDate: date, CheckInType: "makeup", CheckInTime: now.UnixMilli(), CreateTime: now.UnixMilli(), UpdateTime: now.UnixMilli()}
				if err = tx.Create(&picked).Error; err != nil { return err }
				stats.MakeupCards--
				if err = s.refreshStats(tx, stats, userId, now); err != nil { return err }
				summary = CheckInSummary{ConsecutiveDays: stats.ConsecutiveDays, TotalCheckInDays: stats.TotalCheckInDays, MakeupCards: stats.MakeupCards, CheckedInToday: stats.LatestDayName == checkInDay(now), LatestDayName: stats.LatestDayName}
				return nil
			} else if findErr != nil { return findErr }
		}
		return errors.New("没有可补签的日期")
	})
	if err == nil { cache.UserCache.RefreshCheckInRank() }
	return picked, summary, err
}

type TodayRankRow struct { Id, UserId int64; CheckInTime int64; CheckInType string }
type ConsecutiveRankRow struct { Id, UserId int64; ConsecutiveDays, LatestCheckInDate int }

func (s *checkInService) TodayRank(now time.Time, page, limit int) ([]TodayRankRow, int64, error) {
	date := checkInDay(now); var rows []TodayRankRow; var total int64
	q := sqls.DB().Table("t_check_in_day d").Joins("JOIN t_user u ON u.id = d.user_id AND u.status = ?", constants.StatusOk).Where("d.check_in_date = ?", date)
	if err := q.Count(&total).Error; err != nil { return nil, 0, err }
	err := q.Select("d.id, d.user_id, d.check_in_time, d.check_in_type").Order("d.check_in_time ASC").Order("d.user_id ASC").Offset((page-1)*limit).Limit(limit).Scan(&rows).Error
	return rows, total, err
}

func (s *checkInService) ConsecutiveRank(page, limit int) ([]ConsecutiveRankRow, int64, error) {
	var rows []ConsecutiveRankRow; var total int64
	q := sqls.DB().Table("t_check_in c").Joins("JOIN t_user u ON u.id = c.user_id AND u.status = ?", constants.StatusOk).Where("c.total_check_in_days > 0")
	if err := q.Count(&total).Error; err != nil { return nil, 0, err }
	rankSize := SysConfigService.GetConfig().ScoreConfig.ConsecutiveRankSize
	if rankSize > 0 && total > int64(rankSize) { total = int64(rankSize) }
	offset := (page - 1) * limit
	if offset >= int(total) { return rows, total, nil }
	if rankSize > 0 && offset+limit > rankSize { limit = rankSize - offset }
	err := q.Select("c.id, c.user_id, c.consecutive_days, c.latest_day_name").Order("c.consecutive_days DESC").Order("c.latest_check_in_time ASC").Order("c.user_id ASC").Offset(offset).Limit(limit).Scan(&rows).Error
	return rows, total, err
}

func (s *checkInService) Overview(userId int64, month string, now time.Time) (map[string]interface{}, error) {
	location := now.Location(); monthTime := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, location)
	if month != "" { parsed, err := time.ParseInLocation("2006-01", month, location); if err != nil { return nil, errors.New("月份格式错误") }; monthTime = parsed }
	if monthTime.After(time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, location)) { return nil, errors.New("不能查看未来月份") }
	stats := s.GetSummary(userId, now); user := UserService.Get(userId); registerDay := 0; if user != nil { registerDay = checkInDay(time.UnixMilli(user.CreateTime)) }
	start := checkInDay(monthTime); endTime := monthTime.AddDate(0, 1, 0).Add(-time.Nanosecond); end := checkInDay(endTime)
	var days []model.CheckInDay
	if err := sqls.DB().Where("user_id = ? AND check_in_date >= ? AND check_in_date <= ?", userId, start, end).Find(&days).Error; err != nil { return nil, err }
	byDate := make(map[int]model.CheckInDay, len(days)); for _, day := range days { byDate[day.CheckInDate] = day }
	calendar := make([]map[string]interface{}, 0, endTime.Day())
	for cursor := monthTime; cursor.Before(monthTime.AddDate(0, 1, 0)); cursor = cursor.AddDate(0, 0, 1) {
		date := checkInDay(cursor)
		status := "missed"
		var kind interface{}
		if date > checkInDay(now) { status = "future" } else if registerDay > 0 && date < registerDay { status = "before_register" } else if item, ok := byDate[date]; ok { kind = item.CheckInType; if item.CheckInType == "makeup" { status = "makeup" } else { status = "checked" } }
		calendar = append(calendar, map[string]interface{}{"date": fmt.Sprintf("%04d-%02d-%02d", cursor.Year(), cursor.Month(), cursor.Day()), "status": status, "type": kind})
	}
	var todayTotal int64; today := checkInDay(now); if err := sqls.DB().Table("t_check_in_day d").Joins("JOIN t_user u ON u.id = d.user_id AND u.status = ?", constants.StatusOk).Where("d.check_in_date = ?", today).Select("COUNT(DISTINCT d.user_id)").Scan(&todayTotal).Error; err != nil { return nil, err }
	return map[string]interface{}{"today": formatDay(today), "todayTotal": todayTotal, "checkedInToday": stats.CheckedInToday, "consecutiveDays": stats.ConsecutiveDays, "totalCheckInDays": stats.TotalCheckInDays, "makeupCards": stats.MakeupCards, "month": monthTime.Format("2006-01"), "calendar": calendar}, nil
}

func checkInDay(value time.Time) int { local := value.In(time.Local); return local.Year()*10000 + int(local.Month())*100 + local.Day() }
func dayToTime(value int) time.Time { year, month, day := value/10000, time.Month((value/100)%100), value%100; return time.Date(year, month, day, 0, 0, 0, 0, time.Local) }
func formatDay(value int) string { return dayToTime(value).Format("2006-01-02") }

// calculateCheckInScore increases the reward with the consecutive day count,
// capped by CheckInScoreMax. CheckInScore remains a fallback for old configs.
func calculateCheckInScore(consecutiveDays int, config model.ScoreConfig) int {
	if consecutiveDays <= 0 {
		return 0
	}
	if config.CheckInScoreMax > 0 {
		if consecutiveDays > config.CheckInScoreMax {
			return config.CheckInScoreMax
		}
		return consecutiveDays
	}
	return config.CheckInScore
}

func (s *checkInService) GetByUserId(userId int64) *model.CheckIn {
	return s.FindOne(sqls.NewCnd().Eq("user_id", userId))
}
