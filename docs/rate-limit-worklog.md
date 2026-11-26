# 客户端上下行限速 — 工作记录与代码验收说明

> 本文档是「按客户端聚合的独立上行/下行限速」功能的独立交付记录，供另一个 Agent 做代码验收。
>
> - 日期：2026-11-24
> - 功能状态：已完成，已构建并运行于本机 2095 面板
> - 第一轮提交：后端 `936fdd4`、前端 `b08029a`
> - 第二轮修复提交：后端 `2c58566`、前端 `40efcbe`（`work/antd-tls-layout`）
> - 前端子模块指针：`2c58566` 已包含 `40efcbe`
> - 内嵌构建：`assets/b4447417a30f3621.js`
> - 知识图谱：第二轮代码提交时 1296 节点 / 4702 边；当前哈希以 `.ua/meta.json` 为准
> - 运行时：本机 macOS loopback + SOCKS 入站；`1 Mbps = 125,000 B/s`

---

## 1. 目标与结论

### 1.1 需求

1. 每个客户端可以单独设置上行、下行限速，单位整数 Mbps，默认 0 = 不限速。
2. 同一客户端（`metadata.User`）的所有入站、所有 TCP/UDP 会话共用一份上行额度、一份下行额度。
3. 客户端侧读算上传，写算下载；TCP 覆盖 `Read/Write` 和 `ReadBuffer/WriteBuffer`，UDP 覆盖 `ReadPacket/WritePacket`。
4. 重连不能刷新额度；长连接在线改速必须生效，包括 0 → 限速 → 0。
5. 单客户端添加/编辑可设置；批量编辑缺字段必须保留原值，显式 0 才清除。
6. 旧数据默认 0；仅设置速率、没有设置 `maxConns` 的客户端也必须在启动/保存时加载到 tracker。
7. 验收包含单元测试、真实代理链路 60 秒吞吐、第二用户独立、在线改速、流量统计、`maxConns`、TCP 半关闭，以及知识图谱更新。

### 1.2 结论

全部实现并通过本机验收：

- 60 秒跨两个 SOCKS 入站、同一用户 TCP+UDP 并发：**上行 2.000 Mbps，下行 7.999 Mbps**。
- UDP-only：**上行 2.001 Mbps，下行 8.001 Mbps**。
- 两个用户同时下载：**8.021 / 7.969 Mbps**，互不影响。
- 长连接在线改速：8 Mbps → **7.995**，2 Mbps → **1.966**，0（不限速）→ **67,382 Mbps**（loopback 上界）。
- `maxConns`、TCP 半关闭、流量统计均通过。
- 前端添加 2/8、编辑显式改 0 均正确落库。

---

## 2. 需求 → 实现映射

| 需求 | 实现位置 |
|---|---|
| 新增 `upMbps`/`downMbps`，0 不限速 | `database/model/model.go:61-62` |
| 单客户端表单两个输入框 | `frontend/src/layouts/modals/Client.vue:184-218`、`:633-640` |
| 单编辑缺字段保留、显式 0 清除 | `service/client.go:125-140`、`service/client.go:224-229` |
| 批量编辑缺字段保留 | `service/client.go:142-190`、`service/client.go:302-311` |
| 仅设置速率的用户可加载 | `service/client.go:78-99` |
| 启动/保存同步速率 | `service/config.go:331-347`、`core/main.go:111-133` |
| 按用户共享上下行桶 | `core/tracker_session.go:168-244` |
| 热更新 / 0↔限速 | `core/tracker_session.go:48-55`（原子指针）、`:168-205` |
| TCP 读/写/缓冲计费 | `core/tracker_session.go:273-338` |
| UDP 包读写计费且保持整包 | `core/tracker_session.go:340-388` |
| 连接关闭取消等待 | `core/tracker_session.go:665-705` |
| 限速器位于计数器内侧、不被 splice 绕过 | `core/tracker_session.go:479-530` |
| 单元测试 | `core/tracker_rate_test.go`、`service/client_rate_test.go` |
| 真实代理 E2E | `core/rate_e2e_test.go` |
| 握手成功/失败回调穿透限速层 | `core/tracker_session.go:433-443`、`:506-508` |
| 只更新变化方向并唤醒排队请求 | `core/tracker_session.go:213-305` |
| 32 位安全换算 / 17180 范围错误 | `util/rateconv/rateconv.go:17-34`、`service/client.go:105-124` |
| 两阶段（先上传后下载）E2E | `core/rate_e2e_test.go:TestRateLimitE2EThroughput` |
| 回显式 maxConns 验收 | `core/rate_e2e_test.go:TestRateLimitE2EMaxConnsAndHalfClose` |

---

## 3. 数据流

```
SQLite clients.up_mbps / down_mbps
        │
        ├─ GetAll() ───────────────► /api/load 的 clients 列表（含速率，供编辑/批量编辑）
        │
        └─ GetRateLimitsMap() ─────► syncUserLimits()
                                          │
                                          ├─ corePtr.SetUserRates(map[name]UserRate)
                                          │       │
                                          │       ├─ Core.userRates（重启保留）
                                          │       └─ 运行中：SessionTracker.SetUserRates()
                                          │
                                          ▼
                              SessionTracker.userRates[name] = &userRate{up,down atomic.Pointer}
                                          │
                        新会话 newSession() 取同一个 *userRate
                                          │
                    ┌─────────────────────┴─────────────────────┐
                    ▼                                           ▼
              RoutedConnection                          RoutedPacketConnection
        raw conn → rateConn → CounterConn → sessionConn     raw PacketConn → ratePacketConn → CounterPacketConn → sessionPacketConn
```

要点：限速器必须放在 **Counter 内侧**。`CounterConn` 本身不实现 `ReaderReplaceable/WriterReplaceable`，copy/splice 的 unwrap 会在 Counter 处停下；其读路径会自然经过 `rateConn`。如果把限速放在 Counter 外侧，unwrap 会绕过它。

