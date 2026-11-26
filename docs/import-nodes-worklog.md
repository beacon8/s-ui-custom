# API Token 与按节点批量创建客户端｜开发与验收工作记录

日期：2026-11-26  
开发仓库：`/Users/yuzai/Tools/ip_api/s-ui`  
前端子模块：`frontend/`
Git 备份点：主仓库 `beffb26`（分支/标签 `backup/before-token-node-import`、
`backup-before-token-node-import`）；前端 `bf4af73`（同名分支/标签）。
功能提交：前端 `dc15a1f`，主仓库 `15cf4df`（含子模块指针、后端、测试和文档）。

## 1. 开工前已读资料

- `AGENTS.md`：Agent 强制协作规范、本地构建与联调约定、开发与变更工作记录、
  发布仓库规则。
- `.ua/knowledge-graph.json`（基线 `beffb26`，1,440 节点 / 5,013 边）：
  - 面板会话链：`web.initRouter` → `api.APIHandler.initRouter` → `checkLogin` →
    `GetLoginUser` → `ApiService` → 各业务 Service。
  - Token 链：`web` 挂载 `/apiv2` → `APIv2Handler.checkToken`/`findUsername` →
    `TokenInMemory`（`Token`、`Expiry`、`Username`）→ 同一 `ApiService`。
  - 客户端落库链：`Client.vue`/`Data.save` → `api/save` →
    `ConfigService.Save`（事务）→ `ClientService.Save` →
    `updateLinksWithFixedInbounds` → `InboundService.UpdateInboundsUsers` →
    `core.UpdateInboundUsers`；出站走 `OutboundService.Save` +
    `core.AddOutbound`；路由在基础 `config` 的 `route.rules` 中，保存后重启核心。
- 历史工作记录中与本次直接相关的条目：
  - 「设置页四个标签」「添加客户端弹窗基础页/配置页」「路由列表」「按用户 IP
    限制」「客户端限速第二轮验收」「大量客户端的备份、恢复与批量操作」。
  - `docs/rate-limit-worklog.md`、`docs/real-ip-worklog.md` 中关于
    `upMbps/downMbps`、`maxConns`、`ipLimit` 字段保留/同步、批量边界、核心
    热更新的部分。
- 当前工作区：开发仓库干净（`beffb26`），前端子模块干净（`bf4af73`）；
  2095 端口有既有面板进程，未用于本次验收。

## 2. 可复用的实现入口

- `ClientService.Save(tx, "addbulk", data, hostname)`：整批先校验名称/速率，
  再统一设置 `createdAt`、生成身份、重建本地链接、写库。导入沿用它写客户端。
- `ClientService.updateLinksWithFixedInbounds`：按入站列表生成
  `socks5://<name>:<password>@host:port` 等本地链接，并保留 external/sub。
- `InboundService.UpdateInboundsUsers` / `core.UpdateInboundUsers`：把新客户端
  写进入站 `users` 并热更新，失败时回退为移除/重建入站。
- `OutboundService.Save` / `core.AddOutbound`：出站的持久化与热加载；导入直接
  写同一 `model.Outbound` 结构。
- `ConfigService.GetConfig` / `SettingService.SaveConfig`：组装完整 sing-box
  配置与保存基础配置；`ConfigService.RestartCore` 保证提交后统一应用。
- `util.GetOutbound`：SOCKS5/HTTP/HTTPS/VLESS/VMess 等分享链接解析。
- `ConfigService` 的 `_txlock=immediate`（`database.OpenDB`）让导入事务一开始
  就持有写锁，配合事务内名称复核，保证并发抢同一编号范围只有一个成功。
- 新加 `core.ValidateConfig`、`core.ValidateOutbound`：用运行核心同一套协议
  注册表做反序列化/选项校验，不监听、不拨号。

## 3. 与本次有关的历史问题（本次再次核对）

1. **Token 路由遗漏**：改动前 v2 缺少 `changePass`、`tokens`、`addToken`、
   `deleteToken`、`checkUpdate`、`update`、`singbox-config`。本次补齐并用
   `api/route_coverage_test.go` 直接解析两个 handler 源码做机械比对，面板动作
   漏加 v2 时测试失败；`login`/`logout` 明确保持会话专用。
