# 2026-11-28 用户管理功能迭代工作记录

## 1. 范围

本日围绕「用户管理」完成以下改动：

1. 原「落地 IP」改为「出站地址」：不再探测，直接显示用户路由绑定的出站
   `server` 地址。
2. 列顺序、列宽、创建/到期合并列、描述列恢复并固定在最后。
3. 工具栏左对齐、用户搜索、多选批量删除、页脚分页优化。
4. 删除客户端时联动清理对应的 `auth_user` 路由、出站和订阅成员。
5. 修复手动命名出站（`socks-s28`）未被清理的问题，并清理现场残留。

## 2. 提交

- 前端子模块：
  - `e836a6f` style: keep description as the last client column
  - `1a0449c` feat: show configured outbound server address in client list
  - `9dc6c8b` feat: search clients and batch delete with selection
- 主仓库：
  - `472811b` / `113dffc` 列顺序/日期列/描述列
  - `07b0569` / `62dbbbf` 默认只读缓存（后废弃）
  - `1a0449c` / `8b3c604` 出站地址替换落地 IP 探测
  - `7ef87c4` 出站地址文档
  - `b7d7df8` 删除客户端联动清理
  - `0a6b935` 大批量删除按 400 个 ID 分块
  - `784b249` 支持任意出站 tag 清理
  - `ff39b6e` 订阅成员/空组清理
  - `ea1d6ef` / `0e17f14` 相关文档修正

## 3. 最终行为

### 3.1 出站地址

- 列名「出站地址」，宽度 190px。
- 通过当前生效 `route.rules`/`route.final` 找用户的出站 tag，再读该出站
  `server`/`server_port`/`type`。
- 页面打开即显示，不拨号、不缓存、不轮询、无「检测中/未检测」。
- 状态：`ok`、`direct`、`multi`（selector/urltest）、`none`（无 server 字段）、
  `split`、`unbound`。
- 链式代理显示该出站自己的 `server`；域名原样显示；selector/urltest 不展开成员。
- API：`GET /api/outboundAddresses`、`/apiv2/outboundAddresses`，
  参数 `users` JSON 数组。

### 3.2 列与页脚

- 顺序：启用 → 名称 → 组 → 入站管理 → 出站地址 → 创建时间/到期时间 → 在线 →
  连接数 → IP 数 → 流量 → 实时网速 → 操作 → 最后在线 → 描述。
- 创建/到期一列上下两个标签，`YYYY-MM-DD HH:mm:ss`；到期 0 显示「无限」。
- 速度 `↓ 下行 / ↑ 上行`；描述恢复在最后一列。
- 工具栏全部靠左；搜索按名称/组/描述/备注过滤。
- 多选列 + 批量删除(n) + 确认弹窗；页脚显示总数/筛选数/已选数，分页有
  快速跳页和每页条数。
- 1920px 下含 32px 多选列总宽 1680px，无横向滚动。

### 3.3 删除联动

- 单条 `del` 与批量 `delbulk` 共用 `ClientService.DeleteClients`。
- 删除客户端后：
  - 删除顶层、简单、`auth_user` 全部属于本次删除的路由（任意出站 tag，
    不要求 `out-<name>`）；
  - 删除这些路由引用、且不再被其它规则或 `route.final` 引用的出站；
  - `direct`/`block` 内置出站、共享规则/出站、final 出站保留；
  - 删除个人订阅成员；组内无成员时删除组订阅，共享组保留；
  - 清理入站客户端绑定；路由/出站有变化时事务提交后异步重启核心；
  - 响应带 `clients/inbounds/outbounds/config`，前端立即刷新。
- 大批量删除按 400 个 ID 分块，避免 SQLite 绑定变量上限。

## 4. 验证

- `go test ./... -count=1`：全部通过。
- 前端 `npm test`：12 文件 / 159 测试通过；`vue-tsc --noEmit`、`npm run build`
  通过。
- 隔离实例：
  - `socks-s28` 配置 `1.1.1.1:80`，用户绑定后 API 返回
    `{status:"ok",address:"1.1.1.1",port:80,type:"socks"}`，页面直接显示。
  - 两个导入用户搜索/全选/批量删除后：客户端 0 行，出站只剩 `direct`，路由只剩
    `sniff`/`hijack-dns`，核心仍在 22100 监听。
  - 15 个客户端时页脚 `共 15 条`、分页 `1-10 / 15`，表宽 1680px。
- 现场修复：用户手动删除 `WdkM3DF6` 后，旧实现因出站名为 `socks-s28` 未清理；
  已通过面板 API 删除残留路由/出站，2095 出站只剩 `direct`、路由只剩
  `sniff`/`hijack-dns`。

## 5. 当前运行状态

- 本机 2095：PID `63077`，前端产物 `assets/91f0a974162ecb71.js`。
- 数据库客户端 0，出站 `direct`，路由 `sniff`/`hijack-dns`。

## 6. 边界与后续

- 只删除顶层、简单 `auth_user` 路由；逻辑/反选规则不自动删，避免影响其他用户。
- 共享出站、共享订阅组、`route.final` 保留。
- 出站地址是配置的服务器地址，不保证等于外部服务看到的真实公网出口。
- 未在 386 运行本日功能；未做真实公网出口探测（已按需求移除）。