---

## 4. 后端实现细节

### 4.1 数据模型

`database/model/model.go:61-62`

```go
// Per-client bandwidth limits in Mbps (0 = unlimited)
UpMbps   int `json:"upMbps" form:"upMbps" gorm:"default:0;not null"`
DownMbps int `json:"downMbps" form:"downMbps" gorm:"default:0;not null"`
```

- GORM AutoMigrate 自动新增两列，旧行为 NULL/0。
- JSON 字段 `upMbps`/`downMbps`。
- 单位是十进制 Mbps，内部 `mbps * 125_000` 转字节/秒。

### 4.2 列表读取、保存与缺字段语义

**列表读取**：`service/client.go:43` 的 `Select` 必须带 `up_mbps, down_mbps`。否则 `/api/load` 的客户端对象会带 0，前端批量编辑把整个对象回传时会误清零。这是本次专门修的一个坑。

**单编辑**（`service/client.go:224-229`）：

```go
if act == "edit" {
    s.preserveMissingRates(tx, &client, data)
}
if client.MaxConns < 0 { client.MaxConns = 0 }
clampClientRates(&client)
```

`preserveMissingRates`（`:125-140`）：
- 用 `clientRateFieldsPresent(raw)` 检查 JSON 是否有 `upMbps` / `downMbps`。
- 缺哪个字段，就从数据库旧行补哪个。
- 显式 `0` 字段存在 → 不补 → 清除。
- 负数由 `clampClientRates` 归零。

**批量编辑**（`:142-190`、`:302-311`）：
- 把 `data` 再解析成 `[]json.RawMessage`。
- 收集缺字段的客户端 id，一次 `WHERE id IN ?` 查旧速率。
- 只补缺失字段；显式 0 保留。
- 然后逐个 `clampClientRates`。

**新增 / 批量新增**：缺字段自然为 0，负数归零。

### 4.3 同步链路

`service/client.go:78-99`：

```sql
SELECT name, up_mbps, down_mbps
FROM clients
WHERE up_mbps > 0 OR down_mbps > 0
```

- 条件**不包含** `max_conns`，所以“只设置速率”的客户端也会被加载。
- `service/config.go:331-347` 在客户端保存事务提交后、以及核心启动/重启前调用：

```go
limits, _ := GetMaxConnsMap(...)   // 并发上限，保持原逻辑
corePtr.SetUserLimits(limits)

rates, _ := GetRateLimitsMap(...)  // 本次新增
corePtr.SetUserRates(rates)
```

- 两条查询独立；一条失败不阻止另一条生效。
- `Core.SetUserRates`（`core/main.go:111-133`）保存一份仅供重启使用的副本，并在核心运行时立即推送 tracker。
- `Core.Start`（`core/main.go:98-99`）在实例启动后把 `userRates` 应用到新的 `SessionTracker`。

### 4.4 运行时桶与改速

`core/tracker_session.go:31-101`：

```go
type limiterBox struct {
    limiter rateWaiter
    changed chan struct{} // closed when this box is replaced
}

type rateLimiter struct {
    limiter *rate.Limiter
    changed <-chan struct{}
}

type userRate struct {
    up       atomic.Pointer[limiterBox]
    down     atomic.Pointer[limiterBox]
    upMbps   int // owned by SessionTracker.access
    downMbps int
}
```

- 每个客户端名一个 `*userRate`，所有会话共享；两个方向各一个原子指针，`nil` = 该方向不限速。
- `rateStateLocked` 在第一次会话创建时也建一个空状态，保证“初始不限速”的长连接后续可被 0 → 限速热更新。
- `SetUserRates`（`:213-232`）与 `applyUserRateLocked`（`:234-277`）：
  - 对比 `state.upMbps/downMbps` 与目标值，**只替换真正变化的方向**；
  - 未变化的用户/方向保留原 `limiterBox`，令牌数不会被重置（保存用户 B 不会让 A 额外获得 burst）；
  - 替换时 `Swap` 新 box 并关闭旧 box 的 `changed`，唤醒排队中的 `WaitN`；
  - 只改下载时上传指针与 token 完全不动；
  - 从限速改为 0 时换成 nil 并关闭旧 `changed`，排队请求立即放行。
- `newRateLimiter`（`:279-305`）使用 `util/rateconv.BytesPerSecond` 做 int64 换算，`17180 Mbps` 返回范围错误；burst 为 `clamp(bytesPerSecond/8, 16 KiB, 256 KiB)`。
- `rateLimiter.WaitN`（`:52-86`）用 `ReserveN` + timer/context/changed 三路 select；被改速唤醒时 `Cancel` 预留并返回 `errRateChanged`。`waitN`（`:307-335`）每次取当前 limiter，改速时**重试同一块**而不是重复扣费。
- 重连不会新建桶；同一 `*userRate` 的 limiter 持续存在，令牌按时间补充但不重置。
- 限速状态 map 不主动删除，条目数与曾出现过的客户端名同阶；这是为热更新保留旧会话引用的取舍。

### 4.5 TCP/UDP 包装与握手

**TCP**（`core/tracker_session.go:603-630`）：

```
raw net.Conn
  → bufio.NewExtendedConn
  → rateConn（若用户有 rate 状态）
  → bufio.NewInt64CounterConn
  → sessionConn（含 cancel）
```