2. **Token 过期/撤销**：内存 Token 列表每次请求检查 `Expiry`；新增/删除/改密后
   `ReloadTokens()`。历史注释记录的“遍历切片时删除导致跳过下一个 Token”已不
   复现；本次验证过期/撤销下一次请求即 401。
3. **Token 归属**：`GetUserTokens`、`AddToken`、`DeleteToken` 均使用调用者
   身份；`LoadTokens` 对 `User == nil` 的孤儿 Token 直接跳过。跨管理员删除
   返回“no such token”，目标 Token 仍可用。
4. **审计身份**：v2 把 `checkToken` 校验出的用户名放进 context，再传给
   `Save`/`ChangePass`/Token 管理/导入；导入额外写
   `changes(actor=实际管理员, key=clients, action=import)`。
5. **批量写入边界**：SQLite 变量上限。已有值冲突查询按 400 个分块；
   1,000 节点上限。
6. **事务与核心更新**：原 `ConfigService.Save` 对 clients 只同步限额，对
   config 才异步重启；导入不能复用“分别保存”的路径，否则可能出现半批。
   本次在**一个事务**内写 clients/inbounds 绑定/outbounds/route/批次/审计，
   提交后同步调用 `RestartCore`，失败写回批次状态并可重放。
7. **客户端名唯一性**：数据库没有唯一索引，`validateClientName` 是应用层校验；
   本次在事务内再次调用，并在预检阶段对比现有名称与出站标签。
8. **批量字段缺省**：`upMbps`/`downMbps`、`ipLimit` 的“缺字段保留、显式 0 清除”
   历史规则仍适用于单条/批量编辑；导入是新建，显式写 0。
9. **备份表清单**：`database/schema.go` 是迁移与备份的唯一清单。新表
   `import_batches` 已加入，并更新 `database/backup_test.go` 的 fixture，
   避免“迁移了但没备份”。
10. **前端不在 Understand-Anything 扫描范围**：前端改动仍需构建与测试，但
    知识图谱只反映 Go 侧。

## 4. 实现摘要

### 4.1 Token API

- `api/apiHandler.go`：补齐 `importClientNodes`、`reapplyImportBatch`、
  `importBatch`；`changePass`/`addToken`/`deleteToken` 传入会话身份并重载
  Token。
- `api/apiV2Handler.go`：补齐 `changePass`、`tokens`、`addToken`、
  `deleteToken`、`checkUpdate`、`update`、`singbox-config`、导入/重放/批次；
  未认证改为 **401**，身份存入 context。
- `api/apiService.go`：`ChangePass`、`GetTokens`、`AddToken`、`DeleteToken`
  改为显式接收 actor；新增导入、重放、批次查询与 HTTP 错误映射。
- `service/user.go`：`LoadTokens` 跳过无用户行的 Token。
- `frontend/src/layouts/modals/Token.vue`：显示可复制的 API 基础地址
  （`window.BASE_URL` + `apiv2`，包含实际面板路径）。
- `frontend/src/views/Settings.vue`：设置工具栏增加「API 令牌」入口，复用同一
  模态框；与「管理员」入口管理同一批 Token。

### 4.2 `importClientNodes`

- 新文件 `service/import_nodes.go`：请求解析、逐项预检、编号/分组、客户端/
  出站/路由构造、事务提交、幂等、核心应用与重放。
- 新文件 `core/validate.go`：进程内配置/出站校验，不启动监听。
- 新表 `database/model/import_batch.go` + `database/schema.go`：
  `batch_id`、调用者、幂等键、内容哈希、状态、脱敏请求、结果、时间。
- 编号：`endId = startId + 有效节点数 - 1`；客户端名/出站标签补零；组名
  `prefix-startId-endId`。`idWidth` 最少 3。
- 出站：`out-<客户端名>`；每个客户端独立生成密码/UUID/本地链接，支持
  `clientDefaults.config` 的非身份覆盖和 external/sub 链接。
