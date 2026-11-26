# S-UI 真实客户端 IP 与按用户 IP 限制 — 工作记录

> 日期：2026-11-25
> 目标版本：`v1.7.0`（已发布）
> 关联计划：[s-ui-real-ip-plan.md](../../s-ui-real-ip-plan.md)
> 关联评审：[s-ui-real-ip-claude-review.md](../../s-ui-real-ip-claude-review.md)
> 发布仓库：<https://github.com/beacon8/s-ui-custom/releases/tag/v1.7.0>

本文记录这次功能的完整实现、配置方法、验收过程、发布结果和已知边界，供下次升级、排障或复用。

---

## 1. 目标与结论

链路：用户 → 国内入口 → WS/TLS 中转 → 香港出口 → s-ui 落地。

要解决两件事：

1. 会话页显示**真实客户端 IP**，而不是香港出口 IP。
2. 按用户统计不同 IP 数，并可设置 `ipLimit` 限制。

最终方案：

- 中转在每条 TCP 连接开头写 **PROXY protocol**（现场为 v1 TCP）。
- sing-box 监听器在协议握手前消费 PROXY 头，把真实 IP 放入独立的 `ClientAddr`。
- `metadata.Source` 始终保留对端地址，路由、`source_ip_cidr`、访问控制行为不变。
- s-ui `SessionTracker` 按 `ClientAddr` 统计 IP；超限只拒绝新 IP，已在线 IP 不受影响。

结论：功能已发布 `v1.7.0`，amd64 Linux 实测通过；**现场上线仍需要运维按第 4 节配置中转和 s-ui 入站**。

---

## 2. 最终链路与数据流

```
客户端
  │  普通代理连接（VMess/VLESS/Trojan/SS/SOCKS/HTTP…）
  ▼
国内入口（中转）
  │  按规则转发，每条 TCP 连接开头写 "PROXY TCP4 <客户端IP> <目标IP> <源端口> <目标端口>\r\n"
  ▼
香港出口
  │  TCP
  ▼
s-ui / sing-box 入站（监听端口，例如 50757）
  │
  ├─ 对端在“可信中转 CIDR”内？
  │    ├─ 是：读 PROXY 头 → ClientAddr = 头里的客户端 IP
  │    │      头缺失/格式错/超时 → 断开，不回落到香港 IP
  │    └─ 否：直连 → 不解析头，ClientAddr = 对端 IP
  │
  ├─ 入站协议握手（WS/HTTP/REALITY/VMess…）
  │
  └─ Router → SessionTracker
          ├─ Source：对端地址（香港出口），路由继续用它
          └─ Client：真实客户端 IP
                ├─ 按 IPv4 /32、IPv6 /64 归并
                ├─ 每用户 IP 引用计数
                └─ ipLimit > 0 时拒绝新 IP
```

关键原则：

- **PROXY 头必须由可信对端发送**，不能无条件采信客户端自带的头。
- **不修改 `Source`**，否则 `source_ip_cidr` 等路由规则会失效。
- **PROXY 头按 TCP 连接生效**，不是按用户、不是按隧道。

---

## 3. 代码改动

### 3.1 sing-box 补丁（fork）

| 项 | 值 |
| --- | --- |
| 上游 | `SagerNet/sing-box v1.14.4` |
| fork | `beacon8/sing-box` |
| 分支 | `sui-v1.14.4` |
| 标签 | `v1.14.4-sui.2` |
| 引用方式 | `replace github.com/sagernet/sing-box => github.com/beacon8/sing-box v1.14.4-sui.2` |

补丁约 300 行、9 个文件：

| 文件 | 作用 |
| --- | --- |
| `common/listener/preprocess.go`（新增） | 预处理接口、按监听端口策略、传输层待处理连接 |
| `common/listener/listener_tcp.go` | Accept 后在连接自己的 goroutine 里解析；`ListenTCP` 给 transport 返回包装后的 listener |
| `adapter/inbound.go` | `InboundContext` 新增 `ClientAddr` |
| `transport/v2raywebsocket/server.go` | `ConnContext` 中等待解析并写入 ctx |
| `transport/v2rayhttp/server.go` | 同上 |
| `transport/v2rayhttpupgrade/server.go` | 同上 |
| `transport/v2raygrpclite/server.go` | 同上 |
| `transport/v2raygrpc/server.go` | 用 gRPC `StatsHandler.TagConn` 取地址，`Tun` 使用 stream ctx |
| `transport/v2rayhttp/conn.go` | `DupContext` 保留 `ClientAddr` |

设计要点：