`rateConn`（`:337-445`）：
- `Read`：客户端 → 服务端，算上传；最多读 32 KiB，读完后 `waitN(n)`。
- `Write`：服务端 → 客户端，算下载；按 32 KiB 分块，**写前** `waitN(len(chunk))`。
- `ReadBuffer`：读完后对 `buffer.Len()` 计上传。
- `WriteBuffer`：写前对 `buffer.Len()` 计下载。
- `waitN` 按当前 `limiter.Burst()` 分块，单次 `WaitN` 永不超过 burst；改速时通过 loader 重取当前桶。
- 关闭：`sessionConn.Close` 先 `cancel()` 再 `leave`，等待立即返回；等待期间不持 `SessionTracker.access` 锁。
- `FrontHeadroom/RearHeadroom` 用 `N.CalculateFrontHeadroom/RearHeadroom(inner)` 递归向内取，外包一层仍保留 SOCKS 等协议需要的头部空间。
- **握手回调**：`HandshakeSuccess`、`HandshakeFailure`、`ConnHandshakeSuccess`（`:433-443`）直接向内转发。`common.Cast` 会从 `sessionConn → CounterConn → rateConn` 停下，因此回调能送达 SOCKS，而 copy 数据仍必须经过 `rateConn`。`ratePacketConn` 也有对应的 `PacketConnHandshakeSuccess`（`:506-508`）。

**UDP**（`core/tracker_session.go:447-508`）：

```
raw N.PacketConn
  → ratePacketConn
  → bufio.NewInt64CounterPacketConn
  → sessionPacketConn
```

- `ReadPacket`：读完整包后再按 `buffer.Len()` 计上传。
- `WritePacket`：先按整包长度申请令牌，再一次性写出；令牌可分多次申请，但**报文不拆分**。
- 同样转发 `FrontHeadroom/RearHeadroom` 与握手回调。

### 4.6 热更新、重连、关闭

- **重连**：`userRate` 存在 `SessionTracker.userRates`，新会话复用同一个对象，不会新建桶。
- **0 → 限速**：初始 `userRate` 已创建（空指针），`SetUserRates` 填充指针，旧连接立即看到。
- **限速 → 0**：指针置 nil，旧连接立即跳过等待。
- **改速**：新 limiter 指针替换，旧连接下一字节起用新桶。
- **关闭**：连接 context 取消，`WaitN` 返回；不会出现低速连接关闭后 goroutine 长时间卡住。
- **锁**：等待只操作 limiter；`SessionTracker.access` 只在创建/更新状态和会话表时短暂持有。

---

## 5. 前端实现

### 5.1 类型与默认值

`frontend/src/types/clients.ts:32-33`、`:58-59`：

```ts
upMbps?: number
downMbps?: number
```

`defaultClient` 默认 0。

### 5.2 表单

`frontend/src/layouts/modals/Client.vue:184-218` 在「使用限制」组、连接数限制下方新增一行：

- 上行限速：`a-input-number`，min=0，step=1，`addon-after="Mbps"`，placeholder `0`。
- 下行限速：同上。
- 提示：`0 表示不限速，按客户端所有入站合计`（6 种语言）。

计算属性（`:633-640`）做非负/取整：

```ts
upMbps: {
  get() { return this.client.upMbps ?? 0 },
  set(v: number | null) { this.client.upMbps = v && v > 0 ? Math.floor(v) : 0 }
}
```

- 单客户端添加/编辑都使用同一个弹窗。
- 批量编辑没有新增速率输入；批量编辑会回传完整 client 对象，列表现在包含速率字段，所以不会误清零；缺字段场景由后端保留。

### 5.3 API

没有新增端点。`/api/load` 的 `clients` 数组自动带 `upMbps/downMbps`（依赖 `GetAll` 的 Select 已补列）；`/api/save` 直接反序列化 `model.Client`。

---

## 6. 限速参数与语义

| 项目 | 取值 |
|---|---|
| 单位 | 整数 Mbps |
| 换算 | `1 Mbps = 1,000,000 bit/s = 125,000 B/s` |
| 0 | 不限速 |
| 负数 | 保存时归零 |
| 最大可保存值 | **17179 Mbps**（17180 返回明确范围错误） |
| TCP 单次 I/O 块 | 32 KiB |
| 令牌申请 | 按 `limiter.Burst()` 分块，单次 ≤ burst |
| burst | `clamp(bytesPerSecond/8, 16 KiB, 256 KiB)` |
| UDP | 报文保持完整；大包可分多次申请令牌后整体发送 |
| 方向 | 客户端侧读 = 上传；客户端侧写 = 下载 |
| 共享范围 | 同一 `metadata.User` 的所有入站、所有会话 |
| 等待取消 | 会话 context，连接关闭即取消 |
| 等待取消 | 会话 context，连接关闭即取消 |
| 改速唤醒 | 关闭旧桶的 `changed`，排队 `WaitN` 取消预留并重试新桶 |

---

## 7. 测试与实测

### 7.1 单元测试

`core/tracker_rate_test.go`（假 limiter，不依赖真实时间）：

- `TestWaitNChunksAtBurst`：25 字节、burst 10 → `[10,10,5]`。
- `TestWaitNCancels`：已取消 context 立即返回。
- `TestRateConnDirectionsAndBuffers`：TCP `Read` 只扣上行，`Write` 只扣下行；`ReadBuffer/WriteBuffer` 同样。
- `TestRatePacketConnKeepsPacketWhole`：UDP 报文内容不变，令牌分两次申请。
- `TestSetUserRatesHotUpdateAndSharedState`：初始空状态 → 限速 → 显式 0 → 单向限速；同一用户状态复用。
- `TestRateConnShareOneBucketAcrossSessions`：两条连接共享一个桶。
- `TestRateConnShareOneBucketAcrossSessions`：两条连接共享一个桶。
- `TestNewRateLimiterBoundary`：17179 Mbps 成功，17180 Mbps 返回范围错误。
- `TestRateLimiterInterruptsQueuedWait`：排队 `WaitN` 在 changed 关闭后 500ms 内返回 `errRateChanged`。
- `TestSetUserRatesOnlyChangesModifiedDirection`：保存 B 不替换 A；只改下载不重置上传；上传改 0 不动下载。
- `TestSavingAnotherUserDoesNotRefillBurst`：先耗尽 A 的 burst，再保存 B，A 的令牌不被补满。
- `TestRateConnForwardsHandshakeCallbacks`：`ConnHandshakeSuccess` / `HandshakeFailure` 穿透 rateConn。
- `TestRateConnStillLimitsAfterHandshake`：握手回调后数据仍经过限速器。