- 路由：每个客户端一条
  `{"auth_user":["<名>"],"action":"route","outbound":"out-<名>"}`，插在已有
  `sniff`/`hijack-dns` 之后、其它规则之前；已有规则相对顺序不变。
- 预检：前缀、起始/宽度、入站存在且支持用户身份、每个节点表示/协议/端口/
  认证/TLS/传输、客户端名与出站标签冲突、批量内重名、默认值范围。
- 事务：写 clients（复用 addbulk）→ 出站 → 基础 config 的 route → 生成完整
  config 并用 sing-box 反序列化校验 → 写批次与审计 → 提交。
- 幂等：同 actor + key 唯一；同键同内容返回原结果（`replayed`），同键异内容
  409；预览不写批次；键为空时用 `_<batchId>` 占位，仍受唯一索引保护。
- 核心应用：提交后 `RestartCore`；维护模式跳过并返回 `skipped`；失败返回
  `CORE_APPLY_FAILED` + `batchId` + `retry`，批次 `apply_failed`。重放前校验
  原批次客户端/出站/路由仍在，只重启核心，不重建。
- 脱敏：预览/结果/审计/批次视图不含节点密码；批次 `request` 只存
  `{index,line,protocol,address}` 等摘要，不存分享链接或自定义订阅 URI。

### 4.3 面板

- 新文件 `frontend/src/layouts/modals/ClientImportNodes.vue`：
  - 前缀/起始/宽度/默认协议/分组预览；
  - 入站多选；
  - 共享设置：启用、GiB 流量、到期、连接数、IP 限制、上下行 Mbps、描述/备注、
    延迟启动/自动重置/周期；
  - 节点 textarea（原始行或分享链接，一行一个，保留原始行号）；
  - 预览/提交、问题表、结果表、批次 ID、核心状态、失败后「重新应用到内核」；
  - 请求走 `api/importClientNodes`（Cookie），外部走 `apiv2/...`（Token），
    共用后端；
  - 修改表单会清空预览并换新的 `Idempotency-Key`；网络失败重试仍用原键。
- `frontend/src/utils/importNodes.ts` + 测试：行分类、空行行号、组名区间。
- `frontend/src/locales/i18n.test.ts`：编译所有语言消息，防止 `@`、`{` 等
  Vue I18n 语法把表单渲染打断。
- `frontend/src/views/Clients.vue`：工具菜单新增「按节点添加」。

## 5. 涉及文件

后端：

- `api/apiHandler.go`、`api/apiV2Handler.go`、`api/apiService.go`
- `api/apiv2_e2e_test.go`、`api/import_e2e_test.go`、`api/token_lifecycle_test.go`、
  `api/route_coverage_test.go`
- `core/validate.go`
- `database/model/import_batch.go`、`database/schema.go`、`database/backup_test.go`
- `service/import_nodes.go`、`service/import_nodes_test.go`、`service/user.go`

前端：

- `frontend/src/layouts/modals/ClientImportNodes.vue`
- `frontend/src/layouts/modals/Token.vue`
- `frontend/src/views/Settings.vue`、`frontend/src/views/Clients.vue`
- `frontend/src/utils/importNodes.ts`、`frontend/src/utils/importNodes.test.ts`
- `frontend/src/locales/{en,zhcn,zhtw}.ts`、`frontend/src/locales/i18n.test.ts`
- `frontend/dist` 同步到 `web/html`

## 6. 测试与实测结果

### 6.1 自动化测试

| 命令 | 结果 |
|---|---|
| `go test ./... -count=1` | ✅ 全部包通过，exit 0 |
| `go test -race ./service ./api -run 'TestImport|TestToken|TestAPIv2|TestTokenLifecycle' -count=1` | ✅ 通过，无 DATA RACE |
| `npm test`（frontend） | ✅ 10 个文件、**154/154** |
| `npm run build` | ✅ vue-tsc + vite build 通过 |
| `go build -tags "$(tags_for test)"` | ✅ 通过（仅 macOS 链接器重复库 warning） |

新增/更新的测试：

