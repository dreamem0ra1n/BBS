package model

import (
	"bbs-go/model/constants"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/mlogclub/simple/common/arrays"
	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/common/strs"
	"github.com/sirupsen/logrus"
)

// IsForbidden 是否禁言
func (u *User) IsForbidden() bool {
	if u.ForbiddenEndTime == 0 {
		return false
	}
	// 永久禁言
	if u.ForbiddenEndTime == -1 {
		return true
	}
	// 判断禁言时间
	return u.ForbiddenEndTime > dates.NowTimestamp()
}

// HasRole 是否有指定角色
func (u *User) HasRole(role string) bool {
	roles := strings.FieldsFunc(u.Roles, func(r rune) bool {
		return r == ',' || r == '.'
	})
	if len(roles) == 0 {
		return false
	}
	return arrays.Contains(role, roles)
}

// HasAnyRole 是否有指定的任意角色
func (u *User) HasAnyRole(roles ...string) bool {
	if len(roles) == 0 {
		return false
	}
	for _, role := range roles {
		if u.HasRole(role) {
			return true
		}
	}
	return false
}

// GetArgByRole 查看对应的 role 的参数
func (u *User) GetArgByRole(reqRole string) (int, error) {
	if len(u.Roles) == 0 {
		return -1, nil
	}
	roles := strings.Split(u.Roles, ",")
	for _, role := range roles {
		roleItems := strings.Split(role, ".")
		roleType := roleItems[0]
		roleArgv, err := strconv.Atoi(roleItems[1])
		if err != nil {
			return -1, err
		}
		if reqRole == roleType {
			return roleArgv, nil
		}
	}
	return -1, errors.New("no such role")
}

// GetRoleByArg 查看对应的 Role
func (u *User) GetRoleByArg(arg int64) (string, error) {
	if len(u.Roles) == 0 {
		return DefaultUser_NAME, nil
	}
	roles := strings.Split(u.Roles, ",")
	for _, role := range roles {
		roleItems := strings.Split(role, ".")
		roleType := roleItems[0]
		roleArgv, err := strconv.Atoi(roleItems[1])
		if err != nil {
			return "", err
		}
		if arg == int64(roleArgv) {
			return roleType, nil
		}
	}
	return DefaultUser_NAME, nil
}

// GetUserAuthUnit 根据权限类型获取用户对应的权限单元内容
func (u *User) GetUserAuthUnitByRole(role string) (*AuthUnit, error) {
	if role == DefaultUser_NAME {
		return GetAuthUnit(DefaultUser_NAME, -1)
	}
	arg, err := u.GetArgByRole(role)
	if err != nil {
		return nil, err
	}
	au, err := GetAuthUnit(role, arg)
	return au, err
}

// GetUserAuthUnits 获取用户权限单元列表
func (u *User) GetUserAuthUnits() ([]*AuthUnit, error) {
	if len(u.Roles) == 0 {
		return nil, nil
	}
	roles := strings.Split(u.Roles, ",")
	retAu := []*AuthUnit{}
	for _, role := range roles {
		roleItems := strings.Split(role, ".")
		roleType := roleItems[0]
		roleArgv, err := strconv.Atoi(roleItems[1])
		if err != nil {
			logrus.Error("Error happen when split %s's roles!", u.Nickname)
			return nil, err
		}
		au, err := GetAuthUnit(roleType, roleArgv)
		if err != nil {
			logrus.Error("Role %s not found!", roleType)
			return nil, err
		}
		retAu = append(retAu, au)
	}
	return retAu, errors.New("no such role")
}

func (u *User) IsMasterUser() bool {
	if _, err := u.GetArgByRole(MasterUser_NAME); err != nil {
		return false
	}
	return true
}

// GetRoles 获取角色
func (u *User) GetRoles() []string {
	if strs.IsBlank(u.Roles) {
		return nil
	}
	ss := strings.Split(u.Roles, ",")
	if len(ss) == 0 {
		return nil
	}
	var roles []string
	for _, s := range ss {
		s = strings.TrimSpace(s)
		if strs.IsNotBlank(s) {
			roles = append(roles, s)
		}
	}
	return roles
}

// InObservationPeriod 是否在观察期
// observeSeconds 观察时长
func (u *User) InObservationPeriod(observeSeconds int) bool {
	if observeSeconds <= 0 {
		return false
	}
	return dates.FromTimestamp(u.CreateTime).Add(time.Second * time.Duration(observeSeconds)).After(time.Now())
}

// GetTitle 获取帖子的标题
func (t *Topic) GetTitle() string {
	if t.Type == constants.TopicTypeTweet {
		if strs.IsNotBlank(t.Content) {
			return t.Content
		} else {
			return "分享图片"
		}
	} else {
		return t.Title
	}
}
