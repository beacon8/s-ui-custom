# API 功能覆盖表

本表把面板上可操作的功能逐项追到 HTTP 接口。测试环境为隔离实例：临时数据库、
面板路径 `/panel/`、端口 `22095`、一个 `mixed` 测试入站；除非特别注明，外部
调用均**不带 Cookie**，只带 `Token` 头。

权限列：

- **会话**：浏览器 Cookie（`/api`，并带同源校验）。Token 不能调用。
- **管理员**：面板管理员会话，或该管理员签发的 API Token。外部调用身份为
  Token 所属管理员，审计记录该管理员。
- **Token 所有者**：只能管理自己的 Token、查看/重放自己的导入批次。

结果列：

- ✅ 实测：本次在隔离实例实际调用并核对结果/副作用。
- ✅ 自动：由本次新增的 Go/前端测试覆盖。
- ⚠ 未实测：路由和处理逻辑存在，但本次未制造对应副作用（原因写在括号内）。
- — 会话用途：Token 不提供。

## 会话与基础数据

| 面板入口 | 操作 | 方法 | 路径 | 参数 | 权限 | 结果 |
|---|---|---|---|---|---|---|
| 登录页 | 登录 | POST | `/api/login` | `user`、`pass` | 无（建立会话） | ✅ 登录后 Cookie 可读设置 |
| 侧栏「退出登录」 | 退出 | GET | `/api/logout` | — | 会话 | ✅ 退出后 `/api/settings` 返回 401（XHR） |
| 全站 10 秒轮询 | 增量/全量数据 | GET | `/api/load` | `lu` | 管理员 | ✅ 面板与 API 均实测 |
| 用户管理 | 客户端列表 | GET | `/api/clients` | `id` 可选 | 管理员 | ✅ 200，27 条 |
| 入站管理 | 入站列表 | GET | `/api/inbounds` | `id` 可选 | 管理员 | ✅ 200 |
| 出站管理 | 出站列表 | GET | `/api/outbounds` | — | 管理员 | ✅ 200 |
| 节点管理 | 服务列表 | GET | `/api/services` | — | 管理员 | ✅ 200 |
| 节点管理 | 端点列表 | GET | `/api/endpoints` | — | 管理员 | ✅ 200 |
| TLS 设置 | TLS 列表 | GET | `/api/tls` | — | 管理员 | ✅ 200 |
| 路由/基础信息/DNS | 基础配置 | GET | `/api/config` | — | 管理员 | ✅ 200 |
| 设置 | 设置表单 | GET | `/api/settings` | — | 管理员 | ✅ 200 |
| 管理员 | 管理员列表 | GET | `/api/users` | — | 管理员 | ✅ 200 |
| 统计弹窗 | 流量统计 | GET | `/api/stats` | `resource`、`tag`、`limit`、`start`、`end` | 管理员 | ✅ 200 |
| 主页 | 系统/核心状态 | GET | `/api/status` | `r`（sys/cpu/mem/swap/disk/net/db…） | 管理员 | ✅ 200 |
| —（由 `load` 提供） | 在线用户 | GET | `/api/onlines` | — | 管理员 | ⚠ 路由存在，面板未直接调用 |
| 会话弹窗 | 会话列表 | GET | `/api/sessions` | `resource`、`tag` | 管理员 | ✅ 200（无会话） |
| 客户端 IP 明细 | 用户 IP | GET | `/api/userIps` | `user` | 管理员 | ✅ 200 |
| 用户列表实时网速 | 读写快照 | GET/POST | `/api/userRates` | `users`（JSON 数组） | 管理员 | ✅ 200 |
| 日志弹窗 | 日志 | GET | `/api/logs` | `c`、`l` | 管理员 | ✅ 200 |
| 管理员「更改」 | 变更记录 | GET | `/api/changes` | `a`、`k`、`c` | 管理员 | ✅ 200 |
| TLS/端点表单 | 生成密钥 | GET | `/api/keypairs` | `k`、`o` | 管理员 | ✅ 200（x25519） |

## 客户端

