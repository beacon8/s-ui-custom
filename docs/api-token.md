# API Token 与按节点批量添加客户端

本文记录外部程序使用 **API Token** 调用 s-ui 管理接口，以及用
`importClientNodes` 按节点批量创建客户端的方式。文中命令均已在隔离实例中
执行过；示例凭据为 `example-only` 一类占位值。

## 1. 取得 Token 和 API 基础地址

1. 以管理员登录面板，打开「设置」→「API 令牌」（「管理员」页的同一入口也
   可）。两个入口管理同一批 Token，升级前签发的有效 Token 继续可用。
2. 「API 基础地址」会显示当前部署的完整地址，包含实际面板路径，例如：

   ```text
   http://127.0.0.1:22095/panel/apiv2
   ```

   点击右侧复制图标即可复制。面板路径不是 `/app/` 时，这里也会显示真实路径。
3. 点击「添加」，填写说明和有效期（天；0 表示长期），创建后复制一次性显示
   的 Token。列表只显示 `****`；撤销后下一次请求立即失效。
4. 外部请求只带 Token 头，不需要 Cookie：

   ```bash
   export SUI_API='http://127.0.0.1:22095/panel/apiv2'
   export SUI_TOKEN='<创建时复制的 Token>'
   curl -sS "$SUI_API/status" -H "Token: $SUI_TOKEN"
   ```

Token 的身份就是它所属的管理员。跨管理员的 Token 列表、撤销和批次查看会被
拒绝；`changePass` 使用 Token 所属管理员并校验原密码，审计记录实际操作者。
浏览器登录/退出仍是会话用途，Token 不能调用 `login`/`logout`。

### 状态码与凭据失败

| 场景 | HTTP | 响应 |
|---|---:|---|
| 有效 Token | 200 | 正常业务响应 |
| 缺失、错误、过期、已撤销 | 401 | `{"success":false,"msg":"invalid token","obj":null}` |

## 2. 查询可用入站

`inboundIds` 必须是已存在、且能按客户端身份匹配的入站。混合、SOCKS、HTTP、
Shadowsocks（非 managed）、VMess、VLESS、Trojan、Hysteria/2、TUIC、AnyTLS、
Naive、ShadowTLS v3、Snell 等类型会出现在 `users` 里。TUN 等没有用户身份的
入站会被拒绝。

```bash
curl -sS "$SUI_API/inbounds" -H "Token: $SUI_TOKEN" \
  | jq '.obj.inbounds[] | {id, tag, type, enable, users}'
```

示例把 `12` 当作测试入站 ID，实际执行时替换成上一步返回的 ID。

## 3. `POST /importClientNodes`

请求体为 JSON。单个和批量是同一个接口，`nodes` 数组长度决定数量，结果按输入
顺序返回。

| 字段 | 说明 |
|---|---|
| `prefix` | 客户端名前缀，字母/数字/`_`/`.`/`-`，最长 32。 |
| `startId` | 起始编号，默认 1。结束编号 = 起始 ID + 有效节点数 − 1。 |
| `idWidth` | 补零宽度，最少 3，最多 12；数字超过宽度时自然增宽。 |
| `inboundIds` | 本批共用的已有入站 ID，至少一个。 |
| `clientDefaults` | 本批共用设置，见下表。 |
| `defaultProtocol` | 没有协议头的文本行使用的协议，`socks5`（默认）或 `http`。 |
| `nodes` | 每项用结构化字段、`{"raw":"…"}` 或 `{"uri":"…"}` 表示，可混合。 |
| `dryRun` | `true` 时只预览和校验，零写入、无核心变更。 |

### `clientDefaults`

沿用单个客户端表单的字段、单位和校验：