`service/client_rate_test.go`（临时 SQLite）：

- `TestGetRateLimitsMapIncludesRateOnlyClients`：rate-only 客户端进入映射，maxConns-only / 不限速不进入。
- `TestPreserveMissingRates`：单编辑缺字段保留、显式 0 清除。
- `TestPreserveMissingRatesBulk`：批量混合缺字段/显式 0。
- `TestPreserveMissingRatesBulk`：批量混合缺字段/显式 0。
- `TestValidateClientRatesBoundary`：17179 通过、17180 拒绝、负数归零。
- `TestGetRateLimitsMapRejectsOverflow`：数据库存在 17180 时返回范围错误。
- `TestSaveRejectsOverflowRate`：`Save` 路径拒绝 17180。

`util/rateconv/rateconv_test.go`：

- `TestBytesPerSecondBoundary`：17179 → 2,147,375,000；17180 → `ErrTooLarge`。
- `Test32BitTarget`：在 `GOARCH=386` 下断言 `strconv.IntSize == 32` 且两个边界结果正确。

### 7.2 E2E 环境

`core/rate_e2e_test.go`（`//go:build ratee2e`）：

- 内嵌核心启动两个 SOCKS 入站，用户 `alice`/`bob`/`carol`。
- 本机 loopback 上的 TCP sink/source、UDP sink/source；UDP source 支持同时向多个入站回发。
- 最小 SOCKS5 客户端（用户名/密码、CONNECT、UDP ASSOCIATE）。
- 同一用户 `alice`：
  - **上传阶段**：两个入站同时各跑一条 TCP 上传 + 一条 UDP 上传，共 4 条流；
  - **下载阶段**：两个入站同时各跑一条 TCP 下载 + 一条 UDP 下载，共 4 条流；
  - 两个阶段均预热 5 秒、统计 60 秒；上传按 sink 接收字节，下载按客户端接收字节。
- 冒烟测试覆盖 `0/0` 和 `2/8 Mbps` 两种配置，并额外验证到一个关闭端口的 SOCKS 握手失败回调能在 5 秒内返回。
- `maxConns=1` 使用回显服务：第一条发送并读回，第二条必须被拒绝且不能是读超时；关闭第一条后第三条可传输。
- 半关闭用“读到 EOF 后再写 done”的 TCP server 验证。

### 7.3 实测结果（第二轮最终数据）

| 测试 | 结果 |
|---|---|
| 上传阶段（两入站、TCP+UDP，60s） | **2.000 Mbps**（TCP 14,221,095 B + UDP 780,000 B） |
| 下载阶段（两入站、TCP+UDP，60s） | **7.999 Mbps**（TCP 58,915,965 B + UDP 1,080,000 B） |
| UDP-only 10 秒 | 上传 **2.001 Mbps**；下载 **8.000 Mbps** |
| 双用户同时下载 | alice **8.021 Mbps**；bob **8.021 Mbps** |
| 在线改速（单条长连接） | 8 Mbps → **7.995**；2 Mbps → **1.966**；0 → **62,604.882 Mbps** |
| 4 条连接排队后解除限速 | 排队时 **1.049 Mbps**；`SetUserRates(0)` 后 **100 ms**（采样窗）达到 >20 Mbps |
| 流量统计 | 两阶段 tracker 用户流量 **86,206,071 B** |
| `maxConns=1` / 半关闭 | 第一条回显成功、第二条被拒绝（非超时）、释放后可再连；半关闭收到 `done` |

说明：
- 上传和下载改为**先上传 60 秒、再下载 60 秒**，两个方向不再同时抢占同一用户的上下行桶；每个阶段都由两个入站共同承担 TCP+UDP。
- 下载阶段 TCP 仍占多数，UDP 约 1.04 MB；UDP-only 测试单独证明 UDP 可以达到 2/8 Mbps。
- 不限速 62,604.882 Mbps 是 loopback 本机上界，仅作为“没有令牌桶时”的对照，不代表公网能力；不限速上行未单独做 60 秒测试。
- 100 ms 是测试的采样窗上限；实际唤醒发生在 `SetUserRates` 返回后的第一个 100 ms 窗口内。

### 7.4 可复现命令

```bash
cd /Users/yuzai/Tools/ip_api/s-ui

# 全部 Go 测试
go test ./... -count=1

# 冒烟：0/0 与 2/8，10 次，数据竞争检测
go test -race ./core -tags ratee2e -run TestRateLimitE2ESmoke -count=10 -timeout 10m

# 整套限速 E2E（约 171 秒）
go test ./core -tags ratee2e -run 'TestRateLimitE2E' -v -count=1 -timeout 20m

# 单独运行
go test ./core -tags ratee2e -run TestRateLimitE2EThroughput -v -count=1 -timeout 6m
go test ./core -tags ratee2e -run TestRateLimitE2EUDPOnly -v -count=1 -timeout 3m
go test ./core -tags ratee2e -run TestRateLimitE2ELiveChange -v -count=1 -timeout 3m
go test ./core -tags ratee2e -run TestRateLimitE2EQueuedUnlimitedLatency -v -count=1 -timeout 3m
go test ./core -tags ratee2e -run TestRateLimitE2EMaxConnsAndHalfClose -v -count=1 -timeout 3m
```

32 位目标复验（本机 macOS arm64，使用 QEMU system 模式 + Alpine x86 内核/用户态）：

