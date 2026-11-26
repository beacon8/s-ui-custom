# S-UI Custom

基于 alireza0/s-ui 1.6.8 的定制版。包含 Ant Design Vue 管理界面、客户端并发连接与独立上下行限速、出站分页和路由规则编辑器，以及通过可信中转 PROXY protocol 识别真实客户端 IP、按用户统计与限制 IP 数的能力。前端源码位于 `frontend/`，随主仓库一并发布。

## Linux 安装

发行包面向 Debian 等 Linux 系统，提供 x86（386）和 x86-64（amd64）。在测试服务器以 root 身份使用 Bash。安装器根据 CPU 架构下载本仓库的发行包，核对 SHA256SUMS，安装到 `/usr/local/s-ui` 并启用服务。已有同路径安装会被更新，建议先备份数据库。

```bash
SUI_LANG=zhcn bash <(curl -fsSL https://raw.githubusercontent.com/beacon8/s-ui-custom/v1.7.2/install.sh) v1.7.2
```

安装完成后，按终端输出的地址和凭据登录。默认面板端口为 2095，路径为 `/app/`。管理菜单为 `s-ui`。

[下载发行包](https://github.com/beacon8/s-ui-custom/releases) · [上游后端](https://github.com/alireza0/s-ui) · [上游前端](https://github.com/alireza0/s-ui-frontend)

本次发布源自本地后端 6641e05 和前端 c438cc2。保留后端与前端原有授权，见 [LICENSE](LICENSE) 和 [frontend/LICENSE](frontend/LICENSE)。
