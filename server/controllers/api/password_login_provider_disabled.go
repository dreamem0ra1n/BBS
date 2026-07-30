//go:build !passwordlogin
// +build !passwordlogin

package api

import "github.com/kataras/iris/v12/mvc"

func registerPasswordLoginProvider(_ *mvc.Application) {}