| 字段 | 说明 |
|---|---|
| `enable` | 是否启用，默认 `true`。 |
| `desc`、`remark` | 描述、备注。 |
| `volume` / `volumeGiB` | 每人流量额度，二选一。`volume` 为字节，`volumeGiB` 为表单的 GiB。 |
| `expiry` | Unix 秒；0 表示无限。 |
| `delayStart`、`autoReset`、`resetDays` | 延迟启动、自动重置及周期。`delayStart && !autoReset` 时 `expiry` 按表单规则写 0；需要重置周期时缺省为 1 天。 |
| `maxConns` | 连接数限制，0 不限。 |
| `ipLimit` | 去重 IP 数限制，0 不限；IPv6 按全局前缀归并。 |
| `upMbps`、`downMbps` | 上下行限速，0 不限，最大 17179。 |
| `config` | 可共用的协议选项（如 `{"vless":{"flow":"xtls-rprx-vision"}}`）。`name`、`username`、`password`、`uuid`、`auth_str`、`userkey` 属于每个客户端独立的身份字段，在这里设置会被拒绝。 |
| `links` | 每个成员都收到的非本地链接：`{"type":"external"|"sub","remark":"…","uri":"…"}`。 |

### 节点表示与解析

- 结构化：`{"protocol":"socks5","host":"192.0.2.10","port":1080,"username":"u","password":"p"}`。
  `protocol` 支持 `socks5`、`http`、`https`；HTTPS 会启用 TLS。可选 `tls` 对象原样
  传给出站。密码中的 `:`、`@` 原样保留。
- 原始行：`地址:端口:用户名:密码`，密码可含 `:`；IPv6 用
  `[2001:db8::1]:1080:user:pass`。空行忽略。可按输入行号返回错误。
- 分享链接：`socks5://`、`http://`、`https://`、`vless://`、`vmess://` 自动识别，
  复用 Go 解析器；TLS、Reality、WS/gRPC/HTTP 等传输参数随解析结果保留。
- 裸地址不会按端口猜协议；没有 `://` 的行使用 `defaultProtocol`。
- 每个节点生成一个客户端、一个出站 `out-<客户端名>` 和一条
  `auth_user=[客户端名] → 该出站` 路由。客户端凭据、UUID 和本地链接分别生成。

## 4. 编号、分组与幂等

| prefix | startId | 数量 | 客户端名 | 出站标签 | 分组 |
|---|---:|---:|---|---|---|
| US | 1 | 1 | US-001 | out-US-001 | US-1-1 |
| US | 1 | 10 | US-001…US-010 | out-US-001…out-US-010 | US-1-10 |
| US | 2 | 10 | US-002…US-011 | out-US-002…out-US-011 | US-2-11 |

- 任意客户端名或出站标签与现有数据冲突、任意一行格式错误，整批停止；返回全部
  可预判问题（稳定错误码、条目位置/原始行号、冲突名），不自动跳号或覆盖。
- 正式提交在事务内再次校验；客户端、入站绑定、出站、路由、批次和幂等记录
  同一事务提交。预览、校验失败或回滚不会触碰运行中的核心。
- `Idempotency-Key` 相同、调用方相同、内容相同：返回原批次结果（`replayed:true`）；
  同键换内容返回 409 `IDEMPOTENCY_CONFLICT`。预览不占用键。不同键争抢同一编号
  范围时只有一个批次成功，另一个返回 422 冲突。
- 返回 `batchId`、`group`、`count`、`core.status` 和按输入顺序的 `items`
  （`clientId`、`clientName`、`outboundTag`、`protocol`、`address`）。

## 5. 可直接运行的示例

以下示例使用 `SUI_API` 和 `SUI_TOKEN` 环境变量，入站 ID 示例为 `12`。

### 5.1 单个 SOCKS5 节点

```bash
curl -sS "$SUI_API/importClientNodes" \
  -H "Token: $SUI_TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: demo-us-single-001' \
  --data '{
    "prefix": "US",
    "startId": 1,
    "inboundIds": [12],
    "clientDefaults": {"enable": true, "upMbps": 10, "downMbps": 50},
    "nodes": [
      {"protocol":"socks5","host":"192.0.2.10","port":1080,
       "username":"demo-user","password":"example-only"}
    ]
  }'
```

实测返回（凭据已脱敏，ID 取决于测试库）：

```json
{
  "success": true,
  "msg": "",
  "obj": {
    "batchId": "imp_edb6a83a42f5447b843a5cee",
    "batchStatus": "applied",
    "prefix": "US",
    "group": "US-1-1",
    "count": 1,
    "startId": 1,
    "endId": 1,
    "idWidth": 3,
    "inboundIds": [12],
    "inboundTags": ["mixed-test"],
    "items": [{
      "index": 0, "line": 1, "clientId": 1,
      "clientName": "US-001", "outboundTag": "out-US-001",
      "protocol": "socks5", "address": "192.0.2.10:1080"
    }],
    "core": {"status": "applied", "at": 1795675927},
    "dryRun": false, "valid": true
  }
}
```

