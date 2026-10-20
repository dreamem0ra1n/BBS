# bbs-neo

## 项目说明

该项目魔改自 [bbsgo](https://github.com/mlogclub/bbs-go)，由以下组件组成：

| 组件 | 目录 | 默认端口 | 说明 |
| --- | --- | --- | --- |
| Server | `server/` | `8082` | Go API 服务 |
| Site | `site/` | `3000` | Nuxt 2 SSR 前台，访问前缀 `/bbs2` |
| Admin | `admin/` | `8080` | Vue 2 管理后台，访问前缀 `/bbsadmin` |
| Mini Program | `miniprogram/` | - | uni-app + Vue 3 + TypeScript 微信小程序端 |
| MySQL | - | `3306` | 主数据库；推荐使用 MySQL 5.7 |
| MinIO | - | `9000/9001` | 文件对象存储及管理控制台 |
| Meilisearch | - | `7700` | 可选；搜索索引服务，只保存后端同步的索引，不启动则搜索回退到数据库 `LIKE` |

本文档中的命令默认在仓库根目录执行。

## 登录模式与安全边界

后端存在两个不同的构建产物：

| 构建方式 | 本地账号密码登录 | Passport 登录 | 用途 |
| --- | --- | --- | --- |
| 默认构建 | 未编译进二进制 | 由 `LoginMethods.passport` 配置 | 服务器/生产环境 |
| `--target dev` | 使用 `passwordlogin` build tag 编译 | 由 `LoginMethods` 配置 | 本地开发 |

密码登录必须同时满足：

1. 后端镜像使用 Dockerfile 的 `dev` target 构建。
2. `server/bbs-go.yaml` 中设置 `LoginMethods.password: true`。
3. 数据库中存在具有有效 bcrypt 密码的普通用户。

前后端只消费后端公开的 `loginMethods` 能力，不根据 `Env` 判断登录方式。默认生产镜像不包含密码登录控制器，即使误配 `password: true`，`/api/login/password` 也不会注册。

## 本地部署

### 1. 环境要求

- Linux
- Docker 24+
- Node.js `22.x`
- Yarn `1.22.22`
- `curl`、`awk`

仓库根目录的 `.nvmrc` 固定了 Node 主版本，建议使用 nvm：

```bash
nvm install
nvm use
node --version
yarn --version
```

### 2. 创建 Docker 网络

MySQL、MinIO 和后端通过容器名互相访问：

```bash
docker network inspect bbs-net >/dev/null 2>&1 || docker network create bbs-net
```

### 3. 启动 MySQL

本地示例密码仅用于开发，不要用于服务器：

```bash
docker run -d \
  --name bbs-mysql \
  --network bbs-net \
  --restart unless-stopped \
  -e MYSQL_ROOT_PASSWORD=bbs123456 \
  -e TZ=Asia/Shanghai \
  -v bbs-mysql-data:/var/lib/mysql \
  mysql:5.7 \
  --character-set-server=utf8mb4 \
  --collation-server=utf8mb4_unicode_ci
```

等待数据库就绪：

```bash
docker exec -e MYSQL_PWD=bbs123456 bbs-mysql \
  mysqladmin -uroot --silent --wait=30 ping
```

应输出：

```text
mysqld is alive
```

### 4. 初始化数据库

`server/initDB.md` 的“直接复制版”只包含通用建库、建表和系统数据。执行：

```bash
awk '
  found && /^```sql/{sql=1; next}
  sql && /^```/{exit}
  sql{print}
  /^## 直接复制版/{found=1}
' server/initDB.md |
docker exec -e MYSQL_PWD=bbs123456 -i bbs-mysql \
  mysql -uroot --default-character-set=utf8mb4
```

本地开发再单独导入开发账号；服务器部署不要执行该文件：

```bash
docker exec -e MYSQL_PWD=bbs123456 -i bbs-mysql \
  mysql -uroot --default-character-set=utf8mb4 < server/initDB.dev.sql
```

验证初始化结果：

```bash
docker exec -e MYSQL_PWD=bbs123456 bbs-mysql \
  mysql -uroot -D bbsgo_db \
  -e "SELECT id, username, roles, status FROM t_user; SHOW TABLES;"
```

开发种子账号为：

```text
用户名：admin
密码：123456
角色：高管
```

如果 `docker exec` 提示容器未运行，先检查：

```bash
docker ps -a --filter name=bbs-mysql
docker logs --tail 200 bbs-mysql
```

### 5. 启动 MinIO

```bash
docker run -d \
  --name bbs-minio \
  --network bbs-net \
  --restart unless-stopped \
  -p 9000:9000 \
  -p 9001:9001 \
  -e MINIO_ROOT_USER=bbsminio \
  -e MINIO_ROOT_PASSWORD=bbsminio123 \
  -v bbs-minio-data:/data \
  minio/minio server /data --console-address ":9001"
```

MinIO 控制台：`http://localhost:9001`。后端首次启动时会自动创建 `qscbbsbucket`。

### 6. 启动 Redis

“当前在线”功能使用 Redis 保存短期在线状态。Redis 不可用时，后端仍可正常登录和浏览，但首页会自动隐藏“当前在线”板块。

首页侧栏默认折叠显示前 10 位在线用户，点“还有 N 人在线”可展开完整名单。服务端单次快照最多携带 200 位用户（见 `server/services/presence_service.go` 中的 `maxOnlineUsersReturned`），在线人数超过该值时列表底部会标注“仅显示前 200 位”，同时按 Redis 中的真实在线人数显示“还有 N 人在线”。

```bash
docker run -d \
  --name bbs-redis \
  --network bbs-net \
  --restart unless-stopped \
  -v bbs-redis-data:/data \
  redis:7-alpine \
  redis-server --appendonly yes
```

确认 Redis 已启动：

```bash
docker ps --filter name=bbs-redis
docker logs --tail 50 bbs-redis
```

### 7. 启动 Meilisearch

搜索使用独立的 [Meilisearch](https://www.meilisearch.com/) 服务。MySQL 仍然是唯一真值源，Meilisearch 只保存后端同步过去的索引，负责分词、倒排检索、过滤和排序；搜索命中的帖子会拿编号回 MySQL 取原文。

它属于可选组件：不启动也能跑，后端会自动回退到数据库 `LIKE` 查询。回退路径功能可用，但没有中文分词、错字容错和相关度排序，关键词命中量大时也更慢。

```bash
docker run -d \
  --name bbs-meilisearch \
  --network bbs-net \
  --restart unless-stopped \
  -e MEILI_MASTER_KEY=bbsmeili123456 \
  -e MEILI_ENV=development \
  -e MEILI_NO_ANALYTICS=true \
  -v bbs-meili-data:/meili_data \
  getmeili/meilisearch:v1.6
```

确认已启动：

```bash
docker ps --filter name=bbs-meilisearch
docker logs --tail 50 bbs-meilisearch
```

需要注意：

- 容器名必须是 `bbs-meilisearch`，要和下一步 `Search.Url` 里的主机名一致，并且和 MySQL、Redis、后端处于同一个 `bbs-net` 网络。
- `MEILI_MASTER_KEY` 必须和下一步 `Search.ApiKey` **完全一致**。不一致时 Meilisearch 会返回 401，后端每次搜索都会记录 `meilisearch query failed, fallback to database like` 并回退到数据库——页面看起来正常，实际没走搜索引擎。
- 索引数据保存在 `bbs-meili-data` volume 中，不删除就保留，重启容器不需要重新同步。
- 这里没有加 `-p 7700:7700`：后端通过 `bbs-net` 内部访问即可，不需要把搜索服务暴露到宿主机。
- 生产环境必须换成随机密钥；本地可以随便填，但不要把示例值直接带上线。

### 8. 创建本地后端配置

创建 `server/bbs-go.yaml`，内容如下：

```yaml
Env: dev
BaseUrl: http://localhost:3000/bbs2
Port: 8082
LogFile: /tmp/bbs-go.log
ShowSql: true
StaticPath: /tmp

LoginMethods:
  passport: false
  password: true

MinIO:
  # 这里只能写 host:port，不要带 http://
  Endpoint: bbs-minio:9000
  AccessKeyID: bbsminio
  SecretAccessKey: bbsminio123
  UserSSL: false
  BucketLocation: us-east-1

DB:
  Url: root:bbs123456@tcp(bbs-mysql:3306)/bbsgo_db?charset=utf8mb4&parseTime=True&loc=Local
  MaxIdleConns: 10
  MaxOpenConns: 50

# 当前代码启动时要求 OldDB 可连接；无旧 BBS 数据库时可暂时指向主库
OldDB:
  Url: root:bbs123456@tcp(bbs-mysql:3306)/bbsgo_db?charset=utf8mb4&parseTime=True&loc=Local

Redis:
  Url: bbs-redis:6379
  Password: ""
  DB: 0

Presence:
  AllowedOrigins:
    - http://localhost:3000
  TrustedProxies:
    - 127.0.0.1/32
    - 172.16.0.0/12
  MaxConnections: 10000
  MaxConnectionsPerIP: 200
  MaxConnectionsPerUser: 5

# 搜索（Meilisearch）。Enabled 为 false、或 Meilisearch 不可用时，自动回退到数据库 LIKE。
Search:
  Enabled: true
  # 必须和 Meilisearch 容器名一致，且两者在同一 Docker 网络内
  Url: http://bbs-meilisearch:7700
  # 必须和容器启动时的 MEILI_MASTER_KEY 完全一致
  ApiKey: bbsmeili123456
  Index: topics
  # filter：搜索结果只返回当前用户有权访问的主题；mark：返回全部命中主题，由渲染层隐藏无权内容
  PermissionMode: filter
  # 单页条数上限 / 最大页码 / 单次搜索最多扫描的候选条数
  MaxLimit: 50
  MaxPage: 100
  MaxScan: 1000
  TimeoutMs: 3000
  # 启动时异步全量重建新站索引
  SyncOnBoot: true
  # 是否同时同步旧 BBS 数据；旧库数据量大，默认关闭
  SyncOldBBS: false
```

挂载配置前务必确认它是文件：

```bash
test -f /root/BBS/server/bbs-go.yaml
```

如果源路径不存在，Docker 可能创建同名目录，后端会报 `read ./bbs-go.yaml: is a directory`。

关于搜索，本地部署时还需要注意：

- **首次搜索前要等索引同步完。** 后端启动后会异步全量重建索引（`SyncOnBoot: true`），帖子多时需要几十秒到几分钟，期间搜索结果可能不完整。查看进度：

  ```bash
  docker logs --tail 200 bbs-backend | grep 'search:'
  ```

  出现 `search: reindex all topics done, total=N` 表示同步完成。日志里出现 `meilisearch health check failed` 或 `fallback to database like` 说明当前走的是数据库回退路径，按本节末尾的排查项处理。
- **手动重建索引**：重启后端即可；或带 Admin 登录态调用 `POST http://localhost:8082/api/admin/search/reindex`。
- **未登录访客不能搜索**，必须登录后才返回结果。这是本项目的预期行为，不是故障。
- **旧帖默认搜不到**：`SyncOldBBS: false` 时不会同步旧 BBS 数据，需要搜旧帖就改成 `true` 并重启后端；旧库数据量大，首次同步会明显变慢。
- **改 `Search.Index` 等于换成一个空索引**，改完必须重新触发一次全量同步。
- **`Search` 段可以整段省略**：省略等价于 `Enabled: false`，搜索自动走数据库回退，已有的旧配置不改也能正常启动。
- **本地不需要搜索功能**时，把 `Search.Enabled` 设为 `false`、且不启动 `bbs-meilisearch` 容器即可，搜索自动走数据库 `LIKE` 回退。

### 9. 构建并启动开发后端

开发环境必须显式选择 `dev` target：

```bash
docker build --network=host --target dev \
  -t bbs-neo-backend-dev ./server
```

`--network=host` 用于规避部分 Linux 环境中 Docker build 的 DNS 问题；如果当前 Docker 不支持，可以去掉。

启动后端：

```bash
docker run -d \
  --name bbs-backend \
  --network bbs-net \
  --restart unless-stopped \
  -p 8082:8082 \
  -v /root/BBS/server/bbs-go.yaml:/bbs-go.yaml:ro \
  bbs-neo-backend-dev
```

检查启动日志：

```bash
docker logs --tail 200 bbs-backend
```

正常日志应包含：

```text
Now listening on: http://localhost:8082
```

### 10. 安装并启动 Site

请使用 Node 22 和 Yarn 1.22.22。依赖下载证书报错时切换到 Yarn 官方 registry，不要关闭 TLS 校验：

```bash
cd /root/BBS/site
yarn config set registry https://registry.yarnpkg.com
yarn install --frozen-lockfile
yarn dev
```

访问：`http://localhost:3000/bbs2/`。

前端依赖统一使用 Yarn 和各目录中的 `yarn.lock`；不要使用 npm 重新生成 `package-lock.json`。

### 11. 安装并启动 Admin

另开终端：

```bash
nvm use
cd /root/BBS/admin
yarn config set registry https://registry.yarnpkg.com
yarn install --frozen-lockfile
yarn serve
```

访问：`http://localhost:8080/bbsadmin/`。

如果提示 `vue-cli-service: not found`，说明 `yarn install` 没有成功完成；不要直接运行 `yarn serve`，先解决依赖下载错误并重新安装。

### 12. 本地验收

检查公开配置是否识别密码登录能力：

```bash
curl -sS http://localhost:8082/api/config/configs
```

响应中的 `loginMethods` 应为 `{"passport":false,"password":true}`。

验证账号密码：

```bash
curl -sS -X POST \
  -d 'username=admin' \
  -d 'password=123456' \
  http://localhost:8082/api/login/password
```

响应应包含 `"success":true`、`"username":"admin"` 和 token。浏览器中 Site 与 Admin 均可使用 `admin / 123456` 调试登录。

验证搜索可用（把 `<token>` 换成上一步响应里的 token）：

```bash
curl -sS -X POST \
  -H 'X-User-Token: <token>' \
  -d 'keyword=admin' \
  http://localhost:8082/api/search/topic
```

响应应为 `"success":true`；`results` 有命中时是数组，未命中时为空。要确认走的确实是 Meilisearch 而不是回退路径，检查后端日志里没有 `fallback to database like`：

```bash
docker logs --tail 200 bbs-backend | grep 'search:'
```

不带 `X-User-Token` 再请求一次，应返回空结果——未登录访客不可搜索。

最终地址：

| 服务 | 地址 |
| --- | --- |
| Site | `http://localhost:3000/bbs2/` |
| Site 登录页【仅开发环境】 | `http://localhost:3000/bbs2/user/signin` |
| Admin | `http://localhost:8080/bbsadmin/` |
| API | `http://localhost:8082/` |
| MinIO Console | `http://localhost:9001/` |
| Redis | 仅在 `bbs-net` Docker 网络内提供 `bbs-redis:6379` |
| Meilisearch | 仅在 `bbs-net` Docker 网络内提供 `bbs-meilisearch:7700` |

### 13. 本地停止与重启

终止 `yarn dev` 和 `yarn serve` 后，停止应用容器，不删除数据卷：

```bash
docker stop bbs-backend bbs-meilisearch bbs-redis bbs-minio bbs-mysql
```

重新启动：

```bash
docker start bbs-mysql bbs-minio bbs-redis bbs-meilisearch bbs-backend
```

然后再重新启动 Site 和 Admin。

MySQL、MinIO、Redis 和 Meilisearch 的索引分别保存在 `bbs-mysql-data`、`bbs-minio-data`、`bbs-redis-data`、`bbs-meili-data` volume 中。除非确定不需要数据，不要删除这些 volume；删掉 `bbs-meili-data` 后需要重启后端重新同步索引，搜索结果才会恢复完整。

如果已经登录但首页没有显示“当前在线”板块，先检查 Redis 和后端日志：

```bash
docker ps --filter name=bbs-redis
docker logs --tail 50 bbs-redis
docker logs --tail 100 bbs-backend
```

若后端日志出现以下内容，表示 Redis 尚未运行、容器不在 `bbs-net` 网络中，或者配置中的 `Redis.Url` 不正确：

```text
Redis unavailable; online presence is disabled
```

被拒绝的 WebSocket 连接会在后端留下 `presence:` 开头的告警日志，按原因区分：`连接凭证校验失败`（凭证过期、重复使用，或页面停留过久）、`WebSocket 登录态校验失败`（登录 cookie 与凭证中的用户不一致）、`WebSocket 连接被拒绝`（触发了 `Presence` 里的并发上限，日志中会带上具体数值）、`Origin 校验失败`（`AllowedOrigins` 与访问域名不匹配）。日志同时记录 `user_id` 和 `ip_hash`，`ip_hash` 是 IP 的不可逆哈希，用于统计同源连接而不会落盘明文地址。

修复后后端会定期重新连接 Redis，前端也会使用指数退避自动重连 WebSocket；通常等待最多 30 秒或刷新页面即可恢复，无需重启整个 BBS。

## 微信小程序开发

小程序端位于 [`miniprogram/`](miniprogram/)，使用 `uni-app + Vue 3 + TypeScript + Vite`，不依赖 `site/` 的 Nuxt 运行时；当前骨架包含首页话题列表、话题详情、登录页（密码登录）和 Passport 登录页。

### 1. 开发环境

1. 安装 Node.js 22 和 Yarn 1.22.22。
2. 安装[微信开发者工具](https://developers.weixin.qq.com/miniprogram/dev/devtools/download.html)。
3. 进入 `miniprogram/` 安装依赖：

```bash
cd /root/BBS/miniprogram
yarn install
```

4. 在 `miniprogram/src/manifest.json` 中把 `mp-weixin.appid` 替换为自己的小程序 AppID；没有 AppID 时可以先使用测试号或 `touristappid` 预览页面。
5. 根据环境修改 `miniprogram/src/config/env.ts`：

```ts
export const env = {
  apiBaseUrl: 'https://你的域名/bbs2',
  passportUrl: 'https://www.qsc.zju.edu.cn/passport/v4/static/index.html#/login',
  passportBridgeUrl: 'https://你的域名/bbs2/miniprogram/passport',
}
```

本地后端调试时，真机不能使用 `localhost`，应改成运行后端电脑的局域网 IP，例如 `http://192.168.1.10:8082`。微信开发者工具中可以暂时勾选“详情 → 本地设置 → 不校验合法域名、web-view 业务域名、TLS 版本以及 HTTPS 证书”。正式发布不能依赖这个选项。

### 2. Passport 登录说明

网页端 Passport 登录依赖浏览器 Cookie：Passport 把 `SESSION_TOKEN` 放在回调 URL，网页端再调用 `/api/login/signin` 换取 BBS `userToken`。小程序的 `wx.request` 不会共享网页 Cookie，因此不能直接复用网页端的登录回调。

当前骨架使用以下兼容流程：

```text
小程序 web-view
  -> Passport 登录
  -> /bbs2/miniprogram/passport?SESSION_TOKEN=...
  -> Site bridge 调用 /api/login/signin
  -> wx.miniProgram.postMessage({ token, user })
  -> 小程序保存 userToken，并用 X-User-Token 请求 API
```

桥接页面是 [`site/pages/miniprogram/passport.vue`](site/pages/miniprogram/passport.vue)，网页端的全局 `checkSession` 中间件已为该路径跳过自动消费逻辑。部署 Site 后，必须把 `passportBridgeUrl` 配置成实际可访问的 HTTPS 地址，并在微信公众平台配置：

- request 合法域名：API 域名
- uploadFile/downloadFile 合法域名：文件上传和预览域名
- web-view 业务域名：Site 域名

如果不部署 Site bridge，Passport 登录页可以打开，但登录成功后无法把 `userToken` 返回给小程序。`SESSION_TOKEN` 和 Passport 的任何密钥都不要写入小程序代码。

### 3. 用本地后端调试（密码登录）

开发包和发布包连的是不同后端，切换由构建模式决定，不需要手工改地址（见 [`miniprogram/src/config/env.ts`](miniprogram/src/config/env.ts)）：

| 命令 | 产物 | `apiBaseUrl` |
| --- | --- | --- |
| `yarn dev:mp-weixin` | `dist/dev/mp-weixin/` | `http://127.0.0.1:8082`（本地后端） |
| `yarn build:mp-weixin` | `dist/build/mp-weixin/` | 线上地址 |

后端在 WSL 里跑时，Windows 侧的 `127.0.0.1:8082` 会自动转发到 WSL 内监听的端口，不需要额外做端口映射。

本地后端要用密码登录，必须同时满足两个条件：

1. 后端镜像用 Dockerfile 的 `dev` target 构建（与前面「构建并启动开发后端」相同；不装 Go 也能用，编译在镜像里完成）。`dev` target 带 `passwordlogin` build tag；默认的 `prod` target 不含密码登录控制器，`/api/login/password` 不会注册（返回 404）：

```bash
cd /root/BBS
docker build --network=host --target dev -t bbs-neo-backend-dev ./server
```

换成 dev 镜像后重建容器（`-v` 挂载的 `bbs-go.yaml` 里 `LoginMethods.password` 要为 `true`）：

```bash
docker rm -f bbs-backend
docker run -d --name bbs-backend --network bbs-net --restart unless-stopped \
  -p 8082:8082 \
  -v /root/BBS/server/bbs-go.yaml:/bbs-go.yaml:ro \
  bbs-neo-backend-dev
```

改了后端代码后要重新 `docker build --target dev` 并重建容器，否则跑的还是旧二进制。应用容器是无状态的，`docker rm -f bbs-backend` 重建不会丢数据（数据在 MySQL/MinIO 的 volume 里）。

2. `bbs-go.yaml` 中 `LoginMethods.password: true`，见前面「创建本地后端配置」。

登录页会先请求 `/api/config/configs`，按后端返回的 `loginMethods` 决定显示密码表单还是 Passport 按钮——和 Site、Admin 的判定方式一致。本地配置下应看到密码表单，用开发种子账号 `admin / 123456` 登录。

开发者工具里还需要「详情 → 本地设置」勾选**不校验合法域名、web-view 业务域名、TLS 版本及 HTTPS 证书**，否则 `http://127.0.0.1:8082` 会被拦截。真机预览不能用 `127.0.0.1`：把 `DEV_API_BASE_URL` 改成运行后端的电脑局域网 IP（例如 `http://192.168.1.10:8082`），并确保后端监听 `0.0.0.0`、Windows 防火墙放行该端口。

### 4. 调试流程

```bash
# 启动后端（另见前面的本地部署章节）
cd /root/BBS/server
go run . -config ./bbs-go.yaml

# 启动 Site，提供 Passport bridge
cd /root/BBS/site
yarn dev
```

启动小程序端：

```bash
cd /root/BBS/miniprogram
yarn dev:mp-weixin
```

命令会生成 `dist/dev/mp-weixin/`，在微信开发者工具中打开该目录并点击“编译”。首页会请求 `/api/topic/topics`；点击“登录”进入登录页，密码登录走 `/api/login/password`，Passport 登录走 `pages/passport/passport` 的 web-view。登录成功后`userToken` 会保存到本地 Storage，后续请求通过 `X-User-Token` 发送。

也可以直接使用 H5 预览：

```bash
yarn dev:h5
```

发布微信小程序代码：

```bash
yarn build:mp-weixin
```

产物位于 `dist/build/mp-weixin/`，再导入微信开发者工具上传审核。

当前小程序请求层会把 POST 数据编码为 `application/x-www-form-urlencoded`，这是因为现有 Go API 使用表单参数解析。帖子内容的 HTML 目前使用原生 `rich-text` 展示；正式实现图片、评论、发帖和上传时，应增加 `mp-html` 等富文本组件，并处理后端返回的 `data-src` 图片属性。

### 5. 当前范围与后续工作

- 已有：uni-app 基础工程、TypeScript 请求封装、token 持久化、话题列表、话题详情、密码登录、Passport web-view bridge。
- 待实现：评论、发帖、编辑、点赞、收藏、文件上传、个人中心。
- 在线用户功能暂未接入；它依赖带 Origin 校验的 WebSocket 和 Redis，需要单独适配小程序 `wx.connectSocket`。

## 服务器部署

以下流程以 Linux、域名 `www.qsc.zju.edu.cn`、代码目录 `/opt/bbs-neo` 为例。生产环境只启用 Passport，所有示例占位符都必须替换，不能直接作为密码使用。

### 1. 准备运行环境

安装 Docker 24+、Nginx、Node.js `22.x` 和 Yarn `1.22.22`，然后拉取代码：

```bash
git clone git@github.com:dreamem0ra1n/BBS.git /opt/bbs-neo
cd /opt/bbs-neo
git checkout fix/resource-leaks-causing-500-502
```

服务器应通过防火墙只公开 `80/443`。MySQL、MinIO 和后端端口只供本机或 Docker 内部网络访问。

### 2. 创建生产凭据文件

不要把真实密码写进仓库或命令行历史。创建仅 root 可读的 Docker 环境文件：

```bash
sudo install -d -m 700 /etc/bbs-neo
sudo install -m 600 /dev/null /etc/bbs-neo/mysql.env
sudo install -m 600 /dev/null /etc/bbs-neo/minio.env
```

编辑 `/etc/bbs-neo/mysql.env`：

```dotenv
MYSQL_ROOT_PASSWORD=<强随机数据库密码>
MYSQL_APP_PASSWORD=<强随机应用数据库密码>
TZ=Asia/Shanghai
```

编辑 `/etc/bbs-neo/minio.env`：

```dotenv
MINIO_ROOT_USER=<随机访问密钥>
MINIO_ROOT_PASSWORD=<强随机对象存储密钥>
```

可用 `openssl rand -base64 32` 生成随机值。不要提交 `/etc/bbs-neo/` 下的文件。

### 3. 启动生产基础设施

```bash
docker network inspect bbs-net >/dev/null 2>&1 || docker network create bbs-net

docker run -d \
  --name bbs-mysql \
  --network bbs-net \
  --restart unless-stopped \
  --env-file /etc/bbs-neo/mysql.env \
  -v bbs-mysql-data:/var/lib/mysql \
  mysql:5.7 \
  --character-set-server=utf8mb4 \
  --collation-server=utf8mb4_unicode_ci

docker run -d \
  --name bbs-minio \
  --network bbs-net \
  --restart unless-stopped \
  --env-file /etc/bbs-neo/minio.env \
  -v bbs-minio-data:/data \
  minio/minio server /data --console-address ":9001"

docker run -d \
  --name bbs-redis \
  --network bbs-net \
  --restart unless-stopped \
  -v bbs-redis-data:/data \
  redis:7-alpine \
  redis-server --appendonly yes
```

生产环境不需要用 `-p` 暴露 MySQL、MinIO 和 Redis。确需访问 MinIO 控制台时，建议使用 SSH 隧道，不要把 `9001` 直接开放到公网。

### 4. 初始化生产数据库

只导入通用初始化 SQL，绝对不要导入 `server/initDB.dev.sql`：

```bash
set -a
. /etc/bbs-neo/mysql.env
set +a

awk '
  found && /^```sql/{sql=1; next}
  sql && /^```/{exit}
  sql{print}
  /^## 直接复制版/{found=1}
' server/initDB.md |
docker exec -e MYSQL_PWD="$MYSQL_ROOT_PASSWORD" -i bbs-mysql \
  mysql -uroot --default-character-set=utf8mb4

unset MYSQL_ROOT_PASSWORD MYSQL_APP_PASSWORD
```

初始化后为应用创建独立数据库账号，不要让后端使用 root。执行 `docker exec -it bbs-mysql mysql -uroot -p`，输入 root 密码后运行：

```sql
CREATE USER IF NOT EXISTS 'bbsgo'@'%' IDENTIFIED BY '<与 mysql.env 中 MYSQL_APP_PASSWORD 一致>';
GRANT ALL PRIVILEGES ON bbsgo_db.* TO 'bbsgo'@'%';
FLUSH PRIVILEGES;
```

### 5. 创建生产后端配置

复制示例并编辑。`server/bbs-go.yaml` 已被 Git 忽略，但仍需限制文件权限：

```bash
cp server/bbs-go.example.yaml server/bbs-go.yaml
chmod 600 server/bbs-go.yaml
```

生产配置至少确认以下内容，尖括号内容必须替换：

```yaml
Env: prod
BaseUrl: https://www.qsc.zju.edu.cn/bbs2
Port: 8082
LogFile: /tmp/bbs-go.log
ShowSql: false
StaticPath: /tmp

LoginMethods:
  passport: true
  password: false

MinIO:
  Endpoint: bbs-minio:9000
  AccessKeyID: <与 minio.env 一致>
  SecretAccessKey: <与 minio.env 一致>
  UserSSL: false
  BucketLocation: us-east-1

DB:
  Url: bbsgo:<与 mysql.env 中 MYSQL_APP_PASSWORD 一致>@tcp(bbs-mysql:3306)/bbsgo_db?charset=utf8mb4&parseTime=True&loc=Local
  MaxIdleConns: 10
  MaxOpenConns: 50

# 有旧 BBS 数据库时填写真实连接；没有时当前代码仍要求可连接，可暂时指向主库
OldDB:
  Url: bbsgo:<与 mysql.env 中 MYSQL_APP_PASSWORD 一致>@tcp(bbs-mysql:3306)/bbsgo_db?charset=utf8mb4&parseTime=True&loc=Local

Redis:
  Url: bbs-redis:6379
  Password: ""
  DB: 0

Presence:
  AllowedOrigins:
    - https://www.qsc.zju.edu.cn
  TrustedProxies:
    - 127.0.0.1/32
    - 172.16.0.0/12
  MaxConnections: 10000
  MaxConnectionsPerIP: 200
  MaxConnectionsPerUser: 5
```

### 6. 构建并启动生产后端

生产构建必须使用默认的 `prod` target，不能使用 `dev` target：

```bash
docker build --network=host --target prod \
  -t bbs-neo-backend:prod ./server

docker run -d \
  --name bbs-backend \
  --network bbs-net \
  --restart unless-stopped \
  -p 127.0.0.1:8082:8082 \
  -v /opt/bbs-neo/server/bbs-go.yaml:/bbs-go.yaml:ro \
  bbs-neo-backend:prod
```

检查 `docker logs --tail 200 bbs-backend`，确认数据库、OldDB 和 MinIO 均连接成功。

### 7. 构建 Site 和 Admin

```bash
cd /opt/bbs-neo/site
yarn config set registry https://registry.yarnpkg.com
yarn install --frozen-lockfile
NODE_ENV=production yarn build

cd /opt/bbs-neo/admin
yarn config set registry https://registry.yarnpkg.com
yarn install --frozen-lockfile
yarn build
```

Admin 是静态文件，将 `admin/dist/` 部署到 Nginx 的 `/bbsadmin/`。Site 是 SSR 服务，建议交给 systemd 管理。示例 `/etc/systemd/system/bbs-site.service`：

```ini
[Unit]
Description=BBS Nuxt Site
After=network.target

[Service]
Type=simple
User=bbs
WorkingDirectory=/opt/bbs-neo/site
Environment=NODE_ENV=production
Environment=HOST=127.0.0.1
Environment=PORT=3000
Environment=NODE_OPTIONS=--max-old-space-size=8192
ExecStart=/usr/bin/yarn nuxt start
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

先用 `command -v yarn` 确认 `ExecStart` 路径。若服务器尚无专用运行用户，先创建并授予 Site 目录权限：

```bash
sudo useradd --system --home /opt/bbs-neo --shell /usr/sbin/nologin bbs
sudo chown -R bbs:bbs /opt/bbs-neo/site
```

然后启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now bbs-site
sudo systemctl status bbs-site
```

### 8. 配置 Nginx

将 Admin 构建结果放到静态目录：

```bash
sudo install -d /var/www/html/bbsadmin
sudo cp -a /opt/bbs-neo/admin/dist/. /var/www/html/bbsadmin/
```

在 `www.qsc.zju.edu.cn` 的 HTTPS `server` 块中加入：

```nginx
# 在 nginx.conf 的 http 块中配置一次：
# map $http_upgrade $connection_upgrade {
#     default upgrade;
#     '' close;
# }

location ^~ /bbs2/api/ {
    client_max_body_size 32m;
    rewrite ^/bbs2/api/(.*)$ /api/$1 break;
    proxy_pass http://127.0.0.1:8082;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection $connection_upgrade;
    proxy_read_timeout 120s;
    proxy_send_timeout 120s;
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
}

location = /bbs2 {
    return 301 /bbs2/;
}

location /bbs2/ {
    proxy_pass http://127.0.0.1:3000;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
}

location = /bbsadmin {
    return 301 /bbsadmin/;
}

location /bbsadmin/ {
    root /var/www/html;
    try_files $uri $uri/ /bbsadmin/index.html;
}
```

本项目的生产前端与 Passport 回调地址使用 `www.qsc.zju.edu.cn`。更换域名时，还必须同步修改 Admin 生产环境变量、前端 Passport 跳转地址和静态资源地址，不能只修改 Nginx。

上传接口的请求体上限设为 `32m`，为后端允许的最大 30 MiB 文件及 multipart 表单留出空间。若未配置且未从上级继承更大的限制，Nginx 默认只允许 1 MiB 请求体，超过后会直接返回 HTTP 413，上传请求不会到达后端。若前面还有其他反向代理，也需要检查其上传大小限制。

检查并重载 Nginx：

```bash
sudo nginx -t
sudo systemctl reload nginx
```

### 9. 生产验收

```bash
curl -sS https://www.qsc.zju.edu.cn/bbs2/api/config/configs
curl -i -X POST https://www.qsc.zju.edu.cn/bbs2/api/login/password
```

第一条响应的 `loginMethods` 应为 `{"passport":true,"password":false}`；第二条必须返回 `404`。随后验证 Passport 登录、退出、发帖和文件上传。

生产日志排查命令：

```bash
docker logs --tail 200 bbs-backend
sudo journalctl -u bbs-site -n 200 --no-pager
sudo tail -n 200 /var/log/nginx/error.log
```

### 10. 上传失败：MinIO 存储空间不足

若上传接口返回 `Storage backend has reached its minimum free drive threshold. Please delete a few objects to proceed.`，说明 MinIO 存储盘的剩余空间已触及写入保护阈值。这不是图片大小限制；头像、评论图片和附件共用该存储，均可能上传失败。磁盘不必完全写满就会触发保护。

在实际承载上传的 MinIO 服务器上检查挂载位置、可用空间和日志（容器名按部署情况调整）：

```bash
docker inspect bbs-minio --format '{{range .Mounts}}{{println .Source .Destination}}{{end}}'
docker exec bbs-minio df -h /data
docker exec bbs-minio df -i /data
docker logs --tail 200 bbs-minio
```

优先扩容数据盘，或清理同一文件系统中已确认不再需要的日志、缓存与备份；若删除上传文件，应通过 BBS 文件管理确认文件未被引用，或通过 MinIO 的对象管理接口处理。不要直接删除 `/data` 下的存储文件、`.minio.sys` 或 Docker 数据卷，也不要降低写入保护阈值来掩盖容量不足。

恢复足够空间后，用原先失败的图片验证上传及预览。仅更新前端错误提示、调整 Nginx 上传限制或重启服务，不能解决存储容量不足。

通过 Admin 文件管理清理时，直接勾选文件并点击“删除选中文件”。后端在每次删除时检查用户头像、背景图、简介及主页、节点、友链、系统配置、帖子（含回复可见内容）、评论、文章、站内消息和生日祝福中的实际文件引用，包含仍保存在数据库中的软删除内容。没有引用才会删除；来源分类、历史关联 ID、托管标记和上传时间均不作为禁止删除的依据。因此，已替换且不再被引用的旧头像、背景图、历史文件都可以清理。

头像也会出现在列表中，可按来源筛选。来源标签仅表示记录的用途，不表示当前仍在使用；列表不会触发全表回填，也无需手动预核验。批量删除会跳过仍被引用或处理失败的文件，继续处理其他文件，最后显示成功数量和未删除原因。引用查询失败或超时时不会删除该文件。文件从对象存储删除后不能通过后台恢复；数据库以外的外链引用和未提交到服务器的草稿无法核验。

此文件管理修复需要部署后端和 Admin，请先更新后端，使实际引用核验生效，再更新 Admin。

生产模式下，后端每天服务器本地时间 03:00 自动清理前一个自然日 00:00（含）至当天 00:00（不含）上传、且删除时无实际引用的文件。每次分批读取 100 条并逐个复用 Admin 删除核验，适用于所有来源和托管状态；仍被引用的文件保留，单个失败会记录日志并继续，不会删除当天或更早日期上传的文件。任务不会在启动时立即执行，同一进程内不允许重叠运行。服务停机错过的日期以及更早的存量文件可通过 Admin 清理。

任务日志 `daily file cleanup completed` 包含扫描、删除、仍被引用、失败数量及已删除文件大小合计（`freedBytes`）；对象存储的版本保留等策略可能影响实际释放空间。服务器或容器时区决定“前一天”的日期边界，请按生产部署的时区配置核对执行时间。开发模式不启动该任务。
