# Web 接口

先阅读 [公共约定与调用示例](./README.md)。以下路径相对 Go 的 `/api`，返回值指响应的 `data`。参数不等于全部必填，显式标记及业务校验要求为准；内容发布还受禁言、观察期、邮箱和频率限制。

## 登录与配置

| 方法 | 路径 | 参数与条件 | 返回 |
| --- | --- | --- | --- |
| GET | `/config/configs` | 访客可用 | 站点配置 |
| POST | `/login/password` | 表单 `username`、`password` 必填，`ref` 可选；仅编译和配置同时启用时存在 | `{token,user,ref}` |
| POST | `/login/signin` | Cookie `SESSION_TOKEN`；表单 `ref` 可选；Passport 启用时存在 | `{token,user,ref}` |
| GET | `/login/signout` | 当前 BBS 凭证 | `null`；使当前 Token 失效 |

配置主要字段：`siteTitle`、`siteDescription`、`siteKeywords`（字符串数组）、`siteNavs`、`siteNotification`、`recommendTags`、`urlRedirect`、`scoreConfig`、`defaultNodeId`、`articlePending`、`topicCaptcha`、`userObserveSeconds`、`tokenExpireDays`、`loginMethods`、各 `create*EmailVerified`、`enableHideContent`、`birthdayRandomBlessing`。完整定义在 [data.go](../../server/model/data.go)。

登录能力读取 `loginMethods.passport/password`，不要使用历史 `loginMethod` 字段决定入口。密码登录关闭时路径不存在，不能根据前端环境强行开启。

Passport Web 流程：跳转 Passport → 回到 Web 并携带 `SESSION_TOKEN` → 写会话 Cookie → POST 换票 → 保存返回的 BBS Token → 更新用户并清理 URL 中的会话参数。后端返回的 `ref` 只是回显，站内回跳合法性由前端检查。BBS 退出不会自动完成 Passport 全局退出；旧 Web 另请求 Passport `/passport/v4/logout`。

## 节点与帖子

| 方法 | 路径 | 参数与条件 | 返回与行为 |
| --- | --- | --- | --- |
| GET | `/topic/nodes` | 无 | 节点数组 |
| GET | `/topic/node` | `nodeId` 默认 0 | 节点或 `null` |
| GET | `/topic/topics` | `cursor=0`、`nodeId=0`、`recommend`、`sort` | 帖子游标分页；`sort=create` 按创建时间，否则按评论时间 |
| POST | `/topic/topicsnt` | `nodeId`、`tagId` 必填，`cursor=0`、`sort` | 节点与标签联合查询，游标分页 |
| GET | `/topic/tag/topics` | `tagId` 必填，可带 `OLD` 前缀；`cursor=0` | 标签或旧论坛游标分页 |
| GET | `/topic/user/topics` | `userId` 必填，`cursor=0` 或 `page>0` | 有 page 时每页 20 条，否则游标分页 |
| GET | `/topic/recent` | 无 | 最近 10 条 |
| GET | `/topic/newest` | 无 | 最近 6 条，按访客身份渲染 |
| GET | `/topic/recommend` | 无 | 推荐缓存随机取最多 10 条，可能 `null`，按访客身份渲染 |
| GET | `/topic/sticky_topics` | `nodeId=0`、`tagId=0` | 最多 3 条置顶 |
| GET | `/topic/{id}` | 支持 `OLD` ID；`count_view` 默认 1 | 详情，检查阅读权限；新帖有权限时增加浏览量，传 0 禁止累加 |
| POST | `/topic/create` | 登录且可发言；字段见下表 | 简化帖子模型 |
| GET | `/topic/edit/{id}` | 登录且可发言；作者或可管理该帖；帖子类型为 0 | `{topicId,nodeId,title,content,hideContent,tags,access_lv}`，tags 为名称数组 |
| POST | `/topic/edit/{id}` | 作者或管理权限；`nodeId`、`title`、`content`、`hideContent`、`tags`、`access_lv` | 简化帖子；`access_lv` 默认 -1，不能小于 -1 |
| POST | `/topic/delete/{id}` | 登录且可发言，须有管理该帖权限 | `null`；作者身份本身不赋予删除权限 |
| POST | `/topic/recommend/{id}` | 登录且可管理该帖，`recommend` 必填布尔 | `null` |
| POST | `/topic/sticky/{id}` | 管理员，`sticky=false` | `null` |
| POST | `/topic/like/{id}` | 登录 | `null` |
| POST | `/topic/gift/{id}` | 登录，`score`、`reason` 必填 | 赠米记录；见下文限制 |
| GET | `/topic/recentlikes/{id}` | 无 | 最近最多 5 位点赞用户 |
| GET | `/topic/favorite/{id}` | 登录 | `null`，添加收藏，有副作用 |
| GET | `/topic/hide_content` | `topicId`；登录作者或已回复用户才可显示 | `{exists,show,content}` |

