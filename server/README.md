## 介绍
该项目使用Golang进行构建，具体参见：https://mlog.club

## 魔改后的接口介绍

### 通过 passport 登录

`[POST]` `/api/login/signin`

需要前端在走完 passport 逻辑后携带着 success 的 cookie 和登录后希望跳转的地址(这个地址怎么传需要参考前端如何解析这个请求的 response 的对应参数)

```json
{
    "ref": string_of_url_before_login,
}
```

### 修改用户资料

`[POST]` `/api/user/edit/{uid_long}`

```json
{
    "homePage": 个人主页,
    "description": 个人介绍,
    "major": 专业,
    "birthday": 生日,
    "mobile": 手机号,
    "wx": 微信号,
    "qq": QQ号,
}
```