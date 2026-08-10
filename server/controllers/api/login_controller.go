package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/common/passwd"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web"
	"github.com/sirupsen/logrus"

	"bbs-go/controllers/render"
	"bbs-go/model"
	"bbs-go/model/constants"
	"bbs-go/pkg/authproviders"
	"bbs-go/pkg/config"
	"bbs-go/repositories"
	"bbs-go/services"
)

// =============================================================================
// 登录方式注册 —— 各登录 Provider 的统一注册入口
// =============================================================================

// RegisterLoginProviders 将所有登录方式注册到 Iris MVC App 中。
// Passport 由运行时配置控制；密码登录还必须在编译时启用 passwordlogin build tag。
func RegisterLoginProviders(app *mvc.Application) {
	app.Handle(new(LoginController))
	registerPassportLoginProvider(app)
	registerPasswordLoginProvider(app)
}

// registerPassportLoginProvider 注册求是潮 Passport 登录路由。
func registerPassportLoginProvider(app *mvc.Application) {
	if config.Instance.LoginMethods.Passport {
		app.Handle(new(PassportLoginController))
	}
}

// registerPasswordLoginProvider 注册用户名/密码登录路由。
func registerPasswordLoginProvider(app *mvc.Application) {
	// 双重检查可避免生产镜像仅因配置误开 password 就暴露密码登录接口。
	if config.Instance.LoginMethods.Password && authproviders.PasswordCompiled() {
		app.Handle(new(PasswordLoginController))
	}
}

// =============================================================================
// Passport 登录 —— 求是潮 OAuth / 单点登录
// =============================================================================

// LoginUser 是 Passport 接口返回的用户信息结构。
type LoginUser struct {
	Logined bool `json:"logined"`
	User    struct {
		Name      string `json:"Name"`
		ZjuId     string `json:"ZjuId"`
		LoginType string `json:"LoginType"`
		Qsc       struct {
			// Passport 使用小写 qscid；显式声明，避免依赖 encoding/json 的大小写兼容匹配。
			QscId      string `json:"qscid"`
			Gender     string `json:"gender"`
			Position   string `json:"position"` // 职位
			Department string `json:"department"`
			Status     string `json:"status"`
		} `json:"QscUser"`
	} `json:"user"`
}

// PassportLoginController 处理求是潮 Passport 回调登录。
type PassportLoginController struct {
	Ctx iris.Context
}

// passportClient 复用底层连接；超时用于限制 Passport 异常时占用请求协程的时长。
var passportClient = &http.Client{
	Timeout: 10 * time.Second,
}

// passportProfileURL 返回当前 SESSION_TOKEN 对应的 Passport 用户资料。
const passportProfileURL = "https://www.qsc.zju.edu.cn/passport/v4/profile"

// 对外统一返回非成员错误，避免泄露 Passport 响应中的具体认证信息。
var passportNotQscError = errors.New("you are not qscer")

// PostSignin 处理 Passport 登录回调，完成用户认证或自动注册。
func (c *PassportLoginController) PostSignin() *web.JsonResult {
	// 登录成功后跳转的目标地址
	ref := c.Ctx.PostValueTrim("ref")

	// SESSION_TOKEN 是 Passport 会话凭证，只转发给 Passport，不写入日志。
	successCookieVal := c.Ctx.GetCookie("SESSION_TOKEN")

	// 绑定当前请求上下文；客户端断开或服务端取消时，Passport 请求会同步终止。
	req, err := http.NewRequestWithContext(c.Ctx.Request().Context(), http.MethodGet, passportProfileURL, nil)

	if err != nil {
		logrus.Error("error happen when request passport!", err)
		return web.JsonError(err)
	}

	req.Header.Set("User-Agent", "Golang_Spider_Bot/3.0")

	if successCookieVal != "" {
		req.AddCookie(&http.Cookie{Name: "SESSION_TOKEN", Value: successCookieVal})
	}
	passportResp, err := passportClient.Do(req)
	if err != nil {
		logrus.Error("error happen when send request to passport", err)
		return web.JsonError(err)
	}
	defer passportResp.Body.Close()

	// 先校验状态码，防止把 Passport 的错误页当成正常资料解析。
	if passportResp.StatusCode < http.StatusOK || passportResp.StatusCode >= http.StatusMultipleChoices {
		logrus.Errorf("passport returned unexpected HTTP status: %s", passportResp.Status)
		return web.JsonError(errors.New("passport request failed"))
	}

	var resp struct {
		Data LoginUser `json:"data"`
	}
	// 用户资料很小；限制读取量，避免异常上游返回无限或超大响应体。
	decoder := json.NewDecoder(io.LimitReader(passportResp.Body, 1<<20))
	err = decoder.Decode(&resp)
	if err != nil {
		logrus.Error("error happen when unmarshal the resp: ", err)
		return web.JsonError(err)
	}

	// 鉴权、查询和持久化前统一清理字段，避免空白字符造成重复账号或角色匹配失败。
	resp.Data = normalizeLoginUser(resp.Data)
	if resp.Data.User.LoginType != "qsc" || resp.Data.User.ZjuId == "" || resp.Data.User.Qsc.QscId == "" {
		logrus.Info("Can't login! Actually, he/she is not a qscer!")
		return web.JsonError(passportNotQscError)
	}

	// 先按 ZjuId 邮箱查询，再按 QscId 用户名兜底。这样 ZjuId 变更时仍能定位原账号，
	// 避免直接创建账号后触发 QscId 的唯一索引冲突。
	user, err := services.UserService.SignIn(resp.Data.User.ZjuId+"@zju.edu.cn", resp.Data.User.Qsc.QscId)
	if errors.Is(err, services.ErrNoSuchUser) {
		logrus.Info("No such user, try to create a new account.")
		user, err = registerUser(resp.Data)
	} else if err == nil && user != nil {
		err = updateUser(*user, resp.Data)
	}
	if err != nil {
		return web.JsonError(err)
	}

	return render.BuildLoginSuccess(user, ref)
}