帖子列表会按阅读权限处理标题或内容；详情权限不足是业务失败，不能从列表可见推断正文可读。列表与详情可空字段需分别处理。

### 发帖表单

| 字段 | 默认与约束 |
| --- | --- |
| `type` | 0 帖子，1 动态 |
| `nodeId` | 默认 0；具体发布规则由服务端检查 |
| `title` | 普通帖子必填 |
| `content` | 普通帖子必填；动态需内容或图片；最多 5000 字符 |
| `hideContent` | 可选，最多 5000 字符 |
| `tags` | 标签名称逗号分隔字符串 |
| `access_lv` | 默认 1，阅读权限等级 |
| `imageList` | 图片对象数组的 JSON 字符串，每项含 `url` |
| `captchaId`、`captchaCode` | 站点开启 `topicCaptcha` 时需要 |

编辑不是 PATCH 语义：遗漏标题、正文、标签等不能视为保持旧值，应先获取编辑数据再提交表单。赠米不能给自己，必须余额足够；数量在 1 到站点 `scoreConfig.giftScoreMax` 范围内，服务端最高限制 50；理由必填且最多 15 字。赠米成功会转移积分并产生通知，不自动重试。

## 评论

| 方法 | 路径 | 参数与条件 | 返回 |
| --- | --- | --- | --- |
| GET | `/comment/comments` | `entityType`、`entityId`；`page=0`、`cursor=0`、`asc_order=0` | `page>0` 页码分页，否则游标分页；entityId 支持 OLD |
| GET | `/comment/replies` | `commentId`、`cursor=0` | 二级回复游标分页，每批 10 条 |
| GET | `/comment/location/{commentId}` | `asc_order=0`；检查所属新帖阅读权限 | `{page,rootId,entityId,comment}` |
| GET | `/comment/user/comments` | 登录本人，`userId`、`page=1`、`asc_order=0` | 用户评论页码分页 |
| POST | `/comment/create` | 登录且可发言，字段见下文 | 新评论模型 |
| POST | `/comment/edit/{commentId}` | 登录且可发言、可管理评论；`content`、`imageList` | 更新后的评论 |
| POST | `/comment/delete/{commentId}` | 登录且可发言、可删除评论 | `null`；按服务规则维护计数 |
| POST | `/comment/like/{commentId}` | 登录 | `null` |

创建字段：`entityType`（帖子为 `topic`，回复评论为 `comment`）、正数 `entityId`、非空 `content`（最多 5000 字符）、`contentType`（默认 `text`，Markdown 编辑器须显式传 `markdown`）、可选 `imageList` JSON 字符串、`quoteId=0`。两种回复目标不要混淆：回复父评论使用 `entityType=comment` 和父评论 ID；引用评论另用 `quoteId`。

`asc_order=0` 为倒序，非零为正序；帖子评论页码模式每页 10 条。`replies` 可为 `null` 或游标结构，`rawContent` 仅在允许编辑时返回。编辑接口只修改正文与图片，不通过此接口改变评论内容类型。

现有列表和回复控制器没有与帖子详情等价的统一阅读权限检查。前端应先确认详情可读再请求评论；这不能替代后端鉴权，也不应将本文理解为后端所有评论入口都已实现权限保障。

## 用户与个人设置

| 方法 | 路径 | 参数与条件 | 返回或行为 |
| --- | --- | --- | --- |
| GET | `/user/current` | 可无登录态 | 当前 `UserProfile` 或成功的 `null` |
| GET | `/user/node_auth/{nodeId}` | 可无登录态 | `{section,read_lv,manage_lv}`，访客为默认权限 |
| GET | `/user/{userId}` | 登录 | 当前实现返回 `UserProfile`，不是精简公开资料 |
| POST | `/user/edit/{userId}` | 只能本人；字段见下文 | `null` |
| POST | `/user/update/avatar` | 登录，非空 `avatar` URL | `null` |
| POST | `/user/set/background/image` | 登录，非空 `backgroundImage` | `null` |
| POST | `/user/set/username` | 登录，`username` | `null`，受账号服务规则限制 |
| POST | `/user/set/email` | 登录，`email` | `null` |
| POST | `/user/set/password` | 登录，`password`、`rePassword` | `null`，设置密码 |
| POST | `/user/update/password` | 登录，`oldPassword`、`password`、`rePassword` | `null`，修改密码 |
| POST | `/user/send_verify_email` | 登录 | `null`，发送验证邮件 |
| POST | `/user/verify_email` | `token` 必填 | `{email}` |
| GET | `/user/favorites` | 登录，`userId` 默认自己；`page>0` 或 `cursor=0` | 页码或游标分页，每页 20 条；他人的收藏须公开 |
| GET | `/user/msgrecent` | 无登录也可调用 | `{count,messages}`，最多 3 条未读；访客为 0 / 空 |
| GET | `/user/messages` | 登录，`page=1`、`category` | 每页 20 条；将所选分类全部消息标为已读，不仅当前页 |
| GET | `/user/scorelogs` | 登录，`page=1` | 每页 20 条积分流水 |
| GET | `/user/score/rank` | `period=year/all`，其他值为新人榜 | 用户数组，响应 `no-store` |
| POST | `/user/forbidden` | 任意 | 当前固定返回“不可用”，使用后台禁言接口 |
| GET | `/user/dingtalk/settings` | 登录 | `{enabled,webhookConfigured,secretConfigured,keyword}` |
| POST | `/user/dingtalk/settings` | 登录；见下文 | 同上 |
| POST | `/user/dingtalk/test` | 登录 | `null`；实际发送测试通知 |