- `service/import_nodes_test.go`：编号/分组、独立凭据、特殊字符密码、预览零
  写入、客户端/出站冲突整批停止、非法端口行号、VLESS/VMess/HTTPS、幂等重放、
  同键异内容、并发抢同一范围、核心失败重放、无用户身份入站、路由顺序、共享
  身份字段拒绝、预览脱敏。
- `api/import_e2e_test.go`：无 Cookie Token 调用、预览、提交、重放、同键冲突、
  422 校验、401、批次查询/重放脱敏。
- `api/token_lifecycle_test.go`：过期/撤销/跨管理员/列表脱敏/Token 创建与所有
  权/Token 改密。
- `api/route_coverage_test.go`：解析 `apiHandler.go` 与 `apiV2Handler.go` 的
  `case` 字面量，断言除 `login`/`logout` 外 v1 动作都在 v2。
- 前端 `importNodes.test.ts`、`i18n.test.ts`。

### 6.2 隔离实例

环境：`SUI_DB_FOLDER=/tmp/sui-import-acceptance/db`，面板
`http://127.0.0.1:22095/panel/`，Token `zaSA…`（32 位，测试用），入站 1
`accept-in`（mixed，127.0.0.1:22097）。

| 场景 | 请求/操作 | 预期 | 实际 |
|---|---|---|---|
| 有效 Token | `GET /apiv2/status` 等 12 个 GET | 200 | ✅ 全部 200；`singbox-config` 返回原始 JSON |
| 缺失/错误 Token | `GET /apiv2/status` | 401 | ✅ 401 `invalid token` |
| 单个 SOCKS5 | `US`/1/3，1 节点，结构化 | US-001 / out-US-001 / US-1-1 | ✅ `imp_edb6…`，`core=applied` |
| 批量 10 HTTP | `US`/2/3，10 行 `192.0.2.x:8080` | US-002…US-011 / US-2-11 | ✅ 名称、标签、组、协议全符合 |
| 混合协议 | 裸行 SOCKS5、HTTPS 结构化、HTTPS URI、VLESS、VMess | 6 项协议正确 | ✅ `MX-1-6`；密码 `p@ss:word` 保留；VLESS/VMess TLS+WS 保留 |
| 预览 | 同请求 `dryRun=true` | 零写入 | ✅ 客户端/出站/批次均未增加；预览键未被占用 |
| 客户端重名（中间） | `US`/2，3 节点，US-002 已存在 | 422，新增 0 | ✅ 报 CLIENT_NAME_CONFLICT + OUTBOUND_TAG_CONFLICT；前后 clients=20 outbounds=21 rules=22 |
| 出站重名（中间） | 先建 `out-CF-002`，再导 `CF`/1 三节点 | 422，新增 0 | ✅ 报 OUTBOUND_TAG_CONFLICT；前后 outbounds=22 rules=22 |
| 非法端口（中间） | 第 2 行 70000 | 422，行号 2，新增 0 | ✅ `NODE_FORMAT` line=2；前后规则/计数不变 |
| 同键重放 | 同一键 + 同内容 | 原批次 + `replayed` | ✅ 同 `batchId`，客户端不增 |
| 同键异内容 | 同一键 + 不同 body | 409 | ✅ `IDEMPOTENCY_CONFLICT` |
| 不同键争抢 | 两并发请求同一 `RA-1-2` | 只成功一批 | ✅ 一个 200、一个 422；clients 只增 2 |
| 响应丢失 / 进程重启 | 发送完整请求后立即断开 socket；另重启面板后同键重放 | 重试拿到原批次 | ✅ 断开 socket 后 `imp_701b…`、`replayed=true`；重启后同键仍返回 `imp_edb6…`，批次仅 1 条 |
| 核心应用失败 | 维护模式停核心 → 占用 22097 → 退出维护 → 导入 | 500 CORE_APPLY_FAILED，数据已提交 | ✅ `imp_4f26…`，clients +1，批次 apply_failed |
| 恢复应用 | 释放端口 → `reapplyImportBatch` | applied，不重复创建 | ✅ `core=applied`，FAIL 客户端仍 1，出站 1 |
| 共享设置 | UI-050/051 | 各得 100 GiB、10/20 Mbps、maxConns 4、ipLimit 3 | ✅ 表/DB 一致 |
| 扩展字段 | EXT-001 | 规则字段、links、config 覆盖、身份独立 | ✅ 5 GiB、到期、autoReset 7 天、maxConns 6、ipLimit 4、3/9 Mbps；2 条本地+external+sub 链接 |
| 延迟启动 | DELAY-001 | delayStart=true、autoReset=false 时 expiry=0 | ✅ expiry=0，resetDays=1 |
| 实际登录/路由 | ST-001/ST-002 经 22097 到本地 23001/23002 | 各走 `out-ST-00x` | ✅ 返回 `upstream=A` / `upstream=B`；错误密码被拒 |
| 实际限速 | RL-001 downMbps=8，8 MiB 下载 | ≈8 Mbps | ✅ 8.267 s，1,014,675 B/s ≈ **8.12 Mbps** |
| 面板刷新按组 | 刷新后按 UI-50-51 筛选 | 只显示 2 条 | ✅ 2 行；批量编辑各 +10 GiB 后为 110.00 GB |
| 设置页 Token | 设置 → API 令牌 | 显示基础地址/脱敏列表 | ✅ `http://127.0.0.1:22095/panel/apiv2`，`****` |
| 面板重启 | Token 调 `restartApp` | Token 仍可用 | ✅ 5 秒内恢复，`/apiv2/status` 200 |
| 备份/恢复 | getdb → importdb 同库 | 数据/Token 保留 | ✅ 167,936 B，integrity ok；恢复后 29 客户端、12 批次，Token 200 |
| 升级兼容 | 旧二进制建库/签 Token → 新二进制启动 | 旧 Token 可用 | ✅ 旧 Token `VsWC…` 调 `/apiv2/status` 200；`/tokens` 脱敏；导入路由存在 |
| 设置/配置回存 | Token 调 save settings/config | 200 | ✅ 两者 200；配置保存触发核心重启后 Token 仍 200 |
| TLS | Token 新建/删除 | 成功 | ✅ `acceptance-tls`、`svc-tls` 新建后删除，表归零 |
| 客户端删除 | `del` / `delbulk` | 成功 | ✅ DELX-001/002 删除后计数 0 |
| 入站增改删 | 新建/改端口/删除 | 成功 | ✅ `throw-in` 22120→22121→删除 |
| 其他 Token POST | keypairs/linkConvert/subConvert/checkOutbound/getCertPing/restartSb/resetTraffic/closeSessions/update | 路由/错误可辨 | ✅ 见覆盖表；`update` 在 macOS 返回 manual 错误 |

