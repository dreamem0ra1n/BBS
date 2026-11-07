# 接口使用与公共约定

本文对应分支起点 `9cbdd3f` 的后端行为，依据实际路由、控制器、服务及锁定的 `github.com/mlogclub/simple v1.2.1` 源码整理。覆盖当前注册接口；示例尚未逐接口运行联调。

- [Web 接口](./site.md)：登录、帖子、评论、用户、搜索、文件等。
- [后台接口](./admin.md)：用户与内容管理、配置、文件清理等。
- [运行与部署说明](../../README.md)：本地后端、数据库和对象存储准备。

## 地址与请求方式

| 场景 | 地址示例 |
| --- | --- |
| 直连 Go 后端 | `http://localhost:8082/api/topic/topics` |
| 经 Web 同源代理 | `http://localhost:3000/bbs2/api/topic/topics` |
| 新前台规划地址 | `http://localhost:3001/bbs2/api/topic/topics`，尚待实现 |
| 管理 API | Go 路径 `/api/admin/...`，不是后台页面前缀 `/bbsadmin` |

GET 参数放 query；一般 POST 参数使用 `application/x-www-form-urlencoded`。不要默认发送 JSON 请求体。上传使用 `multipart/form-data`，由浏览器或 curl 生成 boundary。例外是某个表单字段内部承载 JSON，例如 `imageList` 或后台 `config`，仍属于表单请求。

接口路径大小写和下划线需保留。例如 `node_auth`、`hide_content`、`sticky_topics`、`send_verify_email` 不应改写。路径中的 `{id}` 表示参数，不包含花括号。文档中的参数默认可省略，标记“必填”的除外；省略不一定代表无副作用，更新接口应参照具体说明。

### 数组、布尔与命名

- `tags`、`tagIds`、`entityIds`、`roles`、`ids` 等由 `FormValue*Array` 读取的数组，发送逗号分隔字符串，如 `tagIds=1,2`。不是 `tagIds[0]=1`，也不是 JSON 数组字符串。标签名本身不要含分隔逗号。
- `imageList` 单独使用 JSON 字符串，如 `[{"url":"/bbs2/api/file/preview/文件UUID"}]`。
- 布尔建议统一发送 `true` / `false`。部分用户设置仅识别大小写不敏感的 `true`，不能统一改为 `1` / `0`。
- 后台列表的数据库列经工具库转换为驼峰参数：`user_id` → `userId`、`source_type` → `sourceType`。响应字段不做统一转换；文件记录依然使用 `file_name` 等下划线字段。
- 时间戳通常为 Unix 毫秒。签到日期另有 `YYYYMMDD` 整数与 `YYYY-MM-DD` 字符串，不当作时间戳处理。

## 身份与权限

BBS 登录凭证读取优先级为：请求参数 `userToken` → 请求头 `X-User-Token` → Cookie `userToken`。推荐请求头或 Cookie，避免把长期凭证放进 URL。旧 Web 还发送 `X-Client: bbs-go-site`，它不是认证凭证。

Passport 换票使用 Cookie **`SESSION_TOKEN`**，BBS 后续请求使用 **`userToken`**；两者不能混用。SSR 向后端请求时需显式传递当前请求的必要凭证，且只能发往配置中的可信后端，不在模块级共享用户状态。

前台是否展示入口以 `/api/config/configs` 的 `loginMethods` 为准。密码登录需后端以 `passwordlogin` build tag 编译且配置启用；生产默认构建不会注册该路径。Passport 也由配置决定是否注册。

管理 API 全部通过 `AdminAuth` 检查登录和 `IsAdminUserOrHigher()`。前台内容操作还可能检查节点管理权限、作者身份、禁言、观察期、邮箱验证、频率限制等。最终结果以后端为准。

## 响应格式

普通接口返回以下包裹结构：

```json
{"errorCode":0,"message":"","data":null,"success":true}
```

未登录示例：

```json
{"errorCode":1,"message":"请先登录","data":null,"success":false}
```

**以 `success` 判断业务结果。** 普通业务错误也可能是 `errorCode: 0`；HTTP 200 不代表业务成功。成功的 `data` 可以为 `null`，列表的 `results` 也可能为 `null`。文件、图片验证码、WebSocket 握手不遵循此 JSON 包裹；网关和未知路由还可能返回 HTML 或纯 HTTP 错误。

| 错误码 | 含义 |
| --- | --- |
| `0` 且 `success=false` | 通用业务错误，查看 `message` |
| `1` | 未登录或凭证无效 |
| `2` | 管理中间件拒绝权限；其他“无权限”错误可能使用 `0` |
| `1000` | 验证码错误 |
| `1001` | 已禁言 |
| `1002` | 账号禁用 |
| `1003` | 观察期未结束 |
| `1004` | 需要邮箱验证 |

后台 HTTP 错误处理可能将 HTTP 状态码填入 `errorCode`。调用方需要同时处理网络失败、HTTP 状态、非 JSON 返回和业务失败，不能全部映射为登录失效或 404。

### 两种分页

页码分页的 `data`：

```json
{"page":{"page":1,"limit":20,"total":42},"results":[]}
```

字段叫 **`page`**，不是 `paging`；总页数由 `total` 和 `limit` 计算。后台通用列表使用 `page=1&limit=20`，非法非正值会回退默认值。前台部分接口固定每页数量，签到排行榜改用 `pageSize`，详见各接口。

游标分页的 `data`：

```json
{"results":[],"cursor":"0","hasMore":false}
```

首次省略 `cursor` 或传 `0`；下一次原样传回响应中的字符串 `cursor`。不要自行加一，也不要认为它一定是帖子 ID；排序方式可能改变游标含义。切换排序、节点、标签时重置列表和游标。

