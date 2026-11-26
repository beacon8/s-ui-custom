# 分组订阅与独立客户端订阅｜开发与验收工作记录

日期：2026-11-26  
开发仓库：`/Users/yuzai/Tools/ip_api/s-ui`  
前端子模块：`frontend/`
Git 备份点：主仓库 `5ef1562`、前端 `dc15a1f`。
功能提交：前端 `ba7c0b3`、主仓库 `bb5dc7f`。

## 1. 开工前已读资料

- `AGENTS.md`：协作规范、构建与联调约定、开发记录；发布规则见
  `../s-ui-custom/AGENTS.md`。
- `.ua/knowledge-graph.json`：导入服务、`sub` 订阅服务、`LinkService`、
  `JsonService`、`ClashService`、设置页/客户端弹窗的现有调用链。
- [上次批量导入工作记录](import-nodes-worklog.md)：`importClientNodes` 的编号、
  事务、幂等、批次和核心应用/恢复。
- [API 文档](api-token.md)、[功能覆盖表](api-coverage.md)。
- 源码：`service/import_nodes.go`、`sub/subHandler.go`、`sub/subService.go`、
  `sub/jsonService.go`、`sub/clashService.go`、
  `frontend/src/layouts/modals/ClientImportNodes.vue`。

## 2. 可复用入口与本次扩展

- `SubService.getClientLinks` / `getSubsForClient`：统一单片客户端的可用性
  检查、禁用入站过滤、base64 编码和 User-Agent 格式。
- `JsonService.getOutboundsShared` / `GetJsonForClients`：复用单客户端出站
  生成，跨成员共享 `takenTags`，最后统一加默认 selector/urltest。
- `ClashService.GetClashForClients`：复用单客户端代理转换与 `ConvertToClashMeta`，
  跨成员去重名称后统一生成 `Proxy/Auto` 组。
- 导入事务：`writeBatch`、`applyAndRecord`、`EnrichImportResult`。
- 新增 `service/subscription.go`：AES-GCM 凭据、组/成员创建、轮换、撤销、
  重命名、迁移、链接生成。
- 新增 `database/model/subscriptions.go` 与 schema/备份清单。

## 3. 实现摘要

### 3.1 组与个人身份

- 每次正式导入创建一行 `subscription_groups`（独立 `SubID`、加密 Token、
  显示名、批次号）和每客户端一行 `subscription_members`（独立 `SubID`、
  加密 Token、稳定 `ClientId`）。
- 组/个人 ID 与批次号、客户端名、协议凭据无关；同名分组或同前缀新批次创建
  新组，不会继承旧链接。
- `ClientId` 唯一：删除后重建的同名客户是新 ID，不会继承旧个人链接。
- 成员按导入顺序（member id）读取；组刷新实时查当前客户端，删除/禁用/失效
  自动不返回。
- Token 用面板 `secret` 派生 AES-GCM 密钥加密保存；授权复制时解密。备份随
  settings 中的 `secret` 恢复，原链接继续有效。

### 3.2 订阅路由

- `GET/HEAD /sub/g/:id?token=`：校验组 Token（常量时间），聚合当前成员。
- `GET/HEAD /sub/c/:id?token=`：校验个人 Token，只返回本人。
- 默认格式：逐客户端取链接后统一 base64 一次；JSON/Clash/Happ 与旧路由一致；
  `format` 和 User-Agent 自动识别共用同一逻辑。
- 组 JSON/Clash：跨成员出站标签去重，生成一份完整配置。
- 组 headers：`Profile-Title` 用组名；`Profile-Update-Interval` 沿用设置；
  `Subscription-Userinfo` 仅在所有成员都有有限额度和到期时求和/取最早到期，
  否则省略。响应 `Cache-Control: no-store`。
- 旧 `/<用户名>`：有成员行的客户一律 400（GET/HEAD、JSON/Clash 都不泄露）；
  未迁移客户继续旧行为。

### 3.3 导入响应与面板

- `ImportResult.groupSubscription.subscriptionId`、`items[].subscriptionId`；
  授权创建/重放/重放应用响应带 `link`/`subscriptionLink`。
- 存储的批次结果只含 ID，`GetBatch` 不返回 Token/链接；令牌只在显式管理读
  接口解密。
- 面板导入结果顶部显示组名、组订阅 ID、「复制整组订阅」；每行显示个人订阅
  ID 和复制按钮。
- 客户端列表新增每行「复制个人订阅」按钮（旧客户首次点击即完成迁移）。
- 工具菜单新增「订阅组」：组列表、组详情、组/个人复制、轮换、撤销、恢复、
  改显示名、移除成员。

## 4. 涉及文件

- `database/model/subscriptions.go`、`database/schema.go`、`database/backup_test.go`
- `service/subscription.go`、`service/import_nodes.go`、`service/import_nodes_test.go`
- `sub/subService.go`、`sub/subHandler.go`、`sub/groupService.go`、
  `sub/jsonService.go`、`sub/clashService.go`、`sub/subscription_test.go`
- `util/genLink.go`（nil Addrs 生成链接修复）
- `api/apiService.go`、`api/apiHandler.go`、`api/apiV2Handler.go`、
  `api/import_e2e_test.go`
- 前端：`ClientImportNodes.vue`、`Subscriptions.vue`、`Clients.vue`、
  `locales/{en,zhcn,zhtw}.ts`
- `docs/api-token.md`、`docs/api-coverage.md`、本记录

## 5. 自动化测试

