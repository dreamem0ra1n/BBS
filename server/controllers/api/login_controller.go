package api

import (
	"bbs-go/controllers/render"
	"bbs-go/model"
	"bbs-go/model/constants"
	"bbs-go/repositories"
	"encoding/json"
	"io/ioutil"
	"net/http"
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

type LoginController struct {
	Ctx iris.Context
}

// No need to use
// 注册
// func (c *LoginController) PostSignup() *web.JsonResult {
// 	var (
// 		captchaId   = c.Ctx.PostValueTrim("captchaId")
// 		captchaCode = c.Ctx.PostValueTrim("captchaCode")
// 		email       = c.Ctx.PostValueTrim("email")
// 		username    = c.Ctx.PostValueTrim("username")
// 		password    = c.Ctx.PostValueTrim("password")
// 		rePassword  = c.Ctx.PostValueTrim("rePassword")
// 		nickname    = c.Ctx.PostValueTrim("nickname")
// 		ref         = c.Ctx.FormValue("ref")
// 	)
// 	loginMethod := services.SysConfigService.GetLoginMethod()
// 	if !loginMethod.Password {
// 		return web.JsonErrorMsg("账号密码登录/注册已禁用")
// 	}
// 	if !captcha.VerifyString(captchaId, captchaCode) {
// 		return web.JsonError(errs.CaptchaError)
// 	}
// 	user, err := services.UserService.SignUp(username, email, nickname, password, rePassword)
// 	if err != nil {
// 		return web.JsonError(err)
// 	}
// 	return render.BuildLoginSuccess(user, ref)
// }

// 用户名密码登录
func (c *LoginController) PostSignin() *web.JsonResult {
	successCookieVal := c.Ctx.GetCookie("SESSION_TOKEN")
	// 跳转前的网址
	ref := c.Ctx.PostValueTrim("ref")

	client := &http.Client{}
	parms := ioutil.NopCloser(strings.NewReader(""))
	req, err := http.NewRequest("GET", "https://www.qsc.zju.edu.cn/passport/v4/profile", parms)
	req.Header.Set("User-Agent", "Golang_Spider_Bot/3.0")

	cookie := &http.Cookie{
		Name:    "SESSION_TOKEN",
		Value:   successCookieVal,
		Expires: time.Now().Add(111 * time.Second),
	}
	req.AddCookie(cookie)
	HTTPresp, err := client.Do(req)
	defer HTTPresp.Body.Close()

	if err != nil {
		logrus.Error("error happen when send request to passport", err)
		return web.JsonError(err)
	}

	body, err := ioutil.ReadAll(HTTPresp.Body)
	if err != nil {
		logrus.Error("error happen when read response from passport", err)
		return web.JsonError(err)
	}
	bodyStr := string(body)

	var resp struct {
		Data struct {
			Logined bool `json:"logined"`
			User    struct {
				Name      string `json:"Name"`
				ZjuId     string `json:"ZjuId"`
				LoginType string `json:"LoginType"`
			} `json:"user"`
		} `json:"Data"`
	}
	err = json.Unmarshal([]byte(bodyStr), &resp)

	username := resp.Data.User.Name
	ZjuId := resp.Data.User.ZjuId
	_ = resp.Data.User.LoginType

	user, err := services.UserService.SignIn(ZjuId, ZjuId)
	if err.Error() == "NO_SUCH_USER" {
		logrus.Info("No such user, try to create a new account.")
		user, err = registeUser(username, ZjuId)
	} else if err != nil {
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

func registeUser(username string, ZJUId string) (*model.User, error) {
	var build strings.Builder
	build.WriteString(ZJUId)
	build.WriteString("@zju.edu.cn")
	email := build.String()

	user := &model.User{
		Username:   sqls.SqlNullString(ZJUId),
		Email:      sqls.SqlNullString(email),
		Nickname:   username,
		Password:   passwd.EncodePassword(ZJUId),
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
