# bbs-neo

## 说明

该项目魔改自 bbsgo，目标项目的相关信息可以查看`mlog.md`

## 部署与启动

### 组件列表

- MySQL
- Server
- Site
- Admin
- MinIO

MySQL 需要根据`./server/initDB.md`中的指导进行初始化。

MinIO 启动后相关配置需要写入 server 的配置文件。

Server, Site, Admin 都有对应的配置文件需要更改，也都有 Dockerfile 使用。