资料表单：`homePage`、`description`、`greeting`（最多 20 字）、`major`、`birthday`、`mobile`、`wechat`、`qq`、`birthdayBlessingEnabled`、`birthdayBlessingNotifyEnabled`、`publicFavorites`。昵称不能通过此接口修改；问候、生日等通知开关还受站点配置影响。此接口会整体写入这些字段，遗漏文本可能清空、遗漏开关会变 false，应先读取当前资料。

消息分类：`system`、`like`、`reply`、`mention`；空值或未知值为全部。读取消息有已读副作用，不做后台预取。

钉钉表单：`enabled`、`webhook`、`secret`、`keyword`、`clearSecret`。空 webhook/secret 保留已存值；`clearSecret=true` 且未提供新 secret 时清除旧密钥；enabled 和 keyword 按本次值写入。开启需要有效 webhook。响应只返回配置状态，不返回保存的 webhook 或密钥。文档示例不调用测试通知。

## 搜索、标签、关注与收藏

| 方法 | 路径 | 参数 | 返回与条件 |
| --- | --- | --- | --- |
| POST | `/search/topic` | `keyword`、`nodeId=0`、`page=1`、`limit=20` | 帖子页码分页；nodeId=-1 表示推荐 |
| POST | `/search/oldbbs` | `keyword`、`page=1`、`limit=20` | 旧帖页码分页 |
| GET | `/tag/{tagId}` | 路径 ID | `TagResponse` |
| GET | `/tag/tags` | `page=1` | 每页 200 条标签 |
| POST | `/tag/autocomplete` | `input` | 补全结果 |
| GET | `/tag/list/{sectionId}` | sectionId=0 为全局，其他值为全局加该部门 | 原始 Tag 数组，字段含 `id/name/section_id` |
| POST | `/fans/follow` | `userId` 正数必填 | 登录，关注，返回 null |
| POST | `/fans/unfollow` | `userId` 正数必填 | 登录，取消关注，返回 null |
| GET | `/fans/isfollowed` | `userId` | 布尔；访客为 false |
| GET | `/fans/fans` | `userId`、`cursor=0` | 粉丝游标分页，每批 10 条 |
| GET | `/fans/follows` | `userId`、`cursor=0` | 关注游标分页，每批 10 条 |
| GET | `/fans/recent/fans` | `userId` | 最近 10 位粉丝，游标结构 |
| GET | `/fans/recent/follow` | `userId` | 最近 10 位关注，游标结构 |
| GET | `/feed/topics` | `cursor=0` | 登录，关注动态游标分页 |
| GET | `/like/is/liked` | `entityType`、`entityIds` 逗号分隔 | `{liked: 已点赞ID数组}` |
| GET | `/like/liked` | `entityType`、`entityId` | `{liked: 布尔}` |
| GET | `/favorite/favorited` | `entityType`、`entityId` | `{favorited: 布尔}` |
| GET | `/favorite/delete` | `entityType`、`entityId` | 登录，取消收藏，有副作用 |

搜索在未登录、空关键词或超出页码上限时返回成功的空结果；Meilisearch 不可用时回退数据库 LIKE。`limit/page` 受 `Search.MaxLimit/MaxPage` 限制，权限按配置过滤或标记。不要假设“搜索成功但空结果”等价于没有匹配。

关注列表控制器未显式要求登录；其访客路径仍需联调验证，不能仅凭接口存在推断所有匿名数据组合均可用。

## 签到、友链、统计与验证码