## 本地调用示例

以下为 Bash / Git Bash 示例，PowerShell 可使用 `curl.exe` 并将多行命令合为一行。只在开发数据上执行写请求。Token 和 ID 均需替换，文档不提供真实凭证。

```bash
curl -sS 'http://localhost:8082/api/config/configs'

curl -sS 'http://localhost:8082/api/login/password' \
  --data-urlencode 'username=开发用户名' \
  --data-urlencode 'password=开发密码'
```

登录返回 `data.token` 和 `data.user`；不会自动替 curl 写入 BBS Token Cookie。复制开发 Token 后调用：

```bash
curl -sS 'http://localhost:8082/api/user/current' \
  -H 'X-User-Token: 替换为开发Token'

curl -sS 'http://localhost:8082/api/topic/topics?sort=create&cursor=0' \
  -H 'X-User-Token: 替换为开发Token'

curl -sS 'http://localhost:8082/api/comment/comments?entityType=topic&entityId=123&page=1' \
  -H 'X-User-Token: 替换为开发Token'

curl -sS 'http://localhost:8082/api/comment/create' \
  -H 'X-User-Token: 替换为开发Token' \
  --data-urlencode 'entityType=topic' \
  --data-urlencode 'entityId=123' \
  --data-urlencode 'contentType=markdown' \
  --data-urlencode 'content=本地联调评论'

curl -sS 'http://localhost:8082/api/file/upload/img' \
  -H 'X-User-Token: 替换为开发Token' \
  -F 'image=@./example.png'
```

浏览器简单表单调用：

```js
const response = await fetch('/bbs2/api/comment/create', {
  method: 'POST',
  credentials: 'same-origin',
  body: new URLSearchParams({
    entityType: 'topic',
    entityId: '123',
    contentType: 'markdown',
    content: '本地联调评论',
  }),
})
if (!response.ok) throw new Error(`HTTP ${response.status}`)
const result = await response.json()
if (!result.success) throw new Error(result.message || '提交失败')
```

示例复用同源 `userToken` Cookie；未设置 Cookie 时需提供 `X-User-Token`。实际应用在统一请求层中处理响应，不在每个页面复制这些代码。POST、赠米等写操作不自动重试，避免重复写入。

## 主要数据结构

完整字段定义以 [response.go](../../server/model/response.go)、[data.go](../../server/model/data.go)、[models.go](../../server/model/models.go) 为准，前台返回模型和后台原始记录不是同一种结构。

| 数据 | 主要字段与语义 |
| --- | --- |
| `UserInfo` | `id`、`nickname`、`avatar`、`smallAvatar`、`description`、计数、积分与个人资料字段 |
| `UserProfile` | 继承用户信息，增加 `username`、`roles` 数组、`email`、`emailVerified`、`passwordSet`、`forbidden`、`status`、通知及收藏设置等；用户详情当前也返回此模型 |
| `NodeResponse` | `nodeId`、`name`、`logo`、`description` |
| `TagResponse` | `tagId`、`tagName`、`nodeId`；原始 `Tag` 是 `id`、`name`、`section_id` 等 |
| `TopicResponse` | `topicId`、`type`（0 帖子、1 动态）、`title`、`summary`、`content`、`user`、`node`、`tags`、`access_lv`、`imageList`、计数、推荐/置顶、创建及编辑信息、`gifts`、`isOldBBS` |
| `CommentResponse` | `commentId`、`entityType`、`entityId`、`content`、`user`、`imageList`、`quoteId`、`quote`、`replies`、计数、`status`、`canEdit`、`canDelete`、可选 `rawContent` 与编辑信息 |
| `ImageInfo` | `url`、`preview`；提交图片列表只需各项 `url` |
| `TopicGiftResponse` | `giftId`、`user`、`score`、`reason`、`createTime` |
| `FavoriteResponse` | `favoriteId`、`entityType`、`entityId`、`deleted`、`title`、`content`、`user`、`url`、`createTime` |
| `MessageResponse` | `messageId`、`from`、`userId`、`title`、`content`、`quoteContent`、`type`、`detailUrl`、`extraData`、`status`、`createTime`、可选生日祝福作者 |
| `FileRecord` | `id` 是数据库编号，`file_id` 是下载使用的 UUID；另有 `file_name`、`file_size`、`bucket_name`、`object_name`、`content_type`、`user_id`、`topic_id`、`comment_id`、`source_type`、`managed`、`create_time` |

正文 `content` 通常已由后端渲染，编辑接口返回原文；不要重复当 Markdown 编译。旧 BBS 正文需单独转换。普通状态常量为 `0` 正常、`1` 删除、`2` 待审核；消息状态单独为 `0` 未读、`1` 已读。

## 特殊行为与维护

- 当前部分 GET 有副作用：收藏、取消收藏、退出、读取消息标记已读、后台同步计数和 sitemap；不能做通用预取、缓存或自动重试。
- 搜索访客返回成功的空分页，不一定返回未登录错误。
- 旧 BBS 只在支持的查询入口用 `OLD` 前缀；响应 ID 仍为数字，并附 `isOldBBS`。不要对新站写接口传 `OLD` ID。
- 当前注册清单在 [router.go](../../server/controllers/router.go)。`/api/article/**` 未注册，`UploadController` 不是当前上传入口。后台部分生成的控制器也未挂载，见后台文档。
- 本文描述当前行为，不证明所有权限路径或边界已经经过运行验证。修改接口时同步维护方法、参数、权限、响应、副作用和本地示例。
