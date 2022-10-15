package model

import "errors"

type AuthUnit struct {
	// Attribute
	BelongSection int // 权限单元所属板块：[0]不属于任何板块；[n]表示对应板块
	// Auth
	ReadLv   int // 帖子阅读等级：[n]表示阅读权重，仅能阅读权限等级不高于该项的帖子，若为 0 表示部门所有帖子都可以查看，若为 -1 表示所有帖子都可以查看
	ManageLv int // 帖子管理等级：[0]无法管理帖子；[1]能够管理自己帖子；[2]能够管理对应板块所有帖子；[3]能够管理所有帖子；
}

// 定义板块的id
const (
	Default_SECTION        = -1 // 默认权限对应板块id
	GLOBAL_ADMIN_SECTION   = 0  // 高管和站长专属
	Announcement_SECTION   = 1  // 公告
	ChanPinYanFa_SECTION   = 2  // 产品研发
	JiShuYanFa_SECTION     = 3  // 技术研发
	ChanPinYunYing_SECTION = 4  // 产品运营
	TuiGuangCeHua_SECTION  = 5  // 推广策划
	XinWenZiXun_SECTION    = 6  // 新闻资讯
	SheJiYuShiJue_SECTION  = 7  // 设计与视觉
	RenLiZiYuan_SECTION    = 8  // 人力资源部门
	SheYing_SECTION        = 9  // 摄影
	ShiPin_SECTION         = 10 // 视频
	Water_SECTIOIN         = 11 // 全潮水板
)

// 定义权限名称
const (
	DefaultUser_NAME = "默认"
	MasterUser_NAME  = "高管"
	AdminUser_NAME   = "中管"
	SeniorUser_NAME  = "高级成员"
	NormalUser_NAME  = "正式成员"
	InternUser_NAME  = "实习成员"
)

func GetAuthUnit(role string, arg int) (*AuthUnit, error) {
	switch role {
	case MasterUser_NAME:
		return MasterUserAuthUnit(), nil
	case AdminUser_NAME:
		return AdminUserAuthUnit(arg), nil
	case SeniorUser_NAME:
		return SeniorUserAuthUnit(arg), nil
	case NormalUser_NAME:
		return NormalUserAuthUnit(arg), nil
	case InternUser_NAME:
		return InternUserAuthUnit(arg), nil
	case DefaultUser_NAME:
		return DefaultAuthUnit(), nil
	default:
		return nil, errors.New("no such role")
	}
}

// 默认权限
func DefaultAuthUnit() *AuthUnit {
	ret := &AuthUnit{
		BelongSection: Default_SECTION,
		ReadLv:        1,
		ManageLv:      1,
	}
	return ret
}

// 实习成员
func InternUserAuthUnit(section int) *AuthUnit {
	ret := &AuthUnit{
		BelongSection: section,
		ReadLv:        1,
		ManageLv:      1,
	}
	return ret
}

// 正式成员
func NormalUserAuthUnit(section int) *AuthUnit {
	ret := &AuthUnit{
		BelongSection: section,
		ReadLv:        2,
		ManageLv:      1,
	}
	return ret
}

// 顾问 / 高级成员
func SeniorUserAuthUnit(section int) *AuthUnit {
	ret := &AuthUnit{
		BelongSection: section,
		ReadLv:        3,
		ManageLv:      1,
	}
	return ret
}

// 中管及项目经理
func AdminUserAuthUnit(section int) *AuthUnit {
	ret := &AuthUnit{
		BelongSection: section,
		ReadLv:        0,
		ManageLv:      2,
	}
	return ret
}

// 高管和站长
func MasterUserAuthUnit() *AuthUnit {
	ret := &AuthUnit{
		BelongSection: GLOBAL_ADMIN_SECTION,
		ReadLv:        -1,
		ManageLv:      3,
	}
	return ret
}