| 面板入口 | 操作 | 方法 | 路径 | 参数 | 权限 | 结果 |
|---|---|---|---|---|---|---|
| 用户管理「添加」 | 新建客户端 | POST | `/api/save` | `object=clients`、`action=new`、`data` | 管理员 | ✅ 既有路径；导入新建已实测 |
| 客户端行「编辑」 | 编辑客户端 | POST | `/api/save` | `object=clients`、`action=edit`、`data` | 管理员 | ⚠ 本次用导入/批量编辑验证，未单独点单条编辑 |
| 工具「批量添加」 | 批量新建 | POST | `/api/save` | `object=clients`、`action=addbulk`、`data` | 管理员 | ⚠ 本次改用 `importClientNodes` 验证 |
| 工具「批量编辑」 | 批量编辑 | POST | `/api/save` | `object=clients`、`action=editbulk`、`data` | 管理员 | ✅ 按组给 UI-050/051 各加 10 GiB，刷新后 110 GiB |
| 客户端行「删除」 | 删除客户端 | POST | `/api/save` | `object=clients`、`action=del`、`data=id` | 管理员 | ✅ 删除 DELX-001，库中剩 0 |
| 工具「批量删除」 | 批量删除 | POST | `/api/save` | `object=clients`、`action=delbulk`、`data=[id]` | 管理员 | ✅ 删除 DELX-002 |
| 工具「重置所有流量」 | 全部流量归零 | POST | `/api/resetTraffic` | — | 管理员 | ✅ 200 |
| 会话弹窗 | 断开用户 | POST | `/api/closeSessions` | `u` | 管理员 | ✅ 200（该用户无在线会话） |
| 客户端「流量图表」 | 客户端统计 | GET | `/api/stats` | `resource=user`、`tag=用户名` | 管理员 | ✅ 200 |
| 客户端「IP 数」 | IP 明细 | GET | `/api/userIps` | `user` | 管理员 | ✅ 200 |
| 客户端「连接数/实时网速」 | 会话/速率 | GET | `/api/sessions`、`/api/userRates` | 见上 | 管理员 | ✅ 200 |
| 工具「按节点添加」 | **按节点批量创建** | POST | `/api/importClientNodes` | JSON 请求体 + `Idempotency-Key` | 管理员 | ✅ 单条/10 条/混合协议/预览/冲突/幂等/恢复均实测 |
| 结果面板「重新应用到内核」 | 重放已提交批次 | POST | `/api/reapplyImportBatch` | `batchId` | 批次所属管理员 | ✅ 端口冲突恢复后 applied，未重复创建 |
| 定位批次 | 批次状态 | GET | `/api/importBatch` | `batchId` | 批次所属管理员 | ✅ 200，返回脱敏请求摘要 |

## 入站、出站、路由

| 面板入口 | 操作 | 方法 | 路径 | 参数 | 权限 | 结果 |
|---|---|---|---|---|---|---|
| 入站管理「添加」 | 新建入站 | POST | `/api/save` | `object=inbounds`、`action=new`、`data` | 管理员 | ✅ 新建 `accept-in`（混合 22097） |
| 入站行「编辑」 | 编辑入站 | POST | `/api/save` | `object=inbounds`、`action=edit`、`data` | 管理员 | ✅ 临时入站改端口 22120→22121 |
| 入站行「删除」 | 删除入站 | POST | `/api/save` | `object=inbounds`、`action=del`、`data=tag` | 管理员 | ✅ 删除临时入站 |
| 入站/客户端 | 按用户更新入站 | POST | `/api/save` | 由 `clients`/`inbounds` 保存内部调用 | 管理员 | ✅ 导入后实际登录成功，证明用户已进核心 |
| 出站管理「添加」/批量 | 新建出站 | POST | `/api/save` | `object=outbounds`、`action=new`、`data` | 管理员 | ✅ 导入创建 `out-US-*` 等；也建过测试出站 |
| 出站行「编辑」 | 编辑出站 | POST | `/api/save` | `object=outbounds`、`action=edit`、`data` | 管理员 | ⚠ 未单独编辑；导入写入路径已实测 |
| 出站行「删除」 | 删除出站 | POST | `/api/save` | `object=outbounds`、`action=del`、`data=tag` | 管理员 | ✅ 删除 `out-CF-002`，库中剩 0 |
| 出站行「测试」 | 出站连通性 | GET | `/api/checkOutbound` | `tag`、`link` | 管理员 | ✅ 200（direct） |
| 路由列表/JSON/Clash | 保存基础配置 | POST | `/api/save` | `object=config`、`action=set`、`data` | 管理员 | ✅ 原配置回存 200，核心重启后 Token 仍可用 |
| 路由列表 | 读取路由 | GET | `/api/config` | — | 管理员 | ✅ 200；导入规则插在 sniff/hijack-dns 后 |
| 出站「导入链接」 | 分享链接转出站 | POST | `/api/linkConvert` | `link` | 管理员 | ✅ SOCKS5 链接解析成功 |
| 批量出站「订阅转换」 | 订阅转出站 | POST | `/api/subConvert` | `link` | 管理员 | ✅ 无效地址返回业务错误（路由已到达） |

## TLS、服务、端点

