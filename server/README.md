## 介绍
该项目使用Golang进行构建，具体参见：https://mlog.club

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

```json
data: {
    "homePage": 个人主页,
    "description": 个人介绍,
    "major": 专业,
    "birthday": 生日,
    "mobile": 手机号,
    "wx": 微信号,
    "qq": QQ号,
}
```

### 根据部门 id 获取 tag

`[GET]` `/api/tag/list/{section_id}`

通过 部门id 查询可用 tag 的列表，部门 id 同权限体系那一部分。

如果 section_id 为 0，则返回全局 tag，如果不为 0，则返回全局 tag 和部门 tag

```json
data: {
    [
        {
            "name": tag 名称,
            "section_id": 部门 id,
            "description": 描述,
        },
    ]
}
```


### 根据 NodeId 和 TagId 获取帖子列表

`[GET]` `/api/topic/topicsnt`

通过 部门id 和 tagid 查询帖子列表

req 的 body 中需要传入：
```json
{
    "cursor": ,
    "nodeId": node id,
    "tagId": tag id,
}
```

resp 同 `/api/topic/topics`