```bash
# 1) 交叉编译 32 位测试二进制（静态）
cd /Users/yuzai/Tools/ip_api/s-ui
CGO_ENABLED=0 GOOS=linux GOARCH=386 go test -c -o /tmp/core386.test ./core
CGO_ENABLED=0 GOOS=linux GOARCH=386 go test -c -o /tmp/rateconv386.test ./util/rateconv

# 2) 放入 Alpine x86 initramfs，init 中执行：
#    /rateconv386.test -test.v -test.run 'TestBytesPerSecondBoundary|Test32BitTarget'
#    /core386.test -test.v -test.run 'TestNewRateLimiterBoundary|TestSetUserRatesOnlyChangesModifiedDirection|TestSavingAnotherUserDoesNotRefillBurst|TestRateLimiterInterruptsQueuedWait'
# 3) 启动
/opt/homebrew/bin/qemu-system-i386 -m 512 -kernel /tmp/alpine386/vmlinuz   -initrd /tmp/alpine386/initramfs-final.cpio.gz   -append 'console=ttyS0 panic=-1' -nographic -no-reboot
```

预期输出：

```text
=== 32-bit rateconv boundary ===
--- PASS: TestBytesPerSecondBoundary
--- PASS: Test32BitTarget
=== 32-bit core rate tests ===
--- PASS: TestNewRateLimiterBoundary
--- PASS: TestRateLimiterInterruptsQueuedWait
--- PASS: TestSetUserRatesOnlyChangesModifiedDirection
--- PASS: TestSavingAnotherUserDoesNotRefillBurst
=== all 32-bit tests done ===
```

## 8. 已知边界、未验证路径与评审关注点

### 8.1 未验证

- **Hysteria2 / TUIC**：已核对 sing-box 1.14.4 源码，二者都会在 `NewConnectionEx` / `NewPacketConnectionEx` 写入 `metadata.User` 并调用 `RouteConnectionEx` / `RoutePacketConnectionEx`，因此经过同一个 tracker 包装层；但**没有用真实 Hysteria2/TUIC 客户端跑吞吐**。
- **TUN / WireGuard**：这些路径通常没有认证用户，`metadata.User` 为空，不会按客户端限速；这是覆盖边界，不是 bug。
- **前端批量编辑速率 UI**：没有新增输入框；后端缺字段保留由单元测试覆盖。
- **不限速上行 60 秒对照**：未单独测量。
- **不限速上行 60 秒对照**：未单独测量。
- **32 位全链路 E2E**：只在 386 下运行了换算边界、限速更新/唤醒等单元测试，没有在 386 下跑 60 秒 SOCKS E2E；全链路 E2E 在 arm64 主机完成。
- **握手上游异常形态**：冒烟用“连接到关闭端口”验证失败回调；没有覆盖真实上游 TLS/协议握手失败的每一种错误类型。

### 8.2 评审时值得重点看的点

1. **TCP 部分写**：`rateConn.Write` 在写前按 `len(chunk)` 申请令牌；若底层 `Write` 只写了 `n < len(chunk)`，这一块会多扣少量令牌（只影响短时精度，不影响 60 秒平均）。如果要求严格逐字节计费，可改为写后按 `n` 扣费或用 `ReserveN` + 取消未用部分。
2. **`userRates` 不删除**：状态按客户端名保留，以支持旧会话热更新；规模与客户端名总数同阶，不随连接数增长。
3. **快速路径接口**：`rateConn/ratePacketConn` 故意不实现 `Upstream`/`ReaderReplaceable`，并隐藏 `ReadWaiter`/MTU 等可选接口，copy/splice 会回退到 buffer 路径。这是为了不绕过限速；E2E 已证明功能正确，但高吞吐无限制场景可能有性能损失。
4. **`FrontHeadroom` 转发**：用 `N.CalculateFrontHeadroom(inner)` 递归取内层值，修复 SOCKS UDP 头预留；如果后续新增协议依赖 `LazyHeadroom`，需要再评估。
5. **UDP 共享桶竞争**：并发 TCP 时 UDP 可能被 TCP 抢走额度；共享桶语义如此，UDP-only 测试证明 UDP 本身可达标。
6. **批量编辑的其它字段**：`GetAll` 仍只 Select 客户端列表需要的子集（这是既有设计）；本次只保证速率字段被列表携带，批量编辑其它字段的语义未在本功能范围内改动。
7. **E2E 端口**：使用随机空闲 TCP 端口；SOCKS 端口通过 `freeTCPPort` 获取，测试期间若被抢占会失败。
7. **E2E 端口**：使用随机空闲 TCP 端口；SOCKS 端口通过 `freeTCPPort` 获取，测试期间若被抢占会失败。
8. **`ReserveN` 取消**：改速时排队请求会 `Cancel` 预留，令牌返还旧桶后重试新桶；如果连续快速改速，可能出现多次重试，但每次重试最多一个 burst 块（≤32 KiB）。
9. **握手转发**：`rateConn` 没有实现 `Upstream`，而是直接转发 `HandshakeSuccess/Failure/ConnHandshakeSuccess`；这保证 `common.Cast` 停在 rateConn，数据不会绕过限速。新增协议若依赖其他 handshake 接口，需要再检查。
10. **32 位上限**：上限是 17179 Mbps（受 signed int32 约束），不是“业务上希望的最大值”；若未来需要更高，需要把 `rate.Limit` 的输入改为 float64 或明确用 uint64，当前选择是返回范围错误。

---

## 9. 改动文件清单

### 9.1 后端（第一轮 `936fdd4`，第二轮 `2c58566`）