| 面板入口 | 操作 | 方法 | 路径 | 参数 | 权限 | 结果 |
|---|---|---|---|---|---|---|
| TLS「添加」 | 新建 TLS | POST | `/api/save` | `object=tls`、`action=new`、`data` | 管理员 | ✅ 新建 `acceptance-tls` / `svc-tls` |
| TLS 行「编辑」 | 编辑 TLS | POST | `/api/save` | `object=tls`、`action=edit`、`data` | 管理员 | ⚠ 未单独编辑 |
| TLS 行「删除」 | 删除 TLS | POST | `/api/save` | `object=tls`、`action=del`、`data=id` | 管理员 | ✅ 删除后 tls 表 0 |
| TLS 出站表单 | 证书探测 | POST | `/api/getCertPing` | `domain`、`port` | 管理员 | ✅ 127.0.0.1:1 返回 connection refused |
| 服务管理「添加」 | 新建服务 | POST | `/api/save` | `object=services`、`action=new`、`data` | 管理员 | ⚠ 路由与校验实测（DERP 报 “missing certificate”），未产生副作用 |
| 服务行「删除」 | 删除服务 | POST | `/api/save` | `object=services`、`action=del`、`data=tag` | 管理员 | ✅ 对不存在/已删 `svc-accept` 返回成功 |
| 端点管理「添加」 | 新建端点 | POST | `/api/save` | `object=endpoints`、`action=new`、`data` | 管理员 | ⚠ 需真实 WireGuard/OpenVPN 参数，本次未建测试端点 |
| 端点行「删除」 | 删除端点 | POST | `/api/save` | `object=endpoints`、`action=del`、`data=tag` | 管理员 | ⚠ 同上 |

## 设置、管理员、Token

| 面板入口 | 操作 | 方法 | 路径 | 参数 | 权限 | 结果 |
|---|---|---|---|---|---|---|
| 设置「保存」 | 保存面板设置 | POST | `/api/save` | `object=settings`、`action=set`、`data` | 管理员 | ✅ 原设置回存 200 |
| 设置「重启面板」 | 重启面板 | POST | `/api/restartApp` | — | 管理员 | ✅ Token 调用后 5 秒内恢复，Token 继续可用 |
| 主页/设置「重启内核」 | 重启核心 | POST | `/api/restartSb` | — | 管理员 | ✅ 200 |
| 设置「停止/启动内核」 | 维护模式 | POST | `/api/maintenance` | `enable` | 管理员 | ✅ 停/启均实测；占用端口时启动报错 |
| 设置/管理员「API 令牌」 | Token 列表 | GET | `/api/tokens` | — | Token 所有者 | ✅ 只返回自己的，`****` 脱敏 |
| Token「添加」 | 新建 Token | POST | `/api/addToken` | `desc`、`expiry`（天） | Token 所有者 | ✅ 新建 32 位 Token 并立即生效 |
| Token 行「删除」 | 撤销 Token | POST | `/api/deleteToken` | `id` | Token 所有者 | ✅ 撤销后下一次请求 401 |
| 管理员行「编辑」 | 改用户名/密码 | POST | `/api/changePass` | `oldPass`、`newUsername`、`newPass` | 管理员 | ✅ Token 调用改密并改回；Token 仍可用；跨用户拒绝 |
| 管理员「更改」 | 审计记录 | GET | `/api/changes` | `a`、`k`、`c` | 管理员 | ✅ 200；导入写 `actor=admin, action=import` |

## 备份、更新、核心与统计

| 面板入口 | 操作 | 方法 | 路径 | 参数 | 权限 | 结果 |
|---|---|---|---|---|---|---|
| 备份与恢复「下载」 | 导出数据库 | GET | `/api/getdb` | `exclude` 可选 | 管理员 | ✅ 167,936 B，`integrity_check=ok`，含 Token/批次 |
| 备份与恢复「恢复」 | 导入数据库 | POST | `/api/importdb` | multipart `db` | 管理员 | ✅ 同库恢复后 Token 仍可用，客户端 29 条 |
| 备份与恢复 | 导出 sing-box 配置 | GET | `/api/singbox-config` | — | 管理员 | ✅ 返回原始 JSON 配置 |
| 更新弹窗 | 检查更新 | GET | `/api/checkUpdate` | — | 管理员 | ✅ 200（可能因网络返回空 latest） |
| 更新弹窗「更新」 | 面板内升级 | POST | `/api/update` | — | 管理员 | ✅ macOS 返回 “cannot be updated from the panel”（无副作用） |

## 外部 Token 鉴权