截图：

- `/Users/yuzai/Tools/ip_api/.preview/settings-api-token-entry.png`：设置页入口。
- `/Users/yuzai/Tools/ip_api/.preview/settings-api-token-modal.png`：API 基础地址
  与脱敏 Token 列表。
- `/Users/yuzai/Tools/ip_api/.preview/import-nodes-preview-tall.png`：按节点添加
  的完整预览（分组/计数/结束编号/逐行名称/出站/协议/地址）。
- `/Users/yuzai/Tools/ip_api/.preview/import-nodes-result.png`：提交结果、批次 ID、
  核心状态、成功提示。
- `/Users/yuzai/Tools/ip_api/.preview/clients-filtered-group.png`：刷新后按
  `UI-50-51` 过滤，两个成员各 110.00 GB、4 连接、3 IP。

## 7. 本次问题与修复

1. **`ImportBatch` 的 `json.RawMessage` 在 SQLite 上扫描失败**
   （`unsupported Scan ... *jsontext.Value`）。改为 TEXT 字符串列，在服务内
   `json.Unmarshal([]byte(...))`；备份测试 fixture 同步。
2. **VMess 分享链接的端口是字符串**（`"port":"8443"`），
   `option.VMessOutboundOptions.server_port` 要求数字。解析后统一
   `out["server_port"] = port`，并让 `toInt` 接受数字字符串。