| 文件 | 行/位置 | 作用 |
|---|---|---|
| `database/model/model.go` | `:61-62` | `Client.UpMbps/DownMbps` 两列 |
| `core/tracker_session.go` | `:31-101` | `UserRate`、`rateWaiter`、`limiterBox`、`rateLimiter`、`userRate` 类型 |
| `core/tracker_session.go` | `:213-305` | `SetUserRates` 只改变变化方向、关闭旧 `changed`、唤醒排队；`newRateLimiter`/`waitN` |
| `core/tracker_session.go` | `:337-445` | `rateConn`：TCP Read/Write/Buffer、headroom、握手转发 |
| `core/tracker_session.go` | `:447-508` | `ratePacketConn`：UDP ReadPacket/WritePacket/headroom/握手 |
| `core/tracker_session.go` | `:603-630` | 在计数器内侧安装限速包装；`session.rate` 取共享状态 |
| `core/tracker_session.go` | `:789-829` | `sessionConn/sessionPacketConn.Close` 先 cancel 等待 |
| `core/main.go` | `:29`、`:98-99`、`:111-133` | Core 保存 `userRates`；启动/热更新应用 |
| `service/client.go` | `:43` | `GetAll` 补 `up_mbps/down_mbps` 列 |
| `service/client.go` | `:79-103` | `GetRateLimitsMap`（rate-only 也能加载，超范围报错） |
| `service/client.go` | `:105-190` | 负数归零、17179 上限校验、字段存在性检测、单/批量缺字段保留 |
| `service/client.go` | `:240-245`、`:317-322` | 单编辑/批量编辑调用校验与保留 |
| `service/config.go` | `:331-347` | `syncUserLimits` 同时推送 maxConns 与速率 |
| `util/rateconv/rateconv.go` | `:17-34` | int64 换算、`MaxMbps=17179`、`ErrTooLarge` |
| `go.mod` | `x/time` | 改为直接依赖 |

### 9.2 后端测试

| 文件 | 作用 |
|---|---|
| `core/tracker_rate_test.go` | 假 limiter 单元测试：分块、取消、方向、buffer、UDP 整包、共享桶、热更新 |
| `service/client_rate_test.go` | 临时 SQLite：rate-only 映射、单/批量缺字段保留 |
| `core/rate_e2e_test.go` | `//go:build ratee2e`，真实 SOCKS 链路 60 秒 E2E、双用户、改速、maxConns、半关闭、UDP-only |
| `core/rate_e2e_test.go` | `//go:build ratee2e`：冒烟 0/0+2/8、两阶段上传/下载 60s、双用户、改速、maxConns 回显、半关闭、UDP-only、排队解除限速延迟 |
| `util/rateconv/rateconv_test.go` | 17179/17180 边界，386 目标断言 |

### 9.3 前端（第一轮 `b08029a`，第二轮 `40efcbe`）

| 文件 | 行/位置 | 作用 |
|---|---|---|
| `frontend/src/types/clients.ts` | `:32-33`、`:58-59` | Client 类型与默认值 |
| `frontend/src/layouts/modals/Client.vue` | `:184-218` | 使用限制组的两个 Mbps 输入（上限 17179）+ 提示 |
| `frontend/src/layouts/modals/Client.vue` | `:633-640` | `upMbps/downMbps` 非负取整 computed |
| `frontend/src/locales/zhcn.ts` | `client.upRate/downRate/rateHint` | 中文文案 |
| `frontend/src/locales/zhtw.ts` | 同上 | 繁中文案 |
| `frontend/src/locales/en.ts` | 同上 | 英文文案 |
| `frontend/src/locales/ru.ts` | 同上 | 俄文文案 |
| `frontend/src/locales/vi.ts` | 同上 | 越南文文案 |
| `frontend/src/locales/fa.ts` | 同上 | 波斯文文案 |

### 9.4 文档/构建

| 文件 | 作用 |
|---|---|
| `AGENTS.md` | 追加本次功能工作记录 |
| `frontend/dist` → `web/html/` | 构建产物（gitignored） |
| `sui` | Go 二进制（gitignored） |
| `.ua/` | Understand-Anything 图谱数据（untracked） |

---

## 10. 验收检查清单

建议另一个 Agent 按此顺序做：

- [ ] `git show --stat 2c58566`、`git -C frontend show --stat 40efcbe`，确认本轮文件范围。
- [ ] 确认 `Client` 模型两列、AutoMigrate 后旧数据为 0。
- [ ] 检查 `GetAll` 的 Select 是否包含 `up_mbps/down_mbps`。
- [ ] 检查 `GetRateLimitsMap` 的 WHERE 不含 `max_conns`，且 >17179 返回范围错误。
- [ ] 手工或单测验证：单编辑缺字段保留、显式 0 清除；批量同上。
- [ ] 检查 `syncUserLimits` 在保存提交后与核心启动前都调用，且 `SetUserLimits`/`SetUserRates` 都推送。
- [ ] 检查 `SetUserRates` 只在 `upMbps/downMbps` 变化时替换对应桶；未变化桶的指针和 token 不变。
- [ ] 检查改速时旧 `changed` 关闭、`rateLimiter.WaitN` 取消预留、`waitN` 重试同一块。
- [ ] 检查握手成功/失败/`ConnHandshakeSuccess` 能从 `rateConn` 转发，且数据仍经过限速器。
- [ ] 检查 `newRateLimiter` 使用 int64 换算，17180 返回错误；前端输入上限 17179。
- [ ] 检查 `RoutedConnection/PacketConnection` 的包装顺序：`rateConn` 在 Counter 内侧。
- [ ] 检查 `rateConn/ratePacketConn` 的 Read/Write/Buffer/Packet 方向。
- [ ] 检查 `Close` 先 cancel 且等待不持全局锁。
- [ ] 检查 `FrontHeadroom/RearHeadroom` 转发。
- [ ] 运行 `go test ./... -count=1`。
- [ ] 运行 `go test -race ./core -tags ratee2e -run TestRateLimitE2ESmoke -count=10 -timeout 10m`，确认无 DATA RACE。
- [ ] 运行 `go test ./core -tags ratee2e -run 'TestRateLimitE2E' -v -count=1 -timeout 20m`，核对第 7.3/13 节数字（允许合理浮动）。
- [ ] 运行 386 QEMU 命令，核对 17179/17180 边界。
- [ ] `cd frontend && npm test && npm run build`。
- [ ] 检查 2095 资产 `assets/b4447417a30f3621.js`。
- [ ] 检查 `.ua/meta.json` 的 `gitCommitHash` 与当前 HEAD 一致。