| 命令 | 结果 |
|---|---|
| `go test ./... -count=1` | ✅ 全部通过 |
| `go test ./sub -run TestGroupAndPersonalSubscriptionLinks -count=1` | ✅ 通过 |
| `go test ./api -run TestImportClientNodesAPI -count=1` | ✅ 通过 |
| `npm test`（frontend） | ✅ 10 文件 / 154 测试 |
| `npm run build` | ✅ 通过 |
| `go build -tags "$(tags_for test)"` | ✅ 通过 |

新增覆盖：组/个人默认链接、JSON/Clash 格式、交叉 ID/Token、缺失/错误 Token、
HEAD、旧路径 400、禁用成员 9/个人 404、个人轮换、组轮换、组撤销、改显示名、
加密落库与解密、导入响应 ID/链接、同键重放 ID 稳定、批次视图无 Token。

## 6. 隔离实例实测

环境：临时 DB，面板 `127.0.0.1:22195/panel`、订阅
`127.0.0.1:22196/sub/`；`S10-1-10` 组 10 个 SOCKS 客户，每人一个入站。

| 场景 | 预期 | 实际 |
|---|---|---|
| 组链接默认格式 | 10 个节点 | ✅ base64 解码 10 行 |
| 第 4 人个人链接 | 1 个节点 | ✅ 仅 `S10-004` |
| 组 JSON | 一份完整配置 | ✅ 13 outbounds（10 + 默认 3）；`sing-box check` 通过 |
| 个人 JSON | 一份完整配置 | ✅ 4 outbounds；`sing-box check` 通过 |
| 组 Clash | 10 个 proxies | ✅ `name: S10-*` 10 行 |
| HEAD 组 | 有效 200 / 错误 404 | ✅ 无正文，正确标题 |
| 交叉 ID/Token | 404 | ✅ 组 ID+个人 Token、个人 ID+组 Token 均 404 |
| 旧用户名路径 | 新客户 400 | ✅ GET 400、HEAD 400；JSON/Clash 400 |
| 禁用第 4 人 | 组 9，个人 404 | ✅ 恢复后个人 200 |
| 个人 Token 轮换 | 旧个人 404，组 200 | ✅ 组仍 10 人 |
| 组 Token 轮换 | 旧组 404，新组 200，个人 200 | ✅ |
| 组撤销 | 组 404，个人 200 | ✅ |
| 改显示名 | 名称变，链接不变 | ✅ `RP-renamed`，链接相同 |
| 同键重放 | 组/个人 ID 不变 | ✅ 10 个个人 ID 全部相同 |
| 面板重启 | 链接仍可用 | ✅ 组/个人 200 |
| 备份/恢复 | 链接仍可用 | ✅ 组/个人 200，旧路径仍 400 |
| 旧客户迁移 | 迁移前旧路径 200；启用后 400，新链接 200 | ✅ API 创建的 `legacy2` 验证 |
| Token 落库 | 非明文、可解密 | ✅ `token_enc` 不含明文；解密等于 URL Token |

`sing-box check` 作为实际代理客户端导入验证：组/个人 JSON 均可被 sing-box
客户端解析；未启动真实 TUN/代理链路。

## 7. 本次问题与修复

1. **事务内读取 secret 死锁**：`CreateImportSubscriptions` 在写事务内调用
   `GetSecret`，会等待自身写锁。改为在 `db.Begin()` 前加载 AES-GCM，事务内只
   加密。
2. **`LinkGenerator` 对 nil `Addrs` 返回空链接**：`json.Unmarshal(nil)` 报错，
   导致新建客户没有本地订阅链接。先判 `len(i.Addrs)>0`。
3. **组 JSON 重复出站标签**：无 addr 分支未走 `uniqueOutboundTag`，多个成员
   的同一入站标签冲突。改为共享 `takenTags` 去重。
4. **批次结果泄漏链接**：`applyAndRecord` 在 enrich 之后保存，把 Token 写进
   `import_batches.result`。改为先保存脱敏结果，再 enrich 响应；`GetBatch`
   只返回 ID。
5. **旧路径/Token 校验**：旧路由按 name 查询；新增订阅成员后必须拒绝。
   `legacySubscriptionBlocked` 在默认、JSON、Clash、HEAD 路径统一生效。
6. **面板导入结果未展示订阅**：新增组/个人 ID 与复制按钮；客户端列表增加
   复制个人订阅；新增订阅组管理弹窗。

## 8. 未验证项/边界

- 没有用 Clash/mihomo 或 Hiddify 真机导入并刷新；Clash 只做结构/代理数量检查，
  JSON 用本机 `sing-box check` 解析。未做真实 TUN/代理链路。
- 未在 HTTPS 订阅域名下验证；隔离实例为 HTTP。
- 386 未运行本功能测试；未做移动端订阅组弹窗完整交互回归。
- 未做大规模组（数百成员）性能与响应体上限测试。
- 旧客户迁移测试使用 API 创建的客户；直接 SQL 插入的异常行未纳入。
- 知识图谱与 AGENTS 收尾记录见提交后的更新。

## 9. 知识图谱更新

- 基线 `15cf4df`（1,522 节点 / 5,446 边 / 297 文件）；功能提交 `bb5dc7f`。
- `ARCHITECTURE_UPDATE`：16 个变化文件、7 个批次；新增
  `service/subscription.go` 23 节点、`sub/groupService.go` 12 节点、
  `database/model/subscriptions.go` 3 节点、`sub/subscription_test.go` 5 节点；
  `api/apiService.go`、`sub/{json,clash,subHandler,subService}.go` 重新生成语义。
- 结果：**1,596 节点 / 5,948 边 / 302 文件 / 9 架构层 / 12 步学习路径**；
  `meta.json` 指向 `bb5dc7f`，symbol validation 通过（无符号丢失）。
