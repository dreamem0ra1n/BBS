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

// 用户名密码登录
func (c *LoginController) PostSignin() *web.JsonResult {
	successCookieVal := c.Ctx.GetCookie("SESSION_TOKEN")
	// 跳转前的网址
	ref := c.Ctx.PostValueTrim("ref")

	client := &http.Client{}
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
	logrus.Info(cookie)
	req.AddCookie(cookie)
	HTTPresp, err := client.Do(req)
	if err != nil {
		logrus.Error("error happen when send request to passport", err)
		return web.JsonError(err)
	}

	body, err := ioutil.ReadAll(HTTPresp.Body)
	HTTPresp.Body.Close()
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

	// logrus.Info("receive data from passport(string): ", bodyStr)
	logrus.Info("receive data from passport(binding): ", resp.Data)

	if resp.Data.User.LoginType != "qsc" {
		logrus.Info("Can't login! Actually, he/she is not a qscer!")
		return web.JsonError(errors.New("you are not qscer"))
	}

	username := resp.Data.User.Qsc.QscId
	ZjuId := resp.Data.User.ZjuId
	_ = resp.Data.User.LoginType

	user, err := services.UserService.SignIn(username, ZjuId)
	if err != nil && err.Error() == "NO_SUCH_USER" {
		logrus.Info("No such user, try to create a new account.")
		user, err = registerUser(resp.Data)
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

func getRoleFromLoginUserData(u LoginUser) string {
	var ret string
	switch u.User.Qsc.Position {
	case "实习成员":
		ret = model.InternUser_NAME
	case "正式成员":
		ret = model.NormalUser_NAME
	case "顾问", "高级成员":
		ret = model.SeniorUser_NAME
	case "中管":
		ret = "中管_1,中管_11," + model.AdminUser_NAME // 添加公告权限
	case "高管":
		return model.MasterUser_NAME // 添加公告权限
	default:
		return ""
	}

	ret += "_"

	switch u.User.Qsc.Department {
	case "产品研发中心": // 产研合并
		return ret + strconv.Itoa(model.ChanPinYanFa_SECTION) + "," + ret + strconv.Itoa(model.JiShuYanFa_SECTION) + "," + ret + strconv.Itoa(model.ChanPinYunYing_SECTION)
	case "技术研发中心":
		return ret + strconv.Itoa(model.ChanPinYanFa_SECTION) + "," + ret + strconv.Itoa(model.JiShuYanFa_SECTION)
	case "产品运营部门":
		return ret + strconv.Itoa(model.ChanPinYanFa_SECTION) + "," + ret + strconv.Itoa(model.ChanPinYunYing_SECTION)
	case "推广策划中心":
		return ret + strconv.Itoa(model.TuiGuangCeHua_SECTION)
	case "新闻资讯中心":
		return ret + strconv.Itoa(model.XinWenZiXun_SECTION)
	case "设计与视觉中心":
		return ret + strconv.Itoa(model.SheJiYuShiJue_SECTION)
	case "人力资源部门":
		return ret + strconv.Itoa(model.RenLiZiYuan_SECTION)
	case "摄影部":
		return ret + strconv.Itoa(model.SheYing_SECTION)
	case "视频":
		return ret + strconv.Itoa(model.ShiPin_SECTION)
	default:
		return ""
	}
}
