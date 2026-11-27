# 用户列表「落地 IP」｜开发与实测工作记录

日期：2026-11-28
开发仓库：`/Users/yuzai/Tools/ip_api/s-ui`
前端子模块：`frontend/`
Git 备份点：主仓库 `e7f4060`（分支/标签 `backup/before-egress-ip-20261128`、
`backup-before-egress-ip-20261128`）；前端 `f8cf913`（同名分支/标签）。另用
`git bundle` 备份到 `backups/s-ui-before-egress-ip-20261128.bundle`、
`backups/s-ui-frontend-before-egress-ip-20261128.bundle`，`git bundle verify`
通过。

## 1. 开工前已读资料

- `AGENTS.md`：Agent 强制协作规范（知识图谱增量更新、工作记录、本地构建与联调、
  发布仓库规则）；「开发与变更工作记录」中的「按节点批量创建客户端」「出站管理
  改为可分页的 Ant Design 表格」「真实客户端 IP（PROXY protocol）与按用户 IP
  限制」「用户列表实时上/下行速率」「用户列表实时网速改为单标签」「入站管理改为
  表格列表并显示 PROXY protocol 状态」。
- `docs/import-nodes-worklog.md`：导入事务、`auth_user` 路由生成与重放校验、
  API Token 路由覆盖测试、测试与隔离实例做法。
- `docs/real-ip-worklog.md`（按需）：入站用户身份、可信 CIDR、会话来源 IP 与
  `userIps` API 的背景；本次不改这条链路。
- `docs/api-token.md`、`docs/api-coverage.md`：现有 API 文档结构、错误码与功能
  覆盖表的写法。
- `.ua/meta.json`（基线 `dc284a7`，1,596 节点 / 5,948 边 / 302 文件）与
  `.ua/knowledge-graph.json` 的 9 个架构层、12 步学习路径；重点节点：
  `core/outbound_check.go:CheckOutbound`、`service/config.go:CheckOutbound`、
  `api/apiService.go:GetCheckOutbound`、`service/import_nodes.go:addImportRouteRules`。
- 两个工作区开工前均为干净：主仓库 `e7f4060`，前端子模块 `f8cf913`
  （分支 `work/antd-tls-layout`）。

## 2. 可复用的实现入口（开工前核对）

- **出站测试链路**：`frontend/src/views/Outbounds.vue` → `api/checkOutbound` →
  `ConfigService.CheckOutbound` → `core.CheckOutbound`。`core.CheckOutbound`
  在 `c.running()` 下取一个 box 快照后释放锁，再执行 `urltest.URLTest`，
  超时 15 秒；这是新增“公网 IP 探测”应沿用的锁与超时策略。
- **运行中 core / outbound**：`core.Core.running()` 返回 `*Box`，
  `box.outbound.Outbound(tag)` 按 tag 取出 `adapter.Outbound`（含
  `DialContext(ctx, network, M.Socksaddr)`），可用于让请求真正经过选中出站。
- **当前生效配置**：`(&ConfigService{}).GetConfig("")` 把数据库中的基础
  `config` 与入站/出站/服务/端点合并为完整 sing-box JSON；路由规则在
  `route.rules`，默认出站在 `route.final`。
- **导入用户绑定**：`service/import_nodes.go:addImportRouteRules` 在开头
  `sniff`/`hijack-dns` 之后插入
  `{"auth_user":[name],"action":"route","outbound":tag}`，每个客户端一条；
  `verifyBatchData` 也用同样的规则读取方式。映射应读当前规则，而不是导入批次。
- **表格与当前页**：`frontend/src/views/Clients.vue` 的 `headers`、`bodyCell`、
  `displayedClients`、`tableData`、`pageNames`、`pageNamesKey`、`onTableChange`
  与 2 秒速度轮询；新列必须复用 `pageNames`（当前页 + 当前排序）且使用独立
  请求与定时器。
- **API 封装与鉴权**：v1 `api/apiHandler.go`、v2 `api/apiV2Handler.go` 的
  `case` 白名单、`checkLogin`/`checkToken`、`jsonObj`；前端
  `HttpUtils.get/post` 与 `api.get`（静默）两种调用方式。
- **多语言**：6 个 locale 文件（en/fa/ru/vi/zhcn/zhtw），key 无类型约束；
  `i18n.test.ts` 只校验消息可编译。
