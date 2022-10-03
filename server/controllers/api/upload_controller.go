package api

import (
	"github.com/kataras/iris/v12"
)

type UploadController struct {
	Ctx iris.Context
}

// func (c *UploadController) Post() *web.JsonResult {
// 	user := services.UserTokenService.GetCurrent(c.Ctx)
// 	if err := services.UserService.CheckPostStatus(user); err != nil {
// 		return web.JsonError(err)
// 	}

// 	file, header, err := c.Ctx.FormFile("image")
// 	if err != nil {
// 		return web.JsonError(err)
// 	}
// 	defer file.Close()

// 	if header.Size > constants.UploadMaxBytes {
// 		return web.JsonErrorMsg("图片不能超过" + strconv.Itoa(constants.UploadMaxM) + "M")
// 	}

// 	record, err := putFile(header.Filename, file, header.Size)
// 	if err != nil {
// 		return web.JsonError(err)
// 	}

// 	host := config.Instance.BaseUrl
// 	url := fmt.Sprintf("%s/api/file/download/%d#", host, record.Id)

// 	return web.NewEmptyRspBuilder().Put("url", url).JsonResult()
// }