| 场景 | 方法 | 路径 | 结果 |
|---|---|---|---|
| 有效 Token | GET | `/apiv2/status` | ✅ 200 |
| 缺失 Token | GET | `/apiv2/status` | ✅ 401 `invalid token` |
| 错误 Token | GET | `/apiv2/status` | ✅ 401 |
| 过期 Token | GET | `/apiv2/status` | ✅ 401（下一次请求即失效） |
| 已撤销 Token | GET | `/apiv2/status` | ✅ 401（下一次请求即失效） |
| 跨管理员读 Token 列表 | GET | `/apiv2/tokens` | ✅ 只返回自己的 Token |
| 跨管理员撤销 Token | POST | `/apiv2/deleteToken` | ✅ 拒绝，目标 Token 仍可用 |
| 升级前签发的 Token | GET | `/apiv2/status` | ✅ 旧二进制签发，新二进制启动后仍 200 |
| 路由覆盖 | — | v1 vs v2 | ✅ `TestAPIv2CoversEveryPanelAction` 机械比对；仅 login/logout 允许会话专用 |

## 未实测项与原因

1. **服务/端点创建的真实副作用**：服务需要有效证书，端点需要真实
   WireGuard/OpenVPN 参数；本次只验证了 `save` 分发、校验错误和删除路径，未在
   隔离实例建立可用服务/端点。
2. **单条客户端编辑**：本次用批量编辑和导入覆盖了客户端写入；单条编辑沿用
   既有 UI 路径，未额外截图/对比。
3. **旧版 `addbulk`**：本次新入口使用 `importClientNodes`，未重复验证原有
   「批量添加」的独立编号逻辑。
4. **真实 TLS/Reality 出站流量**：导入只验证到配置生成与 sing-box 反序列化/
   启动；未用真实 Reality 服务端跑流量。SOCKS5 的路由与限速已用本地受控节点
   实测。
5. **386 运行时**：本次未在 386 目标运行；仅验证当前平台构建与测试。

## 分组订阅与个人订阅

| 面板/入口 | 操作 | 方法 | 路径 | 参数 | 权限 | 结果 |
|---|---|---|---|---|---|---|
| 导入结果顶部 | 组订阅 ID/复制整组链接 | POST | `/api/importClientNodes` | JSON + `Idempotency-Key` | 管理员 | ✅ 10 客户组链接返回 10 个节点 |
| 导入结果每行 | 个人订阅 ID/复制个人链接 | POST | 同上 | 同上 | 管理员 | ✅ 第 4 人个人链接只返回 1 个节点 |
| 订阅服务 | 组订阅 | GET/HEAD | `/sub/g/:id?token=` | `token`、`format`、User-Agent | 组 Token | ✅ 默认/JSON/Clash/Happ；缺/错/撤销 404 |
| 订阅服务 | 个人订阅 | GET/HEAD | `/sub/c/:id?token=` | `token`、`format`、User-Agent | 个人 Token | ✅ 只返回本人；禁用/撤销 404 |
| 订阅服务 | 旧用户名路径 | GET/HEAD | `/sub/:name` | — | 旧规则 | ✅ 有成员的新客户 400；未迁移老客户迁移前 200、迁移后 400 |
| 工具「订阅组」 | 组列表 | GET | `/api/subscriptionGroups` | — | 管理员 | ✅ 只返回 ID/名称/成员数，无 Token |
| 工具「订阅组」→查看 | 组详情/链接 | GET | `/api/subscriptionGroup` | `groupId` | 管理员 | ✅ 组链接 + 成员个人链接 |
| 客户端行「订阅」 | 旧客户显式启用/取链接 | POST | `/api/enableClientSubscription` | `clientId` | 管理员 | ✅ 迁移前旧路径 200，迁移后 400；个人链接 200 |
| 订阅组「轮换」 | 轮换组/个人 Token | POST | `/api/rotateSubscription` | `kind`、`id` | 管理员 | ✅ 旧个人链接 404，组及其余 9 人仍 200 |
| 订阅组「撤销」 | 撤销/恢复组或个人 | POST | `/api/revokeSubscription` | `kind`、`id`、`revoke` | 管理员 | ✅ 撤销组后组 404、个人仍 200 |
| 订阅组「改显示名」 | 改组名 | POST | `/api/renameSubscriptionGroup` | `groupId`、`name` | 管理员 | ✅ 名称变、原链接不变 |
| 订阅组成员「删除」 | 移除成员 | POST | `/api/removeSubscriptionMember` | `memberId` | 管理员 | ✅ 组少一人，个人链接失效 |
| 订阅服务 | 禁用成员后组刷新 | GET | `/sub/g/:id` | — | 组 Token | ✅ 10 → 9；该成员个人 404 |
| 订阅服务 | 跨组/跨人凭据 | GET | `/sub/g|c/...` | 组合 ID/Token | — | ✅ 组 ID+个人 Token、个人 ID+组 Token 均 404 |
| 订阅服务 | 响应缓存 | GET | `/sub/g|c/...` | — | 有效 Token | ✅ `Cache-Control: no-store` |
| 导入响应/批次 | 重放/重启/恢复 | POST/GET | `/importClientNodes`、`/importBatch` | 同键/批次 | 管理员 | ✅ 组/个人 ID 稳定；批次历史无 Token |