- **测试基建**：`core/proxy_e2e_test.go` 的 `NewCore().Start(config)` 与本地
  SOCKS/回显服务、`api/apiv2_e2e_test.go` 的 `apiv2TestServer`、
  `service/config_providers_test.go` 的临时 DB + `GetConfig` 组装方式。

## 3. 尚待验证的问题（开工时登记）

1. 生产可用的“固定 HTTPS IP 回显服务”在本机 IPv4 可达
   （`https://1.1.1.1/cdn-cgi/trace` 返回 `ip=`）；IPv6
   （`https://[2606:4700:4700::1111]/cdn-cgi/trace`）本机不可达，需要本地
   IPv6 回显服务验证解析与完整显示。
2. 出站为 selector/URLTest 或链式代理时，“最近一次检测结果”是否随成员切换
   变化；首版按出站 tag 缓存，需在记录中说明语义。
3. 同名出站重建、热编辑、核心重载三种失效路径是否能覆盖所有写入口（含
   `importClientNodes` 直接写库后重启核心）。
4. 手工规则的判定边界：目的地址/协议/入站/前置规则存在时保守返回“按规则
   分流”，需要构造用例确认不会把明确单出站误判为分流。
5. 前端翻页/排序/刷新与 2 秒速度轮询互不影响；首次检测中、失败、旧值刷新
   失败三种状态的展示与停止重试条件。
6. 现有出站延迟测速、`userIps` 明细、限额与新列共存；打包后的 `web/html`
   需要重新构建并由隔离实例加载。

## 4. 实现摘要

### 4.1 后端探测（`core/egress_ip.go`）

- 新增 `Core.ProbeEgressIP(tag)`：`running()` 取到运行中的 `Box` 后立即释放核心锁，
  再按 tag 取 `adapter.Outbound`，通过它的 `DialContext` 访问固定 HTTPS IP 回显
  服务。HTTP Transport 不设置代理、不提供直连兜底，所有 TCP（含 TLS 握手）都走
  选中出站。
- 回显端点：IPv4 `https://1.1.1.1/cdn-cgi/trace`，IPv6
  `https://[2606:4700:4700::1111]/cdn-cgi/trace`；两者都按 IP 访问，避免在面板
  主机上解析域名。每个端点 5 秒、整体 12 秒；响应体最多 4 KiB；只接受
  `ip=` 行或纯 IP，拒绝私网/回环/链路本地/组播/无法解析的内容。
- 指定出站失败直接返回错误，不回落到面板主机直连；selector/URLTest/链式出口
  统一解释为「该 tag 最近一次检测结果」。
- `SetEgressProbeConfigForTest` 只供测试替换端点与 CA，生产路径仍使用固定端点。

### 4.2 用户映射与缓存（`service/egress_ip.go`）

- `GetEgressIPs(users, refresh)` 读取当前生效配置（`GetConfig("")`）的
  `route.rules`/`route.final`，按规则顺序做三值判定（命中 / 可能命中 / 不命中）：
  - `auth_user`（数组或字符串）是唯一的客户端身份条件；`user` 是进程属主，
    不使用。
  - 无 `auth_user` 限制、无其它条件、动作为 `route` 的规则是确定性出口；第一条
    命中后停止。
  - 带目的地址、域名、协议、入站、端口、规则集等条件的规则只记为「可能出口」；
    与默认出站不同则 `split`。
  - 条件规则全部落空时用 `route.final`；无 `final` 时沿用 sing-box 的首个出站。
  - 逻辑规则 `and`/`or`/`invert`、`reject`、`auth_user` 单值/数组均已处理。
  - 不按用户名拼 `out-<用户名>`，不依赖导入批次。
- 缓存按出站 tag 存在进程内存，**默认只读**：不带 `refresh` 的请求不会发起
  任何探测，没有缓存时返回 `untested`；只有 `refresh=1`（管理员点击刷新）才
  启动探测。相同 tag 的并发探测合并（全局最多 4 个并发）；旧值仍在时刷新失败
  保留旧 IP 并标记 `stale`。出站新建/编辑/改名/删除立即失效对应 tag（含旧
  tag），核心重启时按 `Box` 实例变化整体失效；失效前的在途探测带 epoch，不回
  写新配置。缓存不按时间过期，成功结果一直显示到下次手动刷新、出站变更或核心
  重启。