### 5.2 批量 10 个 HTTP 节点

需要 `jq`：

```bash
jq -n '{
  prefix: "US",
  startId: 2,
  idWidth: 3,
  inboundIds: [12],
  defaultProtocol: "http",
  clientDefaults: {enable: true, maxConns: 8, ipLimit: 2},
  nodes: [range(1; 11) | {raw: "192.0.2.\(.):8080:demo-user:example-only"}]
}' | curl -sS "$SUI_API/importClientNodes" \
  -H "Token: $SUI_TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: demo-us-batch-002' \
  --data-binary @-
```

实测：`group=US-2-11`、`count=10`，名称为 `US-002` 至 `US-011`，
`outboundTag` 对应 `out-US-002` 至 `out-US-011`，协议均为 `http`。

### 5.3 混合协议与特殊字符

```bash
UUID=b831381d-6324-4d53-ad4f-8cda48b30811
VMESS=$(printf '%s' '{"v":"2","ps":"vmess-node","add":"example.org","port":"8443","id":"'$UUID'","aid":"0","net":"ws","path":"/ws","tls":"tls","sni":"example.org"}' | base64 | tr -d '\n')
curl -sS "$SUI_API/importClientNodes" \
  -H "Token: $SUI_TOKEN" -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: demo-mixed-001' \
  --data '{
    "prefix": "MX", "startId": 1, "inboundIds": [12],
    "nodes": [
      {"raw": "192.0.2.50:1080:demo-user:p@ss:word", "line": 1},
      {"raw": "192.0.2.51:8080:demo-user:p@ss:word", "line": 2},
      {"protocol": "https", "host": "192.0.2.52", "port": 443,
       "username": "demo-user", "password": "p@ss:word",
       "tls": {"enabled": true, "server_name": "example.com"}, "line": 3},
      {"uri": "https://demo-user:p%40ss%3Aword@192.0.2.53:443#https-node", "line": 4},
      {"uri": "vless://'$UUID'@example.com:443?security=tls&sni=example.com&type=ws&path=%2Fws#vless-node", "line": 5},
      {"uri": "vmess://'$VMESS'", "line": 6}
    ]
  }'
```

实测 6 项协议依次为 `socks5`、`socks5`、`https`、`https`、`vless`、`vmess`，
密码 `p@ss:word` 原样保存到出站；VLESS/VMess 的 TLS 与 WS 传输保留。裸行没有
协议头，因此仍按 `defaultProtocol=socks5` 解析，不按 8080 端口猜协议。

### 5.4 预览

同一请求加 `"dryRun": true`。不会写库、不会创建批次、不会改核心：

```json
{
  "success": true,
  "msg": "",
  "obj": {
    "batchId": "",
    "prefix": "US", "group": "US-1-1", "count": 1, "startId": 1, "endId": 1,
    "idWidth": 3, "inboundIds": [12], "inboundTags": ["mixed-test"],
    "items": [{"index": 0, "line": 1, "clientId": 0, "clientName": "US-001",
               "outboundTag": "out-US-001", "protocol": "socks5",
               "address": "192.0.2.10:1080"}],
    "errors": [], "core": {"status": "not_applied"},
    "dryRun": true, "valid": true
  }
}
```

预览的键不会被正式创建占用。预览里不出现密码；`items` 只含名称、标签、协议和
节点地址。

### 5.5 重名整批停止

在中间安排一个已有客户端名（或出站标签、非法端口）：

```bash
curl -sS -o /tmp/conflict.json -w '%{http_code}\n' "$SUI_API/importClientNodes" \
  -H "Token: $SUI_TOKEN" -H 'Content-Type: application/json' \
  --data '{"prefix":"US","startId":2,"inboundIds":[12],
           "nodes":[{"raw":"192.0.2.60:1080:u:p","line":1},
                    {"raw":"192.0.2.61:1080:u:p","line":2}]}'
```

实测 HTTP 422：