- listener 只定义接口，策略由 s-ui 通过 context 注入，sing-box 不感知 s-ui 配置。
- 按**监听端口**查策略，所以同一进程里不同入站可以不同配置。
- 接受循环不读取头部；每条连接在独立 goroutine 里读，默认 5 秒超时。
- 头部之后多读的字节（TLS ClientHello / WS 请求）用 `bufio.NewCachedConn` 接回。
- 直连用户不包装、不读头，行为与未开启时一致。
- 旧版 PROXY protocol 字段仍保持废弃状态，不启用 sing-box 自带字段。

升级 sing-box 时按上表函数 rebase，重点检查：

1. `listener.Options`/`Listener` 是否仍由 inbound 构造函数创建。
2. transport 的 `ConnContext` 是否仍存在。
3. `v2rayhttp.DupContext` 是否仍会丢弃 context。
4. gRPC 的 `stats.Handler` 接口是否变化。

### 3.2 s-ui 解析器

`util/proxyproto/proxyproto.go`：

- 支持 v1：`TCP4`、`TCP6`、`UNKNOWN`（UNKNOWN 无地址，s-ui 侧拒绝）。
- 支持 v2：`LOCAL`、`TCP4`、`TCP6`，带 TLV。
- 拒绝 UDP、UNIX、坏签名、超长、短负载。
- v1 行长上限 107 字节；v2 payload 上限 4096 字节。
- 返回时保留 `bufio.Reader` 中多读的字节，由调用方接回连接。

17 个单测覆盖拆包、合包、IPv4/IPv6、TLV、LOCAL、坏包等。

### 3.3 s-ui 预处理器

`core/proxyprotocol.go`：

- 以**监听端口**为 key 查入站策略：`enabled`、`trustedCidrs`、`headerTimeout`。
- 对端不在可信 CIDR：原样返回，不读头。
- 对端可信：
  - 设置读超时 → 读 PROXY 头 → 清除超时。
  - `LOCAL` 命令使用对端地址（中转健康检查，不归属用户）。
  - 头部带真实 IP → 构造 `ClientAddr`。
  - 多余字节写入 `CachedConn`。
  - 错误直接返回，由 listener 关闭连接，不回落。
- `ClientAddr` 同时写入 context 和 `metadata`。
- 策略更新通过 `proxyPolicyStore` 指针生效，无需重启监听器。

### 3.4 s-ui 计数与限制

`core/tracker_session.go`：

- `Session` 新增 `Client`、`ipKey`。
- `SessionTracker` 新增：
  - `userIPLimits map[string]int`
  - `userIPs map[string]map[netip.Prefix]*userIPState`
  - `ipv4PrefixLen`（默认 32）
  - `ipv6PrefixLen`（默认 64）
  - `ipLimitDryRun`
- `newSession`：
  1. `Client = ClientAddr`，无效则用 `Source`。
  2. 按前缀归并 `ipKey`。
  3. 连接数上限检查（原有）。
  4. IP 上限：`ipLimit > 0`、该 IP 不在列表、已有 IP 数 `>= ipLimit` → 拒绝。
  5. 成功才增加 IP 引用计数。
- `leave` / `takeSessions` 精确释放；`CloseByUser`、`CloseByInbound` 等踢人路径也会释放。
- `ipLimitDryRun` 只打日志，不拒绝。
- 公开方法：`SetUserIPLimits`、`SetIPPrefixes`、`SetIPLimitDryRun`、`ActiveIPCounts`、`UserIPs`。

### 3.5 配置与 API

数据库/模型：

- `clients.ip_limit`（GORM AutoMigrate 自动加列）。
- `inbounds.proxy_protocol`（面板侧 JSON，不进入 sing-box 配置）。

服务层：

- `InboundService.GetProxyProtocolPolicies()`：按端口生成策略；非法行记日志并跳过。
- 入站保存时校验 CIDR、超时、非空列表；缺字段保留旧值。
- `ClientService.GetIPLimitMap()`、`Save` 负数归零、缺字段保留。
- 核心启动/重启/保存后同步策略与限额。

API：

- `GET api/sessions`：每条会话新增 `client`（真实 IP），`source` 不变。
- 新增 `GET api/userIps?user=`：`{count, limit, ips:[{ip, sessions, since}]}`。
- `GET api/load` 新增 `ipCounts`。
- 客户端列表接口返回 `ipLimit`。

### 3.6 前端

