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
		return r == ',' || r == '_'
	})
	if len(roles) == 0 {
		return false
	}
	return arrays.Contains(role, roles)
}

// HasAnyRole 是否有指定的任意角色
func (u *User) HasAnyRole(roles ...string) bool {
	if u == nil || len(roles) == 0 {
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

	if u == nil || len(u.Roles) == 0 {
		return -1, nil // -1
	}
	roles := strings.Split(u.Roles, ",")
	for _, role := range roles {
		roleItems := strings.Split(role, "_")
		roleType := roleItems[0]
		var roleArgv int
		var err error
		if len(roleItems) == 2 {
			roleArgv, err = strconv.Atoi(roleItems[1])
		} else {
			if reqRole == MasterUser_NAME && roleType == reqRole {
				return GLOBAL_ADMIN_SECTION, nil
			}
			roleArgv = -1
		}
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
	if u == nil || len(u.Roles) == 0 {
		return DefaultUser_NAME, nil
	}
	roles := strings.Split(u.Roles, ",")
	for _, role := range roles {
		roleItems := strings.Split(role, "_")
		roleType := roleItems[0]
		var roleArgv int
		var err error
		if len(roleItems) == 2 {
			roleArgv, err = strconv.Atoi(roleItems[1])
		} else {
			roleArgv = -1
		}
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
	if u == nil || len(u.Roles) == 0 {
		return nil, nil
	}
	roles := strings.Split(u.Roles, ",")
	retAu := []*AuthUnit{}
	for _, role := range roles {
		roleItems := strings.Split(role, "_")
		roleType := roleItems[0]
		var roleArgv int
		var err error
		if len(roleItems) == 2 {
			roleArgv, err = strconv.Atoi(roleItems[1])
		} else {
			roleArgv = -1
		}
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
	if u == nil {
		return false
	}
	return u.HasAnyRole(MasterUser_NAME)
}

func (u *User) IsAdminUserOrHigher() bool {
	if _, err := u.GetArgByRole(MasterUser_NAME); err == nil {
		return true
	}
	if _, err := u.GetArgByRole(AdminUser_NAME); err == nil {
		return true
	}
	return false
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

func UserCanAccessTopic(user *User, topic *Topic) bool {
	// 没有用户
	if user == nil {
		return false
	}

	// 是站长就不需要做进一步的部门鉴权
	if user != nil && user.IsMasterUser() {
		return true
	}

	// 获取该部门的权限
	role, err := user.GetRoleByArg(topic.NodeId)
	if err != nil {
		logrus.Error("Error happen while getting user's auth")
		return false
	}
	au, err := GetAuthUnit(role, int(topic.NodeId))
	if err != nil {
		logrus.Error("Error happen while getting user's auth unit")
		return false
	}

	// 检查权限单元
	// logrus.Info(fmt.Sprintf("Role: %s, ReadLv: %d, AccessLv: %d", role, au.ReadLv, topic.AccessLv))

	if au.ReadLv == 0 { // 如果是 0 说明可以无条件访问所有帖子
		return true
	} else if topic.AccessLv == 0 { // 如果是 0 说明只有上面那种人才能访问
		return false
	}
	// 如果阅读权限过低则无法访问
	if au.ReadLv < topic.AccessLv {
		return false
	}
	return true
}

func (user *User) CanManageTopic(topic *Topic) bool {
	// 是站长就不需要做进一步的部门鉴权
	if user.IsMasterUser() {
		return true
	}

	// 获取该部门的权限
	role, err := user.GetRoleByArg(topic.NodeId)
	if err != nil {
		logrus.Error("Error happen while getting user's auth")
		return false
	}
	au, err := GetAuthUnit(role, int(topic.NodeId))
	if err != nil {
		logrus.Error("Error happen while getting user's auth unit")
		return false
	}

	// 检查权限单元
	// 如果有部门以上的管理权限则能够修改
	if au.ManageLv >= 2 {
		return true
	}
	return false
}