// normalizeLoginUser 清理 Passport 可能携带的首尾空白，并返回可直接入库的数据副本。
func normalizeLoginUser(u LoginUser) LoginUser {
	u.User.Name = strings.TrimSpace(u.User.Name)
	u.User.ZjuId = strings.TrimSpace(u.User.ZjuId)
	u.User.LoginType = strings.TrimSpace(u.User.LoginType)
	u.User.Qsc.QscId = strings.TrimSpace(u.User.Qsc.QscId)
	u.User.Qsc.Gender = strings.TrimSpace(u.User.Qsc.Gender)
	u.User.Qsc.Position = strings.TrimSpace(u.User.Qsc.Position)
	u.User.Qsc.Department = strings.TrimSpace(u.User.Qsc.Department)
	u.User.Qsc.Status = strings.TrimSpace(u.User.Qsc.Status)
	return u
}

// =============================================================================
// 密码登录（仅本地开发使用）
// =============================================================================

// passwordLoginError 是密码登录相关的通用错误。
var passwordLoginError = errors.New("用户名或密码错误")

// PasswordLoginController 处理用户名/密码登录。
type PasswordLoginController struct {
	Ctx iris.Context
}

// PostPassword 处理用户名+密码的表单登录请求。
func (controller *PasswordLoginController) PostPassword() *web.JsonResult {
	username := strings.TrimSpace(controller.Ctx.PostValue("username"))
	password := controller.Ctx.PostValue("password")
	ref := controller.Ctx.PostValueTrim("ref")
	if username == "" || password == "" {
		return web.JsonError(passwordLoginError)
	}

	user := services.UserService.GetByUsername(username)
	// 用户不存在、被禁用和密码错误均返回相同信息，避免通过接口枚举有效账号。
	if user == nil || user.Status != constants.StatusOk || !passwd.ValidatePassword(user.Password, password) {
		return web.JsonError(passwordLoginError)
	}
	return render.BuildLoginSuccess(user, ref)
}

// =============================================================================
// 通用登录控制器 —— 登出等
// =============================================================================

// LoginController 处理与登录方式无关的通用操作（如登出）。
type LoginController struct {
	Ctx iris.Context
}