- 探测是异步的：手动刷新后 API 立即返回 `probing`，前端做有限次状态轮询；
  失败后不再自动重试，只能再次手动刷新。

### 4.3 API（`api/`）

- 新增 GET 动作 `egressIps`，v1（Cookie）与 v2（Token）同时注册；参数
  `users=<JSON 数组>`、可选 `refresh=1`；沿用 `jsonObj` 响应。
- 响应：`running` + `users` 映射，每项 `status/outbound/ip/checkedAt/probing/
  refreshing/error`；状态语义与示例见 `docs/outbound-address-api.md`（现为出站地址接口）。
- 不新增表、不写客户端配置、不改备份 schema。

### 4.4 前端（`frontend/src/views/Clients.vue`）

- 「入站管理」后新增「落地 IP」列：主文本只显示 IPv4/IPv6，等宽字体、可选中；
  旁有复制与刷新按钮。tooltip 显示出站 tag、最近成功时间、刷新失败原因。
- 状态展示：`split` 显示「按规则分流」，`unbound` 显示「未绑定」，`stopped`
  显示「核心未运行」，无缓存 `untested` 显示「未检测」+ 刷新，手动刷新后的
  `probing` 显示「检测中…」，失败显示红色「检测失败」+ 重试，`stale` 保留旧
  IP 并标黄。
- 取数：`pageNames`（当前页 + 当前排序 + 筛选结果）变化时只请求后端缓存，
  打开页面/翻页/筛选/排序均不发起探测；未检测的行显示「未检测」。只有管理员
  点击刷新才发 `refresh=1`，之后仅 `probing`/`refreshing` 每 1.2 秒轮询状态，
  最多 30 次；失败、页面卸载、隐藏标签页都会停止。与 2 秒速度轮询完全独立。
- 新增 `utils/egressIp.ts` 及单测；6 种语言新增 13 个 `client.egress*` 文案。

## 5. 涉及文件

- 后端：`core/egress_ip.go`、`core/egress_ip_test.go`、`service/egress_ip.go`、
  `service/egress_ip_test.go`、`service/egress_ip_e2e_test.go`、
  `service/outbounds.go`（失效挂钩）、`api/apiService.go`、
  `api/apiHandler.go`、`api/apiV2Handler.go`、`api/egress_ip_e2e_test.go`。
- 前端：`frontend/src/views/Clients.vue`、`frontend/src/utils/egressIp.ts`、
  `frontend/src/utils/egressIp.test.ts`、
  `frontend/src/locales/{en,zhcn,zhtw,ru,vi,fa}.ts`。
- 文档：`docs/outbound-address-api.md`、`docs/api-coverage.md`、本文件、`AGENTS.md`。

## 6. 自动化测试

| 命令 | 结果 |
|---|---|
| `go test ./... -count=1` | ✅ 全部通过（api/cmd/core/database/logger/middleware/network/service/sub/util 等，无 FAIL） |
| `go test -race ./core ./service -run 'TestProbeEgressIP\|TestParseEgressIP\|TestResolveEgress\|TestEgressIPCache\|TestEgressIPService' -count=1` | ✅ 通过，无 DATA RACE |
| `go vet ./...` | ✅ 通过 |
| `go test ./api -run TestEgressIPsAPI` | ✅ 参数、鉴权、stopped/split、非法 JSON |
| `go test ./service -run TestEgressIPServiceIPv6EndToEnd` | ✅ 真实 sing-box 出站 + 本地 IPv6 TLS 回显：`2001:db8::99`；停 relay 后失败且不回落 |
| `cd frontend && npm test` | ✅ 11 文件 / **157** 测试（新增 3 个 egressIp 单测） |
| `cd frontend && npx vue-tsc --noEmit` | ✅ 0 error |
| `cd frontend && npm run build` | ✅ 产物 `assets/368cdbaa0c9fa88e.js` |
| `go build -tags "$(tags_for test)" -o sui main.go` | ✅ 通过；`web/html` 已同步新产物 |

新增 Go 测试覆盖：出站实际拨号、IPv4/IPv6 回显解析、指定出站失败不回落、响应
大小/非法内容拒绝、单用户/共享/逻辑/反选/reject/缺失 tag/默认 final 的映射、
相同出站合并、默认只读不自动探测、刷新失败保留旧值、出站改配置旧探测不回写、核心
重启整体失效。