```json
{
  "success": false,
  "msg": "the batch was not created; every listed issue must be fixed first",
  "obj": {
    "code": "VALIDATION_FAILED",
    "errors": [
      {"code": "CLIENT_NAME_CONFLICT", "line": 1, "conflict": "US-002",
       "message": "client name is already in use"},
      {"code": "OUTBOUND_TAG_CONFLICT", "line": 1, "conflict": "out-US-002",
       "message": "outbound tag is already in use"}
    ]
  }
}
```

同一响应会列出所有可预判问题。实测操作前后数据库新增 0 个客户端、0 个出站、
0 条路由，运行中核心无变化。

### 5.6 同键重试与丢失响应

第一次请求已提交但客户端没收到响应时，用**同一个键和同一份内容**重试：

```bash
curl -sS "$SUI_API/importClientNodes" \
  -H "Token: $SUI_TOKEN" -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: demo-us-single-001' \
  --data '…与第一次完全相同…'
```

实测返回原 `batchId`，并带 `"replayed": true`，不会创建第二批。同键换内容：

```json
{"success": false, "msg": "this Idempotency-Key was already used with a different request body",
 "obj": {"code": "IDEMPOTENCY_CONFLICT", "batchId": "imp_…"}}
```

HTTP 409。不同键同时争抢同一编号范围时，实测一个 200、一个 422，数据库只增加
一批客户端。

### 5.7 核心应用失败与恢复

数据库已提交、核心重启失败时返回 HTTP 500，但**客户端和批次已保存**：

```json
{
  "success": false,
  "msg": "the batch is committed but the core rejected it; use reapplyImportBatch, do not create it again",
  "obj": {
    "code": "CORE_APPLY_FAILED",
    "batchId": "imp_4f2690ec239058332d518db6",
    "core": {"status": "failed",
             "error": "start inbound: … bind: address already in use"},
    "retry": "POST apiv2/reapplyImportBatch {\"batchId\":\"imp_4f2690ec239058332d518db6\"}"
  }
}
```

修复环境后只重试应用，不重新创建：

```bash
curl -sS "$SUI_API/reapplyImportBatch" \
  -H "Token: $SUI_TOKEN" -H 'Content-Type: application/json' \
  --data '{"batchId":"imp_4f2690ec239058332d518db6"}'
```

实测恢复后 `core.status=applied`，客户端数仍为 1，未重复创建。也可先用
`GET "$SUI_API/importBatch?batchId=…"` 定位批次；该视图只返回脱敏后的请求摘要，
不含节点密码、订阅链接或自定义链接 URI。

## 6. 错误码与重试方式

| code | HTTP | 含义 | 处理 |
|---|---:|---|---|
| `BAD_REQUEST` | 400 | JSON 非法、未知字段、键过长 | 改请求 |
| `VALIDATION_FAILED` | 422 | 前缀/编号/入站/节点/冲突等预检失败 | 按 `errors[].line`/`conflict` 改输入 |
| `IDEMPOTENCY_CONFLICT` | 409 | 同键不同内容 | 换键或改回原内容 |
| `CORE_APPLY_FAILED` | 500 | 数据已提交，核心应用失败 | 用 `reapplyImportBatch`，不要重建 |
| `BATCH_STATUS_UNSAVED` | 500 | 数据已提交，批次状态未写回 | 用 `reapplyImportBatch` |
| `BATCH_NOT_FOUND` | 404 | 批次不存在 | 检查 `batchId` |
| `BATCH_FORBIDDEN` | 403 | 批次属于其他管理员 | 用所属管理员的 Token |
| `BATCH_DATA_MISSING` | 409 | 批次数据已被改动/删除，无法安全重放 | 人工检查，不能重建原批次 |
| `DB_BUSY` | 503 | 数据库忙 | 用同键重试原请求 |
| `UNAUTHORIZED` | 401 | 没有有效 Token/会话 | 重新签发 Token |

## 7. 功能覆盖表

面板操作到外部 API 的逐项对照、HTTP 方法、参数、权限和测试结果见
[api-coverage.md](api-coverage.md)。Token 路由遗漏由
`api/route_coverage_test.go` 机械比对 v1/v2 处理器，面板新增动作而 Token
路由漏加时测试会失败。

