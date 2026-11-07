# 管理后台接口

先阅读 [公共约定](./README.md)。以下路径相对 `/api/admin`，全部要求 BBS 登录并通过 `IsAdminUserOrHigher()` 检查。界面是否显示按钮不能替代后端检查。

返回均指 `data`。`AnyList` 接受多个 HTTP 方法，文档统一推荐 GET；页码参数默认 `page=1&limit=20`。列表返回 `{page:{page,limit,total},results}`，results 可为 null。筛选参数为驼峰名称，与原始模型字段可能不同。

## 用户与帖子

| 方法 | 路径 | 参数 | 返回或行为 |
| --- | --- | --- | --- |
| GET | `/user/list` | `id`、`nickname` 模糊、`username` 精确、分页 | 用户分页；含 roles 数组、username/email 字符串、score、forbidden |
| GET | `/user/{id}` | 路径 ID | 用户详情 |
| POST | `/user/update` | `id` 必填；`email`、`roles` 逗号分隔、`status` | 更新用户；不会修改用户名、昵称和密码 |
| POST | `/user/forbidden` | `userId`、`days=0`、`reason` | 0 天解除禁言，其他天数交服务处理，返回 null |
| GET | `/user/synccount` | 无 | 异步同步用户计数，返回 null，有副作用 |
| GET | `/topic/list` | `id/userId/status/recommend` 精确、`title` 模糊、分页 | 简化帖子分页，另附 status |
| GET | `/topic/{id}` | 路径 ID | 原始 Topic，含原始正文 |
| POST | `/topic/update` | `id` 或 `topicId`，`nodeId` 必填，`tagIds` 逗号分隔 | 只更新节点与标签，返回 null |
| POST | `/topic/recommend` | `id` 必填 | 设为推荐 |
| DELETE | `/topic/recommend` | `id` 必填，推荐 query 参数 | 取消推荐 |
| POST | `/topic/delete` | `id` 必填 | 软删除 |
| POST | `/topic/undelete` | `id` 必填 | 恢复帖子及相关处理 |

用户更新直接写入 email、roles、status，不是部分更新；缺失 roles 可能清空角色，缺失 status 默认 -1。先读取详情再提交完整表单。`/user/create` 已注释，未注册。

## 评论、节点与标签

| 方法 | 路径 | 参数 | 返回或行为 |
| --- | --- | --- | --- |
| GET | `/comment/list` | `id`、`userId`、`entityType` 与 `entityId`、`status`、分页 | 评论分页；没有 id/userId/完整实体条件时成功返回 null |
| GET | `/comment/{id}` | 路径 ID | 原始 Comment |
| POST | `/comment/delete/{id}` | 路径 ID | 删除，返回 null |
| GET | `/topic-node/list` | `name` 精确、分页 | 节点分页，按 sortNo 升序、id 倒序 |
| GET | `/topic-node/{id}` | 路径 ID | 原始 TopicNode |
| GET | `/topic-node/nodes` | 无 | 节点数组 |
| POST | `/topic-node/create` | 节点模型表单 | 新节点 |
| POST | `/topic-node/update` | `id` 必填，节点模型表单 | 更新后节点 |
| GET | `/tag/list` | `id/name` 模糊、`status` 精确、分页 | 原始 Tag 分页 |
| GET | `/tag/{id}` | 路径 ID | 原始 Tag |
| POST | `/tag/create` | 标签模型表单，非空唯一 `name` | 新标签，status 置为正常，创建/更新时间由服务端生成 |
| POST | `/tag/update` | `id` 必填，标签模型表单 | 标签名须非空且不与他人冲突 |
| GET | `/tag/autocomplete` | `keyword` | TagResponse 数组；非空关键词按前缀，空值取最近 10 条 |
| GET | `/tag/tags` | `tagIds` 逗号分隔 | TagResponse 数组或 null |
| GET | `/tag/all` | `sectionId=0` | 正常原始 Tag 数组；正数包含全局与该部门，否则仅全局 |

节点表单字段：`name/description/logo/sortNo/status`，模型另有 `createTime`。标签表单字段：`name/section_id/description/status`。标签写入字段是 `section_id`，而 `/tag/all` 使用 `sectionId`。

模型表单由 `ReadForm` 解码，更新先加载现有记录；显式空值会覆盖原字段。只发送需要编辑的业务字段，不把整个响应对象原样回传。

## 文件管理

| 方法 | 路径 | 参数 | 返回 |
| --- | --- | --- | --- |
| GET | `/file/list` | `id/userId/topicId/commentId/sourceType` 精确、`fileName` 模糊、分页 | 文件分页，附 previewUrl/sourceLabel，可能附 sourceUrl/topicTitle/topicStatus |
| GET | `/file/{id}` | 数据库数字 ID | 原始 FileRecord |
| POST | `/file/delete` | 表单数字 `id` | 成功 null，失败返回原因 |

删除使用数据库 id，不是 UUID `file_id`。后端在删除时检查实际引用，包含仍在数据库中的软删除内容。来源、时间、managed 标记和历史关联 ID 均不能代替引用检查。

没有批量删除 API；现有后台逐个调用并汇总结果。引用存在、核验失败或对象存储失败均不能当作成功。删除不可通过后台恢复；未提交草稿与数据库之外外链不在核验范围。定时清理和容量排查见仓库 [README](../../README.md)。

## 文章与标签关系

后台文章 API 已注册，Web `/api/article/**` 当前未注册。