---

## 11. 回滚

- 代码：回退第二轮 `2c58566`（含前端指针）与前端 `40efcbe`；再需要时回退第一轮 `936fdd4` / `b08029a`。
- 数据库：新增两列可保留；把 `up_mbps/down_mbps` 全部置 0 即恢复不限速。GORM 不会自动删列，回滚代码后旧版本会忽略这两列。
- 构建：`frontend/dist` 重新构建，`web/html` 同步，`go build -o sui main.go` 重启。

---

## 12. 环境与已知本机状态

- 面板：`http://127.0.0.1:2095/app/`，前端开发服务 `http://127.0.0.1:3000/`。
- 本机数据库 `db/s-ui.db` 新建列已存在；现有客户端 `WdkM3DF6` 速率为 0/0，非本次测试数据。
- 本机 `db/s-ui.db` 中本次 UI 测试客户端已删除。
- E2E 使用随机端口、本地 loopback，不依赖外网。
- 当前构建资产：`assets/b4447417a30f3621.js`。
- 知识图谱：以 `.ua/meta.json` 为准（第二轮代码提交时为 274 个分析文件、1296 节点、4702 边）。


---

## 13. 第二轮验收修复记录（2026-11-24）

- 后端修复提交：`2c58566`（含前端子模块指针）
- 前端修复提交：`40efcbe`（`frontend` 子模块 `work/antd-tls-layout`）
- 第一轮提交：`936fdd4` / `b08029a`
- 本文档提交：本轮 `docs: update the rate limit review worklog`（具体 hash 以 git log 为准）
- 最终全量 E2E：`go test ./core -tags ratee2e -run 'TestRateLimitE2E' -v -count=1 -timeout 20m`，7 个测试全部通过，总耗时 171.342s。

### 13.1 SOCKS 握手竞争

**问题**：`rateConn`/`ratePacketConn` 包在 Counter 内侧，但只嵌入了 `N.ExtendedConn`/`N.PacketConn`，没有转发 `HandshakeSuccess`、`HandshakeFailure`、`ConnHandshakeSuccess`、`PacketConnHandshakeSuccess`。`common.Cast` 从 `sessionConn → CounterConn` 向内走时在 rate 包装器处停下，SOCKS 握手回执回调丢失；同时包装器不能让 `Upstream()` 暴露，否则数据会绕过限速。

**修复**：
- `core/tracker_session.go:433-443`：`rateConn` 增加 `HandshakeSuccess`、`HandshakeFailure`、`ConnHandshakeSuccess`，用 `N.Report*` 向内转发。
- `core/tracker_session.go:506-508`：`ratePacketConn` 增加 `PacketConnHandshakeSuccess` 及成功/失败转发。
- 不实现 `Upstream()`；`common.Cast` 停在 rateConn，回调送达，但 copy 数据仍经过 rateConn。
- 单元测试 `TestRateConnForwardsHandshakeCallbacks`、`TestRateConnStillLimitsAfterHandshake`。
- 冒烟测试覆盖 `0/0`、`2/8 Mbps`，并额外连接 `127.0.0.1:1` 验证失败回调在 5 秒内返回、不是超时。

**复验命令与结果**：

```bash
go test -race ./core -tags ratee2e -run TestRateLimitE2ESmoke -count=10 -timeout 10m
```

实测：`ok github.com/alireza0/s-ui/core 2.461s`，无 `DATA RACE`。10 次运行中每次均执行 `unlimited-0-0` 与 `limited-2-8` 两个子测试，并各执行一次关闭端口的握手失败。

**仍未验证**：只覆盖了“连接关闭端口”的失败形态；真实上游 TLS/协议握手失败的不同错误类型没有逐一注入。

### 13.2 令牌桶重置与改速延迟

**问题**：原 `SetUserRates` 对 map 中每个正速率用户无条件 `Store(newRateLimiter(...))`，保存用户 B 会替换 A 的桶并补满 burst；已经阻塞在旧桶上的 `WaitN` 不会因为指针替换而醒来。

**修复**：
- `userRate` 增加 `upMbps/downMbps`（受 `SessionTracker.access` 保护）；`applyUserRateLocked` 只在目标值与当前值不同时替换该方向，未变化方向保留原 `*limiterBox` 与 token。
- `limiterBox.changed` 在替换时关闭；`rateLimiter.WaitN` 用 `ReserveN` + timer/context/changed 三路 select，改速时取消预留并返回 `errRateChanged`。
- `waitN` 通过 loader 每次取当前桶，收到 `errRateChanged` 后**重试同一块**，不重复扣费；ctx 已取消时优先返回 `ctx.Err()`。
- 单元测试：
  - `TestSetUserRatesOnlyChangesModifiedDirection`：保存 B 时 A 的两个指针不变；只改下载时上传指针不变；上传改 0 时下载指针不变。
  - `TestSavingAnotherUserDoesNotRefillBurst`：先耗尽 A 的 burst，再保存 B，A 的 token 不会被补满（容差 1024 B）。
  - `TestRateLimiterInterruptsQueuedWait`：排队 `WaitN` 在 changed 关闭后 500ms 内返回 `errRateChanged`。
- E2E `TestRateLimitE2EQueuedUnlimitedLatency`：4 条 TCP 下载连接按 1 Mbps 排队，测得排队速率 **1.049 Mbps**；调用 `SetUserRates(carol, 0)` 后，**100 ms** 采样窗内达到 >20 Mbps（采样窗上限 100ms，实际唤醒在第一个窗口内）。