## 7. 逐项实测（隔离实例）

> 本节记录功能开发阶段的实测；当时“打开页面会自动探测”，后续默认行为改为
> **只读缓存、仅手动刷新探测**（见第 12 节）。表中探测相关结论仍可通过点击
> 刷新按钮复现。

环境：`SUI_DB_FOLDER=/tmp/sui-egress-accept/db`，面板
`http://127.0.0.1:22095/panel/`，核心入站 `accept-in`（mixed，127.0.0.1:22100），
4 个首批客户端 + 8 个分页客户端。出站 `out-AL-001/out-BO-001/out-CA-001/
out-PG-00x` 指向本机可控 SOCKS5 中继 `127.0.0.1:22097`（Python，带连接计数），
回显服务为生产固定端点。中继地址与回显 IP 不同（127.0.0.1 vs 154.17.14.133）。

| # | 场景 | 操作/预期 | 实际 |
|---|---|---|---|
| 1 | 首次批量取数 | 4 用户一次请求；AL、BO、CA 各有明确出站，DA 无 | ✅ AL/BO/CA `probing`，DA `split`；中继只收到 **2** 次探测（AL 一次、BO/CA 共享一次） |
| 2 | 缓存与共享 | 再次请求不再探测 | ✅ 3 秒后全部 `ok`，IP `154.17.14.133`；重复请求中继计数不变 |
| 3 | 上游地址 ≠ 回显 IP | 出站服务器为 127.0.0.1，回显为公网 IP | ✅ 表格/API 显示 `154.17.14.133`，不是 127.0.0.1 |
| 4 | 刷新失败保留旧值 | 停中继，AL `refresh=1` | ✅ 立即 `ok+refreshing`；约 7 秒后 `stale`，仍保留 IP 与 `checkedAt`，`error` 为连接失败 |
| 5 | 指定出站不可用 | `out-AL-001` 改到关闭端口 22098，无旧值 | ✅ 先 `probing`，随后 `failed` 且无 IP；中继恢复、改回 22097 后恢复 `ok` |
| 6 | 同名出站改配置 | 共享 `out-CA-001` 改到 22098 | ✅ BO/CA 同时失效并 `failed`，旧 IP 未串到新配置；改回后只新增 1 次探测 |
| 7 | 更换绑定 | AL 规则改到 `out-CA-001`，再改回 | ✅ 立即返回 `out-CA-001`，切回 `out-AL-001` 后恢复；配置保存触发核心重启时缓存整体失效 |
| 8 | 分页 | 12 个客户端、10 条/页 | ✅ 第 1 页 10 行、第 2 页 PG-007/008；翻页请求的 `users` 恰为 `["PG-007","PG-008"]` |
| 9 | 排序 | 按名称降序 | ✅ 行顺序变为 PG-008…CA-001；`egressIps` 请求的 `users` 与新的 10 行完全一致 |
| 10 | 筛选 | 名称 `PG-00` | ✅ 8 行；请求只包含这 8 个用户名 |
| 11 | 手动刷新 | 点击 AL 行刷新 | ✅ 按钮 loading，返回后 IP 更新时间变化；中继只多 1 次连接 |
| 12 | 复制 | 点击复制图标 | ✅ 页面出现「成功 / 复制 IP」通知（截图 `.preview/egress-ip-copied.png`） |
| 13 | tooltip | 悬浮 AL 的 IP | ✅ 显示「出站: out-AL-001 / 最近成功检测: 2026/11/28 04:01:45」 |
| 14 | 旧值刷新失败 UI | 停中继后刷新 PG-001 | ✅ IP 变橙色并出现告警图标；tooltip 同时显示上次成功时间和连接失败原因 |
| 15 | 失败 UI（无旧值） | PG-008 出站指向关闭端口后刷新 | ✅ 单元格显示红色「检测失败」+ 重试按钮，无伪造 IP |
| 16 | IPv6 完整显示 | 通过本地反向代理只替换 `egressIps` 响应为 `2001:db8:1:2:3:4:5:6` | ✅ 完整显示/可复制；列宽 230px，地址宽 144.5px，不截断（截图 `.preview/egress-ip-ipv6.png`） |
| 17 | 窄屏 | 390×844，滚动表格到落地 IP | ✅ `body 390/390`，表格滚动区 1533/370；IPv6 完整可见（`.preview/egress-ip-ipv6-mobile.png`） |
| 18 | 实际拨号路径 | AL-001 通过 22100 SOCKS 登录请求 `example.com:80` | ✅ 中继日志 `example.com:80`；`api/sessions` 为 `outbound=out-AL-001, rule=auth_user=AL-001` |
| 19 | 既有来源 IP/限额 | 真实会话期间 `api/userIps?user=AL-001` + 打开 IP 弹窗 | ✅ `{count:1, limit:3, ips:[127.0.0.1/32]}`；弹窗显示 IP/会话数/首次出现；连接数 1/4 |
| 20 | 既有出站延迟测速 | 出站管理测试 out-AL-001 | ✅ 显示 `127 毫秒`，不受新接口影响 |
| 21 | 2 秒速度轮询 | 用户管理停留 | ✅ `↑ 0 kbps / ↓ 0 kbps` 持续更新；落地 IP 请求与速度请求分离，无相互触发 |
| 22 | 打包产物 | 重新构建并重启隔离实例 | ✅ 页面加载 `./assets/368cdbaa0c9fa88e.js`；1920px 无页面横向溢出 |

