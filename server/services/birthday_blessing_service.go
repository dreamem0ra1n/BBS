package services

import (
	"bbs-go/model"
	"bbs-go/repositories"
	"math/rand"
	"strings"
	"time"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var BirthdayBlessingService = newBirthdayBlessingService()

type birthdayBlessingService struct{}

func newBirthdayBlessingService() *birthdayBlessingService { return &birthdayBlessingService{} }
func (s *birthdayBlessingService) Get(id int64) *model.BirthdayBlessing {
	return repositories.BirthdayBlessingRepository.Get(sqls.DB(), id)
}
func (s *birthdayBlessingService) ExistsByNickname(nickname string) bool {
	nickname = strings.TrimSpace(nickname)
	return nickname != "" && repositories.BirthdayBlessingRepository.ExistsByNickname(sqls.DB(), nickname)
}
func (s *birthdayBlessingService) FindPageByParams(p *params.QueryParams) ([]model.BirthdayBlessing, *sqls.Paging) {
	list, paging := repositories.BirthdayBlessingRepository.FindPageByParams(sqls.DB(), p)
	return list, paging
}
func (s *birthdayBlessingService) Create(item *model.BirthdayBlessing) error {
	return repositories.BirthdayBlessingRepository.Create(sqls.DB(), item)
}
func (s *birthdayBlessingService) Delete(id int64) error {
	return repositories.BirthdayBlessingRepository.Delete(sqls.DB(), id)
}

func (s *birthdayBlessingService) Random(department string) *model.BirthdayBlessing {
	items := repositories.BirthdayBlessingRepository.Find(sqls.DB(), sqls.NewCnd())
	if len(items) == 0 {
		return nil
	}
	return &items[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(items))]
}

func (s *birthdayBlessingService) RandomForUser(userId int64, department string, preferSameDepartment bool) *model.BirthdayBlessing {
	items := repositories.BirthdayBlessingRepository.Find(sqls.DB(), sqls.NewCnd())
	if len(items) == 0 {
		return nil
	}
	var history []model.BirthdayBlessingHistory
	sqls.DB().Where("user_id = ?", userId).Find(&history)
	used := make(map[int64]bool, len(history))
	for _, item := range history {
		used[item.BlessingId] = true
	}
	available := make([]model.BirthdayBlessing, 0, len(items))
	for _, item := range items {
		if !used[item.Id] {
			available = append(available, item)
		}
	}
	if len(available) == 0 {
		return nil
	}
	normalizedDepartment := strings.TrimSpace(department)
	if preferSameDepartment && normalizedDepartment != "" {
		departmentItems := make([]model.BirthdayBlessing, 0, len(available))
		for _, item := range available {
			if strings.TrimSpace(item.Department) == normalizedDepartment {
				departmentItems = append(departmentItems, item)
			}
		}
		if len(departmentItems) > 0 {
			available = departmentItems
		}
	}
	return &available[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(available))]
}

func (s *birthdayBlessingService) Normalize(item *model.BirthdayBlessing) bool {
	item.Nickname = strings.TrimSpace(item.Nickname)
	item.Department = strings.TrimSpace(item.Department)
	item.Content = strings.TrimSpace(item.Content)
	return item.Content != ""
}
