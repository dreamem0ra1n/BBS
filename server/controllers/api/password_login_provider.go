//go:build passwordlogin
// +build passwordlogin

package api

import (
	"errors"
	"strings"

	"bbs-go/controllers/render"
	"bbs-go/model/constants"
	"bbs-go/pkg/config"
	"bbs-go/services"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/mlogclub/simple/common/passwd"
	"github.com/mlogclub/simple/web"
)

var passwordLoginError = errors.New("用户名或密码错误")

type PasswordLoginController struct {
	Ctx iris.Context
}

func registerPasswordLoginProvider(login *mvc.Application) {
	if config.Instance.LoginMethods.Password {
		login.Handle(new(PasswordLoginController))
	}
}

func (controller *PasswordLoginController) PostPassword() *web.JsonResult {
	username := strings.TrimSpace(controller.Ctx.PostValue("username"))
	password := controller.Ctx.PostValue("password")
	ref := controller.Ctx.PostValueTrim("ref")
	if username == "" || password == "" {
		return web.JsonError(passwordLoginError)
	}

	user := services.UserService.GetByUsername(username)
	if user == nil || user.Status != constants.StatusOk || !passwd.ValidatePassword(user.Password, password) {
		return web.JsonError(passwordLoginError)
	}
	return render.BuildLoginSuccess(user, ref)
}
