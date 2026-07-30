package api

import (
	"bbs-go/pkg/config"

	"github.com/kataras/iris/v12/mvc"
)

func RegisterLoginProviders(login *mvc.Application) {
	login.Handle(new(LoginController))
	if config.Instance.LoginMethods.Passport { // passport登录
		login.Handle(new(PassportLoginController))
	}
	registerPasswordLoginProvider(login) // 密码登录
}