3. **前端 i18n 占位符含 `@`**：Vue I18n 把 `@` 当链接语法，消息编译失败，整个
   `a-form` 子树不渲染。把示例中的 `@` 改为 `{'@'}` 转义，并新增
   `src/locales/i18n.test.ts` 编译所有语言消息。
4. **面板调用了 `/apiv2/importClientNodes` 但没带 Token**，预览 401。面板改用
   Cookie 路由 `/api/importClientNodes`；外部仍用 `/apiv2`。
5. **错误提示在无预览/结果时不显示**：结果面板的 `v-if` 增加
   `issues.length`/`errorMessage`，401/422 可见。
6. **原生 `<textarea />` 自闭合**把后续表单当成文本吞掉；改为显式
   `></textarea>`。
7. **`Token.vue` 复制方法改名**后保留 `copyToClipboard` 兼容包装，避免旧调用
   失效。
8. **`core.ValidateConfig` 的初始化时机**：`core.NewCore()` 会注册协议并写日志，
   包级变量在测试 `TestMain` 之前初始化会空指针；改为 `sync.Once` 懒加载。
9. **SQLite 变量上限**：批量冲突查询按 400 分块；节点上限 1,000。
10. **批次视图泄漏风险**：`GET importBatch` 原计划返回原始请求，改为脱敏摘要
    （不含节点密码、分享链接、订阅 URI）；新增 e2e 断言。

## 8. 未验证项/边界

1. 服务/端点创建的真实副作用未验收：服务需要有效证书（DERP 返回
   `missing certificate`），端点需要真实 WireGuard/OpenVPN 参数；只验证了
   `save` 分发、校验错误和删除路径。
2. 单条客户端编辑与旧「批量添加」未重复走一遍；本次重点是新导入路径和批量
   编辑。
3. VLESS/VMess/HTTPS 只验证到配置生成、出站选项校验和核心启动，没有真实
   Reality/VMess 服务端跑流量；SOCKS5 已用本地受控节点验证登录、路由和限速。
4. 386 运行时未执行；本次只在当前 macOS 上构建/测试。`int64` 速率换算沿用
   `util/rateconv` 的既有 386 测试。
5. 不同键并发抢编号范围只在本机单进程、SQLite `_txlock=immediate` 下验证；
   多进程/多实例共用同一 SQLite 文件不在本次范围。
6. 预览的完整配置校验使用“当前客户端 + 待新增出站/路由”；新增客户端的入站
   `users` 只在正式提交的事务内做完整配置校验。
7. 移动端只检查了 390px 下弹窗宽度与页面无横向溢出（`body 390/390`），未做
   完整移动端截图/交互回归。
8. 前端 `frontend/` 不在 Understand-Anything 扫描范围；知识图谱只更新 Go 侧。

## 9. 知识图谱更新

- 基线：`beffb26`（1,440 节点 / 5,013 边 / 287 文件）；功能提交 `15cf4df`。
- 动作：`ARCHITECTURE_UPDATE`，13 个变化文件、6 个分析批次（batch-2/3/4/5/8/17）；
  `frontend` 子模块按仓库约定忽略；`api/apiHandler.go`、`database/backup_test.go`、
  `database/schema.go`、`service/user.go` 被增量规划判定为 cosmetic，旧节点保留。
- 结果：`knowledge-graph.json` **1,522 节点 / 5,446 边 / 297 文件 / 9 架构层 /
  11 步学习路径**；`service/import_nodes.go` 40 节点、`core/validate.go` 3、
  `database/model/import_batch.go` 2、`api/import_e2e_test.go` 4、3 个文档节点。
  新增节点由 finalize 按目录深度/连通性归入 `business-services`、
  `web-and-api`、`data-persistence` 等原有层；旧层名称/ID 与学习路径保留，新增
  第 8 步介绍 Token 与批量导入。
- 增量校验：symbol validation 通过（无符号丢失）；`meta.json` 指向 `15cf4df`。
  收尾的 AGENTS/本工作记录属于文档收尾提交，未再触发代码结构变化。
