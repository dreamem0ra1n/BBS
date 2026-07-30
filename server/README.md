## 介绍

该目录是项目的 Go 后端。完整的本地与服务器部署流程以仓库根目录的 `README.md` 为准。

## 后端镜像构建

生产镜像默认不编译密码登录提供者；生产示例配置只启用求是潮 Passport：

```bash
docker build -t bbs-neo-backend ./server
```

本地开发镜像需要显式选择 `dev` target，该 target 会使用 `passwordlogin` build tag：

```bash
docker build --target dev -t bbs-neo-backend-dev ./server
```

登录方式由配置文件中的 `LoginMethods` 声明。开发配置启用 `password`，生产配置只启用 `passport`。业务代码不根据 `Env` 选择登录方式；生产镜像即使误配 `password: true`，也不会注册 `/api/login/password`。

密码登录接口为 `[POST] /api/login/password`，只供本地调试使用，要求数据库中的普通用户名和密码。生产构建不会注册该接口。

## 魔改后的接口介绍

### 通过 passport 登录

`[POST]` `/api/login/signin`

需要前端在走完 passport 逻辑后携带着 success 的 cookie 和登录后希望跳转的地址(这个地址怎么传需要参考前端如何解析这个请求的 response 的对应参数)

```json
data: {
    "ref": string_of_url_before_login,
}
```

### 修改用户资料

`[POST]` `/api/user/edit/{uid_long}`

```html
"homePage=个人主页&description=个人介绍&major=专业&birthday=生日&mobile=手机&"wechat=微信号&qq=QQ号"
```

### 根据部门 id 获取 tag

`[GET]` `/api/tag/list/{section_id}`

通过 部门id 查询可用 tag 的列表，部门 id 同权限体系那一部分。

如果 section_id 为 0，则返回全局 tag，如果不为 0，则返回全局 tag 和部门 tag

```json
data: {
    [
        {
            "id": tag id,
            "name": tag 名称,
            "section_id": 部门 id,
            "description": 描述,
            "status": 可用状态，0 为可用,
            "createTime": 创建时间,
            "updateTime": 上次更新时间
        },
    ]
}
```


### 根据 NodeId 和 TagId 获取帖子列表

`[POST]` `/api/topic/topicsnt`

通过 部门id 和 tagid 查询帖子列表

req 的 form 中需要传入：

```html
"cursor=cursor_id&nodeId=node_id&tagId=tag_id"
```

resp 同 `/api/topic/topics`


### 上传文件

`[POST]` `/api/file/upload` (multipart/form)

resp:

```json
data: {
		"file_name":   fileNameOri,
		"file_id":   fileUUID,
		"file_size":   fileSize,
		"bucket_name": bucketName,
}
```

### 下载文件

`[GET]` `/api/file/download/{file_id}`

resp: file

## OLD-BBS 兼容性修改

### ID

- topic.getById和comment.getById，id参数加上字符串前缀OLD表示旧BBS
- 为了Json兼容性，返回的id字段仍然只是int64，不带OLD字符串前缀

### Model修改
- topic和comment都添加了`bool IsOldBBS`字段
- 部分无关字段置为null或0（点赞、收藏等）

### new apis

- `/search/oldbbs`
  - param `page,keyword,limit`
  - return same as `/search/topic`
- `/topic/tag/topics`
  - add param: `tagId="OLD"+forumId`