**复验命令**：

```bash
go test ./core -run 'TestSetUserRates|TestSavingAnotherUser|TestRateLimiterInterruptsQueuedWait' -v -count=1
go test ./core -tags ratee2e -run TestRateLimitE2EQueuedUnlimitedLatency -v -count=1 -timeout 3m
```

**仍未验证 / 说明**：
- 改速本身会为新桶补满 burst（这是切换策略的预期行为）；“不额外获得 burst”只针对**未变化**的用户/方向。
- TCP 部分写仍按整块预扣令牌，短时可能轻微多扣；60 秒平均不受影响。

### 13.3 补强端到端测试

**改造**：
- `TestRateLimitE2EThroughput` 改为两个阶段：
  - 上传阶段：两个入站同时各跑 **TCP 上传 + UDP 上传**，共 4 条流；
  - 下载阶段：两个入站同时各跑 **TCP 下载 + UDP 下载**，共 4 条流；
  - 每阶段预热 5 秒、统计 60 秒；上传按 sink 接收字节、下载按客户端接收字节。
- `udpSource` 支持同一 source 向多个入站回发。
- `maxConns=1` 改为回显服务：第一条发 `first` 并读回 `first`；第二条必须被拒绝（`socksDial` 失败或读 EOF/重置），**读超时明确判失败**；关闭第一条后第三条可以再次回显。半关闭仍由“读到 EOF 后再写 done”的 server 验证。

**最终实测**：

| 阶段 | 结果 |
|---|---|
| 上传（两入站 TCP+UDP，60s） | **2.000 Mbps**（TCP 14,221,095 B + UDP 780,000 B） |
| 下载（两入站 TCP+UDP，60s） | **7.999 Mbps**（TCP 58,915,965 B + UDP 1,080,000 B） |
| 两阶段 tracker 用户流量 | **86,206,071 B** |
| `maxConns=1` | 第一条回显成功，第二条被拒绝（非超时），释放后第三条成功 |
| 半关闭 | 收到 `done` |

**复验命令**：

```bash
go test ./core -tags ratee2e -run TestRateLimitE2EThroughput -v -count=1 -timeout 6m
go test ./core -tags ratee2e -run TestRateLimitE2EMaxConnsAndHalfClose -v -count=1 -timeout 3m
```

**仍未验证**：Hysteria2、TUIC 真实协议吞吐；TUN/WireGuard 无 `metadata.User`；下载阶段 TCP 仍占多数，UDP 份额由共享桶竞争决定（UDP-only 测试另行证明 UDP 可打满）。

### 13.4 32 位换算溢出

**问题**：`mbps * 125000` 使用 `int`，在 32 位目标上 17180 Mbps 会溢出为负数。

**修复**：
- 新增 `util/rateconv/rateconv.go`：
  - `BytesPerMbps=125000`；
  - `MaxMbps=(2^31-1)/125000=17179`；
  - `BytesPerSecond` 用 int64 计算；17180 返回 `ErrTooLarge`。
- `newRateLimiter` 改用 `rateconv.BytesPerSecond`，签名改为 `(*limiterBox, error)`；`rate.Limit` 用 int64 结果。
- `service/client.go` 的 `validateClientRates` 在新增/单编辑/批量编辑路径拒绝 >17179，返回 `upMbps must not exceed 17179` 或同样的下行错误；`Save` 路径测试覆盖。
- `GetRateLimitsMap` 对数据库中的 >17179 返回明确错误，避免带病加载。
- 前端两个输入 `:max="17179"`，6 种语言提示补充最大 17179 Mbps。

**32 位复验（实际 386 目标，非模拟）**：
- 交叉编译 `linux/386` 的 `core` 和 `util/rateconv` 测试二进制；
- 用 QEMU system 模式 + Alpine x86 内核/用户态启动 32 位 Linux；
- 运行结果：

```text
=== 32-bit rateconv boundary ===
--- PASS: TestBytesPerSecondBoundary
--- PASS: Test32BitTarget
=== 32-bit core rate tests ===
--- PASS: TestNewRateLimiterBoundary
--- PASS: TestRateLimiterInterruptsQueuedWait
--- PASS: TestSetUserRatesOnlyChangesModifiedDirection
--- PASS: TestSavingAnotherUserDoesNotRefillBurst
=== all 32-bit tests done ===
```

- 17179 Mbps → **2,147,375,000 B/s**；17180 Mbps → **明确范围错误**（`ErrTooLarge`）。
- 本机 `GOARCH=386` 测试二进制为 `ELF 32-bit LSB executable, Intel 80386`，`strconv.IntSize=32`。

**复验命令**：见第 7.4 节“32 位目标复验”。

**仍未验证**：没有在 386 下运行完整 60 秒 E2E；32 位只覆盖换算边界、限速更新/唤醒/握手相关单元测试。

### 13.5 整体复验结论

```bash
go test ./... -count=1
go test -race ./core -tags ratee2e -run TestRateLimitE2ESmoke -count=10 -timeout 10m
go test ./core -tags ratee2e -run 'TestRateLimitE2E' -v -count=1 -timeout 20m
cd frontend && npm test && npm run build
```

- Go 全部包测试通过；`npm test` 133/133；`vue-tsc`、`npm run build` 通过。
- 全部 E2E 通过：冒烟（0/0、2/8、握手失败）、两阶段 2.000/7.999、双用户 8.021/8.021、在线改速 7.995/1.966/62,604.882、UDP-only 2.001/8.000、队列解除 100ms、maxConns 回显、半关闭。
- 仍未验证的项目汇总：Hysteria2/TUIC 真实客户端吞吐；TUN/WireGuard 用户识别；不限制上行 60 秒对照；386 全链路 E2E；真实上游 TLS/协议握手失败的错误类型；前端批量编辑的速率输入 UI（后端缺字段保留已测）。