- 会话来源列：主显示 `client`；与 `source` 不同时第二行显示“经 source”。
- 客户端列表：新增“IP 数 / 限额”列，可点击查看 IP 明细。
- 客户端编辑：新增“IP 限制”，0 = 不限制。
- 入站编辑：新增 PROXY protocol 开关、可信 CIDR 列表、头部超时。
- 设置页：IPv4 前缀、IPv6 前缀、IP 限制演练开关。
- 六种语言文案。

---

## 4. 配置手册（重点）

### 4.1 中转侧（截图对应的配置）

示意配置：

| 项 | 值 | 说明 |
| --- | --- | --- |
| 入口 | 国内专线 | 客户端连接的第一跳 |
| 监听端口 | 例如 50757 | 客户端实际连接的入口端口 |
| 出口 | 香港 CUG/CMI | 到 s-ui 的出口 |
| 协议 | WS 隧道 | 需确认是透明 TCP，还是终止 WS 后重建 |
| 目标地址 | `s-ui服务器IP:s-ui入站端口` | **不是面板 2095** |
| 接受 Proxy Protocol | 关闭 | 入口是直连用户，正确 |
| 发送 Proxy Protocol | v1 (TCP) | 到 s-ui 的方向，正确 |
| 负载均衡 | 随机 | 每个出口 IP 都要加入 s-ui 可信 CIDR |
| 规则限速 / IP 限制 / 连接数 | 0 | 中转侧限制，按需 |

必须确认：

1. **目标端口是 s-ui 入站端口**。如果 s-ui 入站是 443/8443，目标就写那个；50757 只是这条中转的入口监听端口时，不要直接抄到目标。
2. 中转和 s-ui 是否同一台机器。如果是，目标写 `IP:50757` 可能回环到中转自己；应写 `127.0.0.1:<s-ui 入站端口>`，并让入口用另一个端口。
3. `WS 隧道` 如果是**纯 TCP 转发**：s-ui 的 WS path、TLS/REALITY 都由客户端和 s-ui 端到端协商，中转不关心。
4. `WS 隧道` 如果会**终止客户端 WS 再向目标重建 WS**：s-ui 入站的 WS path、Host/SNI 必须和中转重建时完全一致，否则 404/握手失败。
5. 中转必须**每条 TCP 连接**写一次 PROXY 头。如果多条用户连接复用一条隧道，只写一次头，s-ui 只能看到首条连接的 IP 或全部算成香港 IP。
6. 客户端到入口的 TLS：
   - 截图显示“TLS 入站：未开启”。如果客户端链接是 `wss://入口`，会握手失败。
   - 如果中转是透明 TCP、TLS 由 s-ui 终止，则中转保持关闭是正确的，客户端仍用 `wss://s-ui域名`。
   - 如果是明文 `ws://入口`，公网链路无 TLS，建议尽快改为 TLS。

### 4.2 s-ui 入站侧（本次现场问题所在）

在 s-ui 入站编辑中：

- 打开“接受可信中转的 PROXY 头”。
- **可信中转 CIDR 填中转出口在 s-ui 上看到的源 IP**，例如：

```text
1.2.3.4/32
```

多个出口：

```text
1.2.3.4/32
1.2.3.5/32
```

IPv6 出口：

```text
2001:db8:abcd::/64
```

- 头部超时保持 5 秒。

**关键坑：`0.0.0.0/24` 不是“所有 IPv4”。**

它只匹配 `0.0.0.0` – `0.0.0.255`。填这个值时：

- 香港出口不在可信列表 → s-ui 不解析 PROXY 头。
- 中转发来的 `PROXY TCP4 ...` 文本被当成 WS 请求第一行。
- 结果：WS/HTTP 握手失败，连接直接断。

正确做法是填实际出口 IP/32，或服务商给出的出口网段。`0.0.0.0/0` 才是所有 IPv4，但**不要长期使用**：任何能直连入站端口的人都能伪造 PROXY 头。

### 4.3 如何找到“可信 CIDR”里的 IP

在 s-ui 服务器上，中转连接存在时：

```bash
ss -tnp | grep ':50757'
```

或抓包：

```bash
sudo tcpdump -n -i any 'tcp port 50757'
```

看 `Remote Address`/抓包源地址，那就是要填的香港出口 IP。也可以直接向中转服务商索要出口 IP 列表。

注意：填的是 **s-ui 看到的对端地址**，不是：

- 用户自己的公网 IP
- 中转入口的 IP
- s-ui 面板地址
- PROXY 头里的客户端 IP

### 4.4 客户端限额