截图：

- `.preview/egress-ip-desktop-final.png`（1920，最终构建）
- `.preview/egress-ip-tooltip.png`（悬浮出站/时间）
- `.preview/egress-ip-copied.png`（复制通知）
- `.preview/egress-ip-stale.png`（旧值 + 刷新失败）
- `.preview/egress-ip-failed.png`（无旧值失败）
- `.preview/egress-ip-ipv6.png`、`.preview/egress-ip-ipv6-mobile.png`（IPv6 与 390px）
- `.preview/egress-ip-page2.png`、`.preview/egress-ip-sorted.png`、`.preview/egress-ip-filtered.png`
- `.preview/egress-ip-userips-modal.png`（既有来源 IP 弹窗）

## 8. 本次问题与修复

1. **映射漏掉默认出口**：第一版在条件规则命中后只返回该规则出口，没有把
   `route.final` 作为另一可能出口，逻辑 AND 用例被误判为确定出口。已改为
   「未命中确定性规则时追加默认出口」，并补逻辑规则用例。
2. **测试缓存与全局缓存混淆**：`invalidateEgressIPCache` 只操作全局实例，单测
   的本地 cache 需要方法级 `invalidate`；已拆分方法并补按 tag 失效。
3. **探测异步导致测试取到 0 次调用**：单测需等探测 goroutine 真正开始后再断言；
   已用 started channel 同步，并给可能阻塞的测试加 cleanup。
4. **前端请求失败会永久转圈**：API 返回 `success:false` 或网络错误时，原本没有
   结果的行会一直「检测中」。已改为把从未有结果的行标记 `failed`；已有旧值则
   保留，旧请求序号不能覆盖新状态。
5. **探测错误只显示最后一个端点**：IPv4 失败后只看到 IPv6 端点的错误；已改为
   `errors.Join` 同时报告两个端点，便于现场判断。
6. **IPv6 UI 验证**：本机无全局 IPv6 路由，生产 IPv6 回显不可达；后端用本地
   `[::1]` TLS 回显做真实出站测试，UI 用只替换 `egressIps` 响应的本地反向代理
   验证完整显示/复制。该验证不改变交付代码。

## 9. 未验证项与边界

- 生产 IPv6 回显端点未在本机端到端跑通（无全局 IPv6 路由）；IPv6 拨号、解析、
  缓存与失败路径由 `core`/`service` 的本地 IPv6 TLS 回显测试覆盖。
- selector/URLTest/链式代理只验证「按 tag 探测」的语义与缓存，未用真实成员切换
  压测；UI 不展示成员。
- 复杂规则集（`rule_set`、`geoip`、`process_*`）只按「存在条件 → 可能出口」
  保守处理；可能把实际单一出口显示为「按规则分流」，本次未穷举。
- 未在 386 目标运行；未测真实 Clash/mihomo/Hiddify 订阅刷新（与本功能无关）。
- 缓存不按时间过期，未做长时间驻留回归；只读/手动刷新行为由单测和隔离实例
  覆盖。
