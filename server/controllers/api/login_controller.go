package api

import (
	"bbs-go/controllers/render"
	"bbs-go/model"
	"bbs-go/model/constants"
	"bbs-go/repositories"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/common/passwd"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web"
	"github.com/sirupsen/logrus"

	"bbs-go/services"
)

type LoginUser struct {
	Logined bool `json:"logined"`
	User    struct {
		Name      string `json:"Name"`
		ZjuId     string `json:"ZjuId"`
		LoginType string `json:"LoginType"`
		Qsc       struct {
			QscId      string `jsone:"qscid"`
			Gender     string `json:"gender"`
			Position   string `json:"position"` // 职位
			Department string `json:"department"`
			Status     string `json:"status"`
		} `json:"QscUser"`
	} `json:"user"`
}

type LoginController struct {
	Ctx iris.Context
}

// Shared HTTP client for passport requests to avoid leaking connections.
// Creating a new http.Client per request leaks TCP connections and file descriptors.
var passportClient = &http.Client{
	Timeout: 10 * time.Second,
}

// 用户名密码登录
func (c *LoginController) PostSignin() *web.JsonResult {
	successCookieVal := c.Ctx.GetCookie("SESSION_TOKEN")
	// 跳转前的网址
	ref := c.Ctx.PostValueTrim("ref")

	parms := ioutil.NopCloser(strings.NewReader(""))
	req, err := http.NewRequest("GET", "https://www.qsc.zju.edu.cn/passport/v4/profile", parms)

	if err != nil {
		logrus.Error("error happen when request passport!", err)
		return web.JsonError(err)
	}

	req.Header.Set("User-Agent", "Golang_Spider_Bot/3.0")

	cookie := &http.Cookie{
		Name:    "SESSION_TOKEN",
		Value:   successCookieVal,
		Expires: time.Now().Add(111 * time.Second),
	}
	req.AddCookie(cookie)
	HTTPresp, err := passportClient.Do(req)
	if err != nil {
		logrus.Error("error happen when send request to passport", err)
		return web.JsonError(err)
	}
	defer HTTPresp.Body.Close()

	body, err := ioutil.ReadAll(HTTPresp.Body)
	if err != nil {
		logrus.Error("error happen when read response from passport", err)
		return web.JsonError(err)
	}
	bodyStr := string(body)

	var resp struct {
		Data LoginUser `json:"data"`
	}
	err = json.Unmarshal([]byte(bodyStr), &resp)
	if err != nil {
		logrus.Error("error happen when unmarshal the resp: ", err)
		return web.JsonError(err)
	}

	if resp.Data.User.LoginType != "qsc" || resp.Data.User.Qsc.QscId == "" {
		logrus.Info("Can't login! Actually, he/she is not a qscer!")
		return web.JsonError(errors.New("you are not qscer"))
	}

	// BUGFIX: 导入的字段有可能前后有空格（应该已经修复）；给字段Trim一下

	ZjuId := strings.TrimSpace(resp.Data.User.ZjuId)
	QscId := strings.TrimSpace(resp.Data.User.Qsc.QscId)

	// BUGFIX: 当ZjuID (email) 修改后，QscID (username) 的Unique约束会导致注册失败；改为查询ZjuId OR QscId

	user, err := services.UserService.SignIn(ZjuId+"@zju.edu.cn", QscId)
	if err != nil && err.Error() == "NO_SUCH_USER" {
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

// 退出登录
func (c *LoginController) GetSignout() *web.JsonResult {
	err := services.UserTokenService.Signout(c.Ctx)
	if err != nil {
		return web.JsonError(err)
	}
	return web.JsonSuccess()
}

// 注册
func registerUser(u LoginUser) (*model.User, error) {
	email := u.User.ZjuId + "@zju.edu.cn"

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

// 从 passport 更新信息
func updateUser(user model.User, upd LoginUser) error {
	email := upd.User.ZjuId + "@zju.edu.cn"

	user.Username = sqls.SqlNullString(upd.User.Qsc.QscId)
	user.Nickname = upd.User.Qsc.QscId
	user.Password = upd.User.ZjuId
	user.Email = sqls.SqlNullString(email)
	user.Password = passwd.EncodePassword(upd.User.ZjuId)
	user.Roles = getRoleFromLoginUserData(upd)

	err := services.UserService.Update(&user)

	if err != nil {
		return err
	}
	return nil
}

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