## 8. 分组订阅与个人订阅链接

导入成功后，组和个人各自拥有**独立的订阅 ID 和 Token**。订阅 ID 与批次号、
客户端名、代理协议凭据都没有关系；同名分组或同前缀的新批次会得到新的组订阅
身份，不会继承旧链接的权限。

```text
<订阅基础地址>/g/<组订阅ID>?token=<组订阅Token>
<订阅基础地址>/c/<个人订阅ID>?token=<个人订阅Token>
```

订阅基础地址取现有订阅设置：优先 `subURI`，否则由协议、`subDomain`、
`subPort`、`subPath` 组合。链接为完整可复制的 HTTP(S) URL，Token 放在查询串，
兼容不能设置请求头的订阅客户端。

### 8.1 导入响应新增字段

`POST /importClientNodes` 的 `obj` 新增：

```json
{
  "groupSubscription": {
    "subscriptionId": "5f19d4929d0b512dffa06c67ce84a0d8",
    "name": "S10-1-10",
    "link": "http://host:2096/sub/g/5f19d4929d0b512dffa06c67ce84a0d8?token=…"
  },
  "items": [
    {
      "clientId": 4,
      "clientName": "S10-004",
      "subscriptionId": "8dd94f15e172e9735b7198104bd43827",
      "subscriptionLink": "http://host:2096/sub/c/8dd94f15e172e9735b7198104bd43827?token=…"
    }
  ]
}
```

链接只在授权的创建、同键重放和重放应用响应中出现。`GET /importBatch` 只返回
订阅 ID，不返回 Token/链接。预览（`dryRun`）不签发任何 Token。

### 8.2 订阅行为

- 整组链接每次请求返回该组**当前**可订阅成员：成员被删除、禁用或按现有规则
  失效时自动不再返回；成员的入站变化会在下次刷新生效。
- 个人链接只返回该客户自己的内容。组内其他人、其他组的 ID 或 Token 交叉
  组合均返回 404。
- 默认格式由现有设置统一 base64 编码一次；`format=json`、`format=clash`、
  `format=happ` 和 User-Agent 自动识别与旧订阅一致。组 JSON/Clash 生成一份
  完整配置，出站标签跨成员去重。
- 旧式 `/<用户名>` 对已有订阅成员的客户返回 400（正文和 HEAD 都不泄露）。
  未迁移的老客户继续按旧方式使用；调用下面任一接口即可显式迁移：

```bash
curl -sS -X POST "$SUI_API/enableClientSubscription" \
  -H "Token: $SUI_TOKEN" --data "clientId=4"
```

### 8.3 管理接口

| 方法 | 路径 | 参数 | 说明 |
|---|---|---|---|
| GET | `/subscriptionGroups` | — | 组列表；不含 Token/链接 |
| GET | `/subscriptionGroup` | `groupId` | 组详情、组链接、成员个人链接 |
| POST | `/enableClientSubscription` | `clientId` | 旧客户显式启用；幂等 |
| POST | `/rotateSubscription` | `kind=group|member`、`id` | 轮换组/个人 Token，返回新链接 |
| POST | `/revokeSubscription` | `kind`、`id`、`revoke=true|false` | 撤销/恢复 |
| POST | `/renameSubscriptionGroup` | `groupId`、`name` | 只改显示名，链接不变 |
| POST | `/removeSubscriptionMember` | `memberId` | 从组移除；个人链接随之失效 |

轮换个人 Token 只替换该人的链接，组链接和其他 9 人不受影响。要让某人从整组
消失，应移除成员或禁用客户，而不是只轮换个人 Token。组 Token 与个人 Token
可分别撤销；撤销组不影响个人链接。已分发的代理登录凭据（节点密码/UUID）有
自己的生命周期，轮换订阅 Token 不会改变它们。

### 8.4 存储与备份

组/个人 Token 使用面板 `secret` 派生的 AES-GCM 密钥加密保存（数据库里是
密文，不是明文，也不是不可恢复的哈希），授权复制时解密。`secret` 本身在
settings 表中，随数据库备份/恢复；恢复后原链接继续有效。普通列表、历史批次
结果、审计和访问日志不写 Token；订阅响应带 `Cache-Control: no-store`。