// GetSignout 处理用户登出。
func (c *LoginController) GetSignout() *web.JsonResult {
	err := services.UserTokenService.Signout(c.Ctx)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

// =============================================================================
// 辅助函数 —— 用户注册 / 信息同步 / 角色映射
// =============================================================================

// registerUser 根据 Passport 返回的用户数据创建本地账号。
func registerUser(u LoginUser) (*model.User, error) {
	email := u.User.ZjuId + "@zju.edu.cn"

	// 历史逻辑以 ZjuId 作为初始本地密码；EncodePassword 使用 bcrypt 哈希后再存库。
	// 生产构建不会注册密码登录路由，但这里仍保留兼容既有账号的数据格式。
	user := &model.User{
		Username:   sqls.SqlNullString(u.User.Qsc.QscId),
		Email:      sqls.SqlNullString(email),
		Nickname:   u.User.Qsc.QscId,
		Password:   passwd.EncodePassword(u.User.ZjuId),
		Realname:   u.User.Name,
		Department: u.User.Qsc.Department,
		Roles:      getRoleFromLoginUserData(u),
		Status:     constants.StatusOk,
		CreateTime: dates.NowTimestamp(),
		UpdateTime: dates.NowTimestamp(),
	}
	err := repositories.UserRepository.Create(sqls.DB(), user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// updateUser 用 Passport 的最新数据同步更新本地用户信息。
// 不更新密码
func updateUser(user model.User, upd LoginUser) error {
	username := upd.User.Qsc.QscId
	email := upd.User.ZjuId + "@zju.edu.cn"
	nickname := upd.User.Qsc.QscId
	department := upd.User.Qsc.Department
	roles := getRoleFromLoginUserData(upd)

	// 仅同步 Passport 负责维护的字段，不覆盖用户自行维护的密码和个人资料。
	// 使用按列更新也能避免 Save 整个用户对象时覆盖其他请求刚写入的数据。
	updates := make(map[string]interface{}, 6)
	if !user.Username.Valid || user.Username.String != username {
		updates["username"] = username
	}
	if !user.Email.Valid || user.Email.String != email {
		updates["email"] = email
	}
	if user.Nickname != nickname {
		updates["nickname"] = nickname
	}
	if user.Department != department {
		updates["department"] = department
	}
	if user.Roles != roles {
		updates["roles"] = roles
	}

	// Passport 资料没有变化时不写数据库，也不触发用户缓存失效。
	if len(updates) == 0 {
		return nil
	}
	updates["update_time"] = dates.NowTimestamp()
	return services.UserService.Updates(user.Id, updates)
}

// getRoleFromLoginUserData 根据 Passport 中的职位和部门信息映射为系统角色字符串。
func getRoleFromLoginUserData(u LoginUser) string {
	var ret string
	prefix := ""
	switch u.User.Qsc.Position {
	case "实习成员":
		ret = model.InternUser_NAME
	case "正式成员":
		prefix = model.NormalUser_NAME + "_1," + model.NormalUser_NAME + "_11,"
		ret = model.NormalUser_NAME
	case "顾问", "高级成员":
		prefix = model.SeniorUser_NAME + "_1," + model.SeniorUser_NAME + "_11,"
		ret = model.SeniorUser_NAME
	case "退休老干部":
		prefix = model.RetireUser_NAME + "_1," + model.RetireUser_NAME + "_11,"
		ret = model.RetireUser_NAME
	case "中管":
		prefix = model.AdminUser_NAME + "_1," + model.AdminUser_NAME + "_11," + model.OLDBBSUser_NAME + ","
		ret = model.AdminUser_NAME
	case "高管":
		return model.MasterUser_NAME + "," + model.OLDBBSUser_NAME
	default:
		return ""
	}

	ret += "_"

	switch u.User.Qsc.Department {
	case "产品研发中心": // 产研合并
		return prefix + ret + strconv.Itoa(model.ChanPinYanFa_SECTION) + "," + ret + strconv.Itoa(model.JiShuYanFa_SECTION) + "," + ret + strconv.Itoa(model.ChanPinYunYing_SECTION)
	case "技术研发中心":
		return prefix + ret + strconv.Itoa(model.ChanPinYanFa_SECTION) + "," + ret + strconv.Itoa(model.JiShuYanFa_SECTION)
	case "产品运营部门":
		return prefix + ret + strconv.Itoa(model.ChanPinYanFa_SECTION) + "," + ret + strconv.Itoa(model.ChanPinYunYing_SECTION)
	case "推广策划中心":
		return prefix + ret + strconv.Itoa(model.TuiGuangCeHua_SECTION)
	case "新闻资讯中心":
		return prefix + ret + strconv.Itoa(model.XinWenZiXun_SECTION)
	case "设计与视觉中心":
		return prefix + ret + strconv.Itoa(model.SheJiYuShiJue_SECTION)
	case "人力资源部门":
		return prefix + ret + strconv.Itoa(model.RenLiZiYuan_SECTION)
	case "摄影部":
		return prefix + ret + strconv.Itoa(model.SheYing_SECTION)
	case "视频团队":
		return prefix + ret + strconv.Itoa(model.ShiPin_SECTION)
	default:
		return ""
	}
}
