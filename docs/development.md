# 开发指南

## 项目与代码入口

项目基于 [mlogclub/bbs-go](https://github.com/mlogclub/bbs-go) 修改。当前运行方式以[仓库 README](../README.md) 为准。

| 模块 | 入口与职责 |
| --- | --- |
| Go 后端 | [main.go](../server/main.go) 为启动入口；[router.go](../server/controllers/router.go) 注册路由 |
| Web API | [controllers/api](../server/controllers/api/) 处理前台请求 |
| 管理 API | [controllers/admin](../server/controllers/admin/) 处理后台请求；[middleware](../server/middleware/) 包含权限中间件 |
| 业务与数据访问 | [services](../server/services/)、[repositories](../server/repositories/)、[model](../server/model/) |
| 后台任务 | [scheduler](../server/scheduler/) |
| 后端配置 | [配置示例](../server/bbs-go.example.yaml) 与[配置定义](../server/pkg/config/config.go) |
| 现有 Web 前台 | [site](../site/) 使用 Nuxt 2；页面在 [pages](../site/pages/)，状态在 [store](../site/store/)，运行与代理配置在 [nuxt.config.js](../site/nuxt.config.js) |
| 现有管理后台 | [admin](../admin/) 使用 Vue 2 与 Element UI；页面在 [views](../admin/src/views/)，路由在 [router](../admin/src/router/) |
| 微信小程序 | [miniprogram](../miniprogram/) 使用 uni-app、Vue 3 与 TypeScript；页面在 [pages](../miniprogram/src/pages/)，环境配置在 [env.ts](../miniprogram/src/config/env.ts) |

涉及前端迁移时，按工作区中已有的 `FRONTEND_REFACTOR.md` 和用户确认的任务范围执行。计划中未实现的内容不能当作当前代码能力。

## 开发时如何核对行为

- 启动环境前阅读仓库 README 的环境要求、登录模式和初始化步骤。普通生产构建不包含密码登录；本地密码登录还需要开发构建、配置与有效开发账号。
- 修改接口前，从路由注册定位控制器，再检查业务服务、权限校验与数据模型。控制器文件存在不代表接口已经注册。
- 前后端联调先遵循[接口公共约定](api/README.md)，特别是表单编码、Token、业务成功判断和分页格式。
- 修改返回字段时，同时检查前台消费方式、后台原始模型以及小程序调用，不假定各端使用相同的数据结构。
- 区分源码核对、构建通过与真实联调。验证记录应写明实际执行的检查和未验证的行为。

## 文档维护

- 接口参数、权限、返回结构或副作用变更时，同步更新对应的 `api/` 文档。
- 环境、启动或部署步骤变更时，更新仓库 README；其他文档链接到该处。
- 新增专题时使用普通 Markdown 和相对链接，在[文档索引](README.md) 中登记；引用实现时优先链接源码文件。
- 记录计划时明确标注状态，不将待实现或未联调的功能写成既有能力。
- 文档调整后检查相对链接和旧路径引用即可；本目录没有构建或发布步骤。