- 未重启用户正在使用的 2095 面板；验收使用隔离实例 22095，交付为本地开发仓库。
- 知识图谱：见下节。

## 10. 知识图谱更新

- 基线：`.ua/meta.json` 的 `dc284a7`（图谱 1,596 节点 / 5,948 边 / 302 文件）。
- `prepare-incremental.mjs` 判定 `ARCHITECTURE_UPDATE`：本轮需分析 11 个文件
  （3 个 Go 实现、4 个 Go 测试、AGENTS.md、3 个文档），删除 0、cosmetic 3、
  ignored 1（`frontend/` 不在扫描范围）；`compute-batches.mjs` 生成 6 个批次。
- 用 6 个 file-analyzer 子代理并行完成：`core/egress_ip*`（批次 1）、
  `service/egress_ip*`（批次 2）、`api/egress_ip_e2e_test.go`（批次 3）、
  `api/apiService.go`（批次 8，旧 50 个符号全部保留 + 新增 `GetEgressIPs`）、
  `AGENTS.md`（批次 14）、3 个文档（批次 17）。
- `merge-batch-graphs.py` + 符号校验通过：10 个已有/新增文件函数节点均有增无减，
  无缺失符号；合并后 1,650 节点 / 6,268 边 / 310 文件。
- 架构层沿用 9 层，新增文件按目录归类：`core/egress_ip*` → sing-box 核心层，
  `service/egress_ip*` → 业务服务层，`api/egress_ip_e2e_test.go` → Web/API 层，
  两个新文档 → 文档层；`layers.json` 校验无遗漏、无重复归属。
- 学习路径新增第 9 步「客户端落地 IP 与出站探测」（core → service → api →
  API 文档），原 9–12 步顺延为 10–13；共 13 步。
- 架构阶段 `finalize-incremental.mjs` 保存成功，`.ua/meta.json` 指向功能提交
  `acbdd18`；文档收尾后再做一次 `PARTIAL_UPDATE`（只重分析 `AGENTS.md` 与本
  记录），最终指向文档提交，`analyzedFiles` 310。`code-lessons.json` 本轮之前
  不存在，因此未生成该侧车文件（仪表盘按 overview 摘要降级显示）。
- 知识图谱数据目录在 `.gitignore` 中，不随代码提交入库。

## 11. 列顺序、日期合并与列宽微调（同日）

- 列顺序改为：启用 → 名称 → 组 → 入站 → 落地 IP → 创建时间/到期时间 →
  在线 → 连接数 → IP 数 → 流量 → 实时网速 → 操作 → 最后在线 → 描述（描述
  恢复显示并固定在最后一列）。
- 「创建时间」「到期时间」合并为一列，上下两个 `a-tag` 展示，格式固定为
  `YYYY-MM-DD HH:mm:ss`（`formatDateTime`，本地时区）；到期为 0 显示「无限」，
  已过期标红；标签 tooltip 分别标明「创建时间」「到期时间」。
- 实时网速保持下行在前：`↓ 下行 / ↑ 上行`。
- 14 列宽度（px）：启用 64、名称 112、组 76、入站 92、落地 IP 200、日期 160、
  在线 66、连接数 74、IP 数 72、流量 135、实时网速 210、操作 146、最后在线 125、
  描述 140。
- 验证：`npm test` 12 文件 / **159** 测试（新增 `formatDate` 2 个用例）、
  `vue-tsc --noEmit`、`eslint`、`npm run build` 通过；1920px 表宽 1680px、
  无横向滚动、14 个表头均单行；1440px/390px 仅表格内部横向滚动，
  `body.scrollWidth == body.clientWidth`；行高约 68–70px（两条日期标签）。
  截图：`.preview/egress-ip-dates-1440.png`、`.preview/egress-ip-desc-last.png`。
- 本次只改前端（`Clients.vue`、`formatDate.ts`、6 语言文案），无 Go/API/
  数据库变更；`frontend/` 不在知识图谱扫描范围。

## 12. 默认改为只读缓存（同日调整）

- 页面打开、翻页、筛选、排序以及普通状态轮询都不再发起探测；后端
  `GetEgressIPs` 默认只读缓存，没有缓存时返回 `untested`（前端「未检测」）。
- 只有管理员点击「重新检测」发 `refresh=1` 时才启动探测；相同出站仍合并，
  最多 4 并发；探测中前端有限次轮询，失败后不自动重试。