| 方法 | 路径 | 参数与条件 | 返回 |
| --- | --- | --- | --- |
| POST | `/checkin/checkin` | 登录且可发言 | 签到汇总，重复签到失败 |
| GET | `/checkin/checkin` | 可无登录态 | 签到记录附 `checkIn` 布尔，或 null |
| GET | `/checkin/overview` | 登录，`month`，格式 YYYY-MM | 月度签到概览 |
| POST | `/checkin/makeup` | 登录且可发言，无日期参数 | 自动补最近漏签日；返回 `date/checkInType/consecutiveDays/totalCheckInDays/makeupCards` |
| GET | `/checkin/rank/today` | `page=1`、`pageSize=20`，最大100 | 当日排名，页码结构；各项含 rank、user、checkInTime、checkInType |
| GET | `/checkin/rank/consecutive` | 同上 | 连续签到排名；各项含 rank、user、consecutiveDays、latestCheckInDate |
| GET | `/checkin/rank` | 无 | 前 10 位签到用户数组 |
| GET | `/link/{id}` | 无 | `{linkId,url,title,summary,logo,createTime}` |
| GET | `/link/links` | `page=1` | 每页 20 条友链 |
| GET | `/link/toplinks` | 无 | 前 10 条友链 |
| GET | `/stats/site` | 无 | `totalTopics/monthlyTopics/totalUsers/monthlyActiveUsers` |
| GET | `/captcha/request` | 可选 `captchaId` 用于重载 | `{captchaId,captchaUrl}` |
| GET | `/captcha/show` | `captchaId` | PNG 图片；不存在时 HTTP 404 |
| GET | `/captcha/verify` | `captchaId`、`captchaCode` | `{success: 布尔}`，这是 data 内字段，不能与外层混淆 |

签到汇总字段：`consecutiveDays`、`totalCheckInDays`、`makeupCards`、`checkedInToday`、`latestDayName`。月概览返回 `today/todayTotal/checkedInToday/consecutiveDays/totalCheckInDays/makeupCards/month/calendar`；calendar 各项为 `date/status/type`，status 可为 `missed/future/before_register/makeup/checked`。month 省略为当月，不允许未来月份。

签到与月统计日期边界采用服务端本地时区；补签会消耗卡片，不能传任意日期。验证码验证会消费验证码，不要先独立 verify 再用同一个验证码发布。

## 文件与在线状态

| 方法 | 路径 | 参数与条件 | 返回 |
| --- | --- | --- | --- |
| POST | `/file/upload` | 登录且可发言；multipart 字段 `file` | `FileRecord` |
| POST | `/file/upload/img` | 登录且可发言；multipart 字段 `image`，可选 `source` | `{url}`，使用服务端 BaseUrl 构造预览地址 |
| GET | `/file/download/{fileId}` | UUID，不是数据库 id | 附件文件流 |
| GET | `/file/preview/{fileId}` | UUID | 内联文件流，错误可能为 HTTP 400 |
| GET | `/img/proxy` | query `url` | 代理内容流，非 JSON；上游请求失败为 HTTP 502 |
| POST | `/presence/ticket` | 登录，Redis 可用 | `{ticket}`，30 秒有效、一次性 |
| GET | `/presence/online` | 登录 | `{users,total}`，不可用时空快照 |
| GET / WebSocket | `/ws/presence` | query `ticket`、同一用户登录态、允许的 Origin | WebSocket 消息 |

文件上限为 30 MiB，图片和附件分别使用 `image` 和 `file` 字段。图片来源识别 `avatar/background/node_logo/link_logo`，其他来源按 unattached 处理。后续发帖、编辑、评论关联实际引用；来源标签本身不等于文件仍被引用。下载和预览控制器当前没有登录检查，不应将其描述为“随帖子权限保护的私有下载”。

WebSocket 地址经代理通常为 `/bbs2/api/ws/presence?ticket=...`。浏览器需携带同源登录 Cookie；ticket 不能代替登录态。服务端校验 Redis、Origin、ticket、登录用户一致性和连接上限；失败可能为 503、403、401、429。

推送格式为 `{"type":"snapshot","users":[],"total":0}` 或 `{"type":"unavailable"}`。用户项为 `id/nickname/avatar/greeting`；users 最多 200 位，total 是真实在线人数。服务端每 20 秒发协议 Ping，浏览器自动 Pong；重连需新 ticket，不发送 BBS Token 作为 ticket。

## 源码入口

[路由](../../server/controllers/router.go)、[Web 控制器](../../server/controllers/api/)、[表单定义](../../server/model/form.go)、[响应模型](../../server/model/response.go)、[错误码](../../server/pkg/errs/code_errors.go)、[反垃圾规则](../../server/spam/)。当前 `/api/article/**` 未注册，不从旧页面或控制器文件推断其可用性。
