# bbs-neo

## 项目说明

该项目魔改自 [bbsgo](https://github.com/mlogclub/bbs-go)，由以下组件组成：

| 组件 | 目录 | 默认端口 | 说明 |
| --- | --- | --- | --- |
| Server | `server/` | `8082` | Go API 服务 |
| Site | `site/` | `3000` | Nuxt 2 SSR 前台，访问前缀 `/bbs2` |
| Admin | `admin/` | `8080` | Vue 2 管理后台，访问前缀 `/bbsadmin` |
| MySQL | - | `3306` | 主数据库；推荐使用 MySQL 5.7 |
| MinIO | - | `9000/9001` | 文件对象存储及管理控制台 |

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
- Node.js `16.20.2`
- Yarn `1.22.22`
- `curl`、`awk`

前端依赖较旧，建议使用 nvm：

```bash
nvm install 16.20.2
nvm use 16.20.2
node --version
yarn --version
```

若当前 shell 没有加载 nvm，可以临时使用：

```bash
export PATH=/root/.nvm/versions/node/v16.20.2/bin:$PATH
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

### 6. 创建本地后端配置

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
```

挂载配置前务必确认它是文件：

```bash
test -f /root/BBS/server/bbs-go.yaml
```

如果源路径不存在，Docker 可能创建同名目录，后端会报 `read ./bbs-go.yaml: is a directory`。

### 7. 构建并启动开发后端

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

### 8. 安装并启动 Site

不要混用 Node 24。依赖下载证书报错时切换到 Yarn 官方 registry，不要关闭 TLS 校验：

```bash
cd /root/BBS/site
yarn config set registry https://registry.yarnpkg.com
yarn install --frozen-lockfile
yarn dev
```

访问：`http://localhost:3000/bbs2/`。

`package-lock.json found` 只是仓库同时存在 npm lockfile 的警告；本项目本地流程统一使用 Yarn 和 `yarn.lock`。

### 9. 安装并启动 Admin

另开终端：

```bash
export PATH=/root/.nvm/versions/node/v16.20.2/bin:$PATH
cd /root/BBS/admin
yarn config set registry https://registry.yarnpkg.com
yarn install --frozen-lockfile
yarn serve
```

访问：`http://localhost:8080/bbsadmin/`。

如果提示 `vue-cli-service: not found`，说明 `yarn install` 没有成功完成；不要直接运行 `yarn serve`，先解决依赖下载错误并重新安装。

### 10. 本地验收

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

最终地址：

| 服务 | 地址 |
| --- | --- |
| Site | `http://localhost:3000/bbs2/` |
| Site 登录页【仅开发环境】 | `http://localhost:3000/bbs2/user/signin` |
| Admin | `http://localhost:8080/bbsadmin/` |
| API | `http://localhost:8082/` |
| MinIO Console | `http://localhost:9001/` |

### 11. 本地停止与重启

终止 `yarn dev` 和 `yarn serve` 后，停止应用容器，不删除数据卷：

```bash
docker stop bbs-backend bbs-minio bbs-mysql
```

重新启动：

```bash
docker start bbs-mysql bbs-minio bbs-backend
```

然后再重新启动 Site 和 Admin。

MySQL 和 MinIO 数据分别保存在 `bbs-mysql-data`、`bbs-minio-data` volume 中。除非确定不需要数据，不要删除这两个 volume。

## 服务器部署

以下流程以 Linux、域名 `www.qsc.zju.edu.cn`、代码目录 `/opt/bbs-neo` 为例。生产环境只启用 Passport，所有示例占位符都必须替换，不能直接作为密码使用。

### 1. 准备运行环境

安装 Docker 24+、Nginx、Node.js `16.20.2` 和 Yarn `1.22.22`，然后拉取代码：

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
```

生产环境不需要用 `-p` 暴露 MySQL 和 MinIO。确需访问 MinIO 控制台时，建议使用 SSH 隧道，不要把 `9001` 直接开放到公网。

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
location ^~ /bbs2/api/ {
    rewrite ^/bbs2/api/(.*)$ /api/$1 break;
    proxy_pass http://127.0.0.1:8082;
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