| 方法 | 路径 | 参数 | 返回或行为 |
| --- | --- | --- | --- |
| GET | `/article/list` | `id/userId/status/title` 精确、分页 | 文章分页，含 user、summary、tags，排除正文 |
| GET | `/article/{id}` | 路径 ID | 原始 Article |
| POST | `/article/update` | `id` 必填，文章模型字段 | 更新文章；title/content/contentType 必须非空 |
| GET | `/article/tags` | `articleId` | TagResponse 数组 |
| POST | `/article/tags` | `articleId`、`tags` 名称逗号分隔 | 更新后的标签数组 |
| POST | `/article/delete` | 正数 `id` | 删除 |
| POST | `/article/pending` | 正数 `id` | status 设为正常（审核通过） |
| GET | `/article/sitemap` | 无 | 异步生成 sitemap，有副作用 |
| GET | `/article-tag/list` | 分页 | 原始 ArticleTag 分页 |
| GET | `/article-tag/{id}` | 路径 ID | 原始 ArticleTag |
| POST | `/article-tag/create` | 关系模型表单 | 新记录 |
| POST | `/article-tag/update` | `id` 必填，关系模型表单 | 更新后记录 |

文章可绑定字段以 [models.go](../../server/model/models.go) 的 form tag 为准，常用 `title/content/contentType/summary/sourceUrl/status/access_lv/sectionId`。关系模型为 `articleId/tagId/status/createTime`，通过主键 id 更新。

## 收藏、友链与生日祝福

| 方法 | 路径 | 参数 | 返回或行为 |
| --- | --- | --- | --- |
| GET | `/favorite/list` | 分页 | 原始 Favorite 分页 |
| GET | `/favorite/{id}` | 路径 ID | 原始 Favorite |
| POST | `/favorite/create` | `userId/entityType/entityId/createTime` | 新记录 |
| POST | `/favorite/update` | `id` 必填，同上模型字段 | 更新记录 |
| GET | `/link/list` | `status` 精确、`title/url` 模糊、分页 | 原始 Link 分页 |
| GET | `/link/{id}` | 路径 ID | 原始 Link |
| POST | `/link/create` | `url/title/summary/logo/status` | 新记录，服务端设置创建时间 |
| POST | `/link/update` | `id` 必填，同上模型字段 | 更新记录 |
| GET | `/birthday-blessing/list` | `nickname/content` 模糊、分页 | 祝福分页 |
| POST | `/birthday-blessing/create` | `nickname/content` 模型字段 | 新祝福，内容不能为空 |
| POST | `/birthday-blessing/import` | `nickname/content` | 新祝福，内容不能为空 |
| POST | `/birthday-blessing/batch/import` | 表单 `data`，值为 CSV 文本 | `{created: 数量}` |
| POST | `/birthday-blessing/delete` | `ids` 逗号分隔或单个 `id` | null |

CSV 前两列为昵称、内容，支持 `昵称,内容` 表头和带引号多行字段；不足两列或归一化后空内容行跳过。导入逐行写入，后续失败不回滚此前成功行，重试前核对已导入数据。批量删除也可能部分成功。

现有后台页面调用的是 `/birthday-blessing/batch-import`，而控制器 `PostBatchImport` 按 Iris MVC 命名映射为 `/birthday-blessing/batch/import`，未发现显式别名注册。此处记录差异供联调核对，本次文档工作不修改原业务调用。

## 配置、日志与维护

| 方法 | 路径 | 参数 | 返回或行为 |
| --- | --- | --- | --- |
| GET | `/sys-config/list` | 分页 | 原始 SysConfig 分页 |
| GET | `/sys-config/{id}` | 路径 ID | 原始 SysConfig |
| GET | `/sys-config/all` | 无 | 汇总站点配置，不是原始记录数组 |
| POST | `/sys-config/save` | 表单 `config`，值为 JSON 对象字符串 | 事务写入提供的 key，返回 null |
| GET | `/user-score-log/list` | `userId/sourceType/sourceId/type`、分页 | 积分流水分页，附 user |
| GET | `/user-score-log/{id}` | 路径 ID | 原始 UserScoreLog |
| GET | `/operate-log/list` | `userId/opType/dataType/dataId`、分页 | 原始 OperateLog 分页 |
| GET | `/operate-log/{id}` | 路径 ID | 原始 OperateLog |
| GET | `/common/systeminfo` | 无 | `{os,arch,numCpu,goversion,hostname}` |
| POST | `/search/reindex` | 无 | 异步重建新站索引；SyncOldBBS 启用时也重建旧索引 |

配置保存仅写提供的 key；不要把公开配置中的派生字段全部回传。实际登录能力来自后端构建及配置，不能通过保存 `loginMethods` 打开密码登录。异步维护接口成功只表示任务启动，完成情况查看日志；当前无任务进度 API。

本地调用示例（Bash / Git Bash），仅对开发数据执行写操作：

```bash
curl -sS 'http://localhost:8082/api/admin/file/list?sourceType=avatar&page=1&limit=20' \
  -H 'X-User-Token: 替换为开发管理员Token'

curl -sS 'http://localhost:8082/api/admin/sys-config/save' \
  -H 'X-User-Token: 替换为开发管理员Token' \
  --data-urlencode 'config={"siteTitle":"本地论坛"}'

curl -sS 'http://localhost:8082/api/admin/topic/update' \
  -H 'X-User-Token: 替换为开发管理员Token' \
  --data-urlencode 'id=123' \
  --data-urlencode 'nodeId=11' \
  --data-urlencode 'tagIds=1,2'
```

## 未注册控制器与核对范围

当前 [router.go](../../server/controllers/router.go) 未挂载后台 check_in、email_code、message、topic_tag、user_feed、user_follow、user_token 控制器。源文件存在不代表路由可调用。

后台详情通常返回原始模型，不与前台渲染模型混用。字段见 [models.go](../../server/model/models.go)，行为见 [后台控制器](../../server/controllers/admin/)。本文为源码核对结果，未执行删除、导入、通知或维护任务验证。