- 客户端编辑 → “IP 限制”，0 = 不限制。
- 建议先设 `ipv4Prefix=32`、`ipv6Prefix=64`。
- 先开设置页的“IP 限制演练”，只记录“本应拒绝”，确认误伤率。
- 再按用户设置 `ipLimit`；已在线 IP 不会被踢，只有新 IP 被拒。
- 限额建议至少 2，避免 WiFi/4G 切换时旧连接未释放导致短时不可用。

### 4.5 防火墙

- s-ui 入站端口只放行香港出口 IP。
- 面板 2095 不要和入站端口混用。
- 如果中转有多个出口，逐一放行。

### 4.6 常见现象与排查

| 现象 | 最可能原因 | 检查 |
| --- | --- | --- |
| 中转显示已连接，客户端连不上 | s-ui 未开 PROXY 接受，或可信 CIDR 不含出口 IP | 4.2、4.3 |
| 连接立即断，日志有 WS/handshake 错误 | `PROXY ...` 被当成协议首行 | 可信 CIDR 是否 `/32` 实际出口 |
| 会话 `client` 显示香港 IP | 没解析头，或对端不在可信列表 | `source` 是否香港；日志 peer |
| 两个用户显示同一个 IP | 中转复用一个隧道，只写一次头 | 4.1 第 5 条 |
| 客户端用 `wss://` 连不上 | 入口 TLS 未开启 | 4.1 第 6 条 |
| 目标端口连不上 | 写成了面板 2095 或中转自身监听口 | 4.1 第 1、2 条 |
| `api/userIps` 为空但会话正常 | 用户没有认证名，或 IP 未归并 | 会话 `user`、`client` |
| 限额后旧网络还在线 | 旧连接未释放，属预期 | 等 TCP 超时；限额至少 2 |

---

## 5. 验证记录

### 5.1 单测与 race

```bash
go test ./util/proxyproto/ -count=1
go test ./core/ ./service/ ./api/ ./database/ -count=1
go test -race ./core/ -run 'TestSessionTracker|TestProxyProtocol|TestIPKey' -count=1
```

结果：全部通过，无 `DATA RACE`。

### 5.2 真实监听端到端

测试文件：`core/proxy_e2e_test.go`

- SOCKS + PROXY v1：`client=192.0.2.10:50000`，`source=127.0.0.1:…`。
- 同一用户第二个 IP：被 `ipLimit=1` 拒绝；第一个 IP 不受影响。
- VLESS 经 WS / HTTP / HTTPUpgrade / gRPC 四种 transport：均携带真实 IP。
- 完整 gRPC：`go test -tags with_grpc ./core/ ...`。
- REALITY：`go test -tags with_utls ./core/ -run TestProxyProtocolEndToEndReality`。
- 伪造 `X-Forwarded-For`：`source` 变为伪造地址，`client` 仍是 PROXY 地址。
- PROXY 头逐字节发送：正常。
- PROXY v2 IPv6：`2001:db8:1:2::1` 归并为 `/64`。
- 200 条慢连接：新连接 1.5 秒内完成。
- `source_ip_cidr`：命中中转 IP 会拒绝，命中真实客户端 IP 不会拒绝。

### 5.3 面板实测（临时实例）

- 保存入站后 `api/load` 回显 `proxyProtocol`。
- 客户端 `ipLimit` 保存/回显。
- 首连接后：
  - `api/sessions` 的 `client/source` 正确。
  - `api/userIps` 返回 `{count:1, limit:1, ips:[192.0.2.10/32]}`。
  - `api/load` 的 `ipCounts.alice=1`。
- 第二个 IP 被切断，计数仍为 1；断开后归零。
- 设置页三项持久化。

### 5.4 386

```bash
. ./build-tags.sh
TAGS=$(tags_for test)
CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -tags "$TAGS" -o /tmp/sui-linux-386 main.go
```

结果：静态 ELF 32-bit 构建通过；未在 386 运行时执行测试。

### 5.5 v1.7.0 Linux 安装验收

- x86_64 Alpine Lima 虚拟机，从 v1.6.10 升级。
- 安装器校验、`Migration done!`、服务启动成功。
- `sui -v`：`S-UI Panel 1.7.0` / `Sing-Box v1.14.4`。
- `/app/` 返回 307 到 `/app/login`。
- 升级后设置自动补齐 `ipv4Prefix=32`、`ipv6Prefix=64`、`ipLimitDryRun=false`。
- 创建带 `proxyProtocol` 的入站和 `ipLimit=2` 的客户端，回读正确。
- 真实连接：两个 IP 正常计数，第三个 IP 被拒绝；会话 `client` 为客户端 IP、`source` 为 127.0.0.1；断开后计数归零。