- 缓存不再按 10 分钟/30 秒 TTL 自动刷新；成功 IP 一直显示，直到手动刷新、
  出站变更或核心重启。`stale` 仍表示“有旧值但最近一次手动刷新失败”。
- 单测调整为 `TestEgressIPCacheViewDoesNotProbeByDefault`、
  `TestEgressIPCacheStaysCachedWithoutRefresh`、
  `TestEgressIPCacheDoesNotRetryWithoutRefresh` 等；`TestEgressIPServiceIPv6EndToEnd`
  先断言默认 `untested`，再手动 `refresh` 触发探测。前端 `npm test` 仍 12 文件 /
  159 测试通过。

## 13. 最终改为「出站地址」（同日）

- 用户确认：表格要显示的 `1.1.1.1` 是出站 `socks-s28` 配置的服务器地址，不是
  回显服务实测的公网出口；要求页面打开立即显示，不做探测。
- 删除探测链路：`core/egress_ip.go` 与其测试、`service` 的缓存/TTL/并发控制、
  `service/egress_ip_e2e_test.go`、`api/egress_ip_e2e_test.go`；出站保存不再失效
  探测缓存。
- 新接口：`GET {api,apiv2}/outboundAddresses?users=<JSON数组>`。按
  `route.rules`/`route.final` 找到出站 tag，读当前配置的 `server`/
  `server_port`/`type`，不拨号、不缓存、不建表。
- 状态：`ok`、`direct`、`multi`（selector/urltest）、`none`（无 server 字段）、
  `split`、`unbound`；链式代理显示该出站自己的 server，域名原样显示。
- 前端列改名「出站地址」：直接显示、tooltip 显示出站与 `地址:端口`、可复制；
  无刷新按钮、无轮询、无「检测中/未检测」；直连/多出口/无地址/分流/未绑定分别
  展示。
- 验证：`go test ./...` 全通过；前端 159/159、`vue-tsc`、构建通过，产物
  `assets/96be4f917e410da5.js`；隔离实例 `socks-s28` 配置 `1.1.1.1:80`，API 与
  页面均立即返回/显示 `1.1.1.1`，无探测。
- 边界：显示的是配置地址，不保证等于真实公网出口；域名不解析；selector/
  urltest 不展开当前成员。

## 14. 搜索、多选批量删除与联动清理（同日）

- 工具栏全部靠左；新增按名称/组/描述/备注搜索的输入框，保留原筛选器；
  搜索后回到第 1 页。
- 表格新增多选列和「批量删除(n)」按钮，确认弹窗显示数量；单条删除确认提示
  联动删除出站与路由。页脚显示总数/筛选数/已选数，分页增加快速跳页和
  每页条数。
- 后端 `ClientService.DeleteClients`：
  - 删除顶层、`auth_user` 全部处于本次删除集合、且不是 logical/invert 的路由
    （不要求出站名为 `out-<user>`）；
  - 删除被这些路由引用、且不再被其它规则或 `route.final` 引用的出站；
    `direct`/`block` 内置出站保留；
  - 删除客户端对应的个人订阅成员；组内无成员时删除组订阅，仍有其它成员时
    保留组；
  - 共享路由、条件规则、final 出站保留；
  - 单条 `del` 与 `delbulk` 共用；路由/出站有变化才异步重启核心；
  - 响应带 `outbounds/config`，前端立即刷新。
- 验证：`go test ./...` 通过（新增 4 个删除清理测试）；前端 159/159、
  `vue-tsc`、构建通过；隔离实例两个导入用户批量删除后出站只剩 `direct`、
  路由只剩 `sniff`/`hijack-dns`，核心仍在 22100 监听；15 条时分页
  `1-10 / 15`、页脚 `共 15 条`，1920px 表宽 1680px。
- 边界：订阅组仅在没有成员时删除，共享组保留；共享/条件/final 出站不删除。
- **修复（同日）**：首版只识别 `out-<name>`；现场手动命名为 `socks-s28` 的
  出站和 `WdkM3DF6` 路由在删客户端后仍残留。已改为任意 tag，并保护
  `direct`/`block`。用面板 API 清除现场残留后，出站只剩 `direct`、路由只剩
  `sniff`/`hijack-dns`。新增任意 tag 与 direct 保护测试；修复提交 `784b249`。
