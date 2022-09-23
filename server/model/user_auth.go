package model

type AuthUnit struct {
	// Attribute
	BelongSection int // 权限单元所属板块：[0]不属于任何板块；[n]表示对应板块
	// Auth
	ReadLv   int // 帖子阅读等级：[n]表示阅读权重，仅能阅读权限等级不高于该项的帖子，若为 0 表示部门所有帖子都可以查看，若为 -1 表示所有帖子都可以查看
	ManageLv int // 帖子管理等级：[0]无法管理帖子；[1]能够管理自己帖子；[2]能够管理对应板块所有帖子；[3]能够管理所有帖子；
}

// 定义板块的id
var (
	GLOBAL_ADMIN = 0 // 高管和站长专属
	SAMPLE_PLATE = 1
)

// 实习成员
func InternUserAuthUnit(section int) AuthUnit {
	ret := AuthUnit{
		BelongSection: section,
		ReadLv:        1,
		ManageLv:      1,
	}
	return ret
}

// 正式成员
func NormalUserAuthUnit(section int) AuthUnit {
	ret := AuthUnit{
		BelongSection: section,
		ReadLv:        2,
		ManageLv:      1,
	}
	return ret
}

// 顾问 / 高级成员
func SeniorUserAuthUnit(section int) AuthUnit {
	ret := AuthUnit{
		BelongSection: section,
		ReadLv:        3,
		ManageLv:      1,
	}
	return ret
}

// 中管及项目经理
func AdminUserAuthUnit(section int) AuthUnit {
	ret := AuthUnit{
		BelongSection: section,
		ReadLv:        0,
		ManageLv:      2,
	}
	return ret
}

// 高管和站长
func MasterUserAuthUnit() AuthUnit {
	ret := AuthUnit{
		BelongSection: GLOBAL_ADMIN,
		ReadLv:        -1,
		ManageLv:      3,
	}
	return ret
}