---

## 6. 发布

- 版本：`v1.7.0`
- 发布提交：`77257b7`
- 标签：`v1.7.0`
- Release：<https://github.com/beacon8/s-ui-custom/releases/tag/v1.7.0>
- 工作流：`45576896933`，成功
- 附件：`install.sh`、`s-ui-linux-386.tar.gz`、`s-ui-linux-amd64.tar.gz`、`SHA256SUMS`
- SHA256：
  - `install.sh`：`6ed646c314f20b8587f81fb3de5a5d179ecc999bc4f1b61cddd2fe3c9401e5e3`
  - 386：`2bea1c7e8af1a9b26cb7880939f7cbc2609ff856b5a492ed45327570aa57cc05`
  - amd64：`32da2fb89d1cc7d98a1d5403a86b6638fba8a5d649ecbf564301065d015df311`

开发仓库提交：

- 后端功能：`6641e05`
- 前端：`c438cc2`
- 补充验收测试：`dd5c4d6`
- 发布仓库功能：`0ef0164`
- 发布仓库测试：`961fc99`

---

## 7. 已知边界

- 阶段 0 现场验证和阶段 4 灰度由运维执行，代码已就绪。
- REALITY 握手与首包已覆盖；Vision flow 吞吐/升级回归未压测。
- 中转来的 SS 原生 UDP 拿不到真实 IP，不计数、不限制；QUIC 协议仅直连可计数。
- 运行中修改 IPv4/IPv6 前缀只影响新会话，重启后统一。
- 386 只验证交叉编译，未做运行时测试。
- 上游 `X-Forwarded-For` 可伪造 `metadata.Source` 的既有问题仍在；本功能不依赖 `Source`。
- sing-box 升级需要 rebase 补丁；fork 必须保持公开可拉取。

---

## 8. 下次检查清单

- [ ] s-ui 入站端口与中转目标端口一致，且不是面板 2095。
- [ ] s-ui 入站已开启 PROXY protocol。
- [ ] 可信 CIDR 是实际香港出口 IP/32，不是 `0.0.0.0/24`，不是 `0.0.0.0/0`（除非临时测试）。
- [ ] 多个出口 IP/网段都列入。
- [ ] 中转每条 TCP 连接都写 PROXY 头；复用隧道方案已实测不串 IP。
- [ ] 客户端 `wss://` 与入口 TLS 状态匹配。
- [ ] 入站防火墙只允许香港出口访问。
- [ ] 会话页 `client` 为真实 IP、`source` 为香港出口。
- [ ] 两个不同网络客户端同时连接，IP 不串。
- [ ] 先 `ipLimitDryRun`，再设 `ipLimit`。
- [ ] 断开连接后 `api/userIps` 计数归零。
- [ ] 升级 sing-box 前按第 3.1 节 rebase 补丁。

---

## 9. 文件索引

| 文件 | 说明 |
| --- | --- |
| `util/proxyproto/proxyproto.go` | PROXY v1/v2 解析器 |
| `util/proxyproto/proxyproto_test.go` | 解析器单测 |
| `core/proxyprotocol.go` | 按端口策略、可信 CIDR、读头、接回余字节 |
| `core/proxyprotocol_test.go` | 预处理器单测 |
| `core/proxy_e2e_test.go` | SOCKS/WS/HTTP/HTTPUpgrade/gRPC/REALITY 端到端 |
| `core/tracker_ip_test.go` | IP 计数、限额、归并、并发 |
| `core/tracker_session.go` | IP 计数与限制实现 |
| `core/main.go` | 策略注入、IP 限额/前缀/演练状态 |
| `service/inbounds.go` | proxyProtocol 校验、持久化、策略生成 |
| `service/client.go` | ipLimit 保存/查询 |
| `service/config.go` | 启动/重启/保存后的同步 |
| `service/stats.go` | `GetActiveIPCounts`、`GetUserIPs` |
| `api/apiService.go` | sessions/client、userIps、ipCounts |
| `frontend/src/layouts/modals/Sessions.vue` | 会话来源列 |
| `frontend/src/views/Clients.vue` | IP 数列与明细入口 |
| `frontend/src/layouts/modals/UserIps.vue` | IP 明细弹窗 |
| `frontend/src/layouts/modals/Client.vue` | IP 限制输入 |
| `frontend/src/components/ProxyProtocol.vue` | 入站 PROXY 配置 |
| `frontend/src/views/Settings.vue` | 前缀与演练开关 |
| `docs/real-ip-worklog.md` | 本文 |
