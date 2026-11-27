# 用户管理「出站地址」查询 API

日期：2026-11-28
相关实现：`service/outbound_address.go`、`api/apiService.go`、
`frontend/src/views/Clients.vue`

## 1. 用途

管理员打开「用户管理」时，前端按当前页客户端名批量查询每个客户端**通过路由
规则绑定到的出站**，并直接显示出站管理里配置的服务器地址。该地址来自当前
生效配置，不发起任何网络探测。

- 会话：`GET {base}/api/outboundAddresses`
- API Token：`GET {base}/apiv2/outboundAddresses`，请求头 `Token: <token>`
- 两个地址由同一个 `ApiService.GetOutboundAddresses` 处理。

> 这是“出站配置的服务器地址”，不保证等于外部服务看到的公网出口 IP。需要真实
> 公网出口时请另用探测工具，本接口刻意不拨号。

## 2. 请求

| 参数 | 位置 | 必填 | 说明 |
|---|---|---|---|
| `users` | query（或 POST 表单回退） | 否 | JSON 字符串数组，如 `["alice","bob"]`；空数组返回 `users:{}`。 |

示例：

```bash
curl -sS -b cookie.txt \
  -H 'X-Requested-With: XMLHttpRequest' \
  --get \
  --data-urlencode 'users=["WdkM3DF6"]' \
  http://127.0.0.1:2095/app/api/outboundAddresses
```

```json
{
  "success": true,
  "msg": "",
  "obj": {
    "users": {
      "WdkM3DF6": {
        "status": "ok",
        "outbound": "socks-s28",
        "address": "1.1.1.1",
        "port": 80,
        "type": "socks"
      }
    }
  }
}
```

## 3. 响应字段

| 字段 | 说明 |
|---|---|
| `status` | 见下表。 |
| `outbound` | 规则解析出的出站 tag。 |
| `address` | 出站配置的 `server` 字段；只有 `ok` 有。原样返回，可能是域名。 |
| `port` | 出站配置的 `server_port`；只有 `ok` 且非 0 时返回。 |
| `type` | 出站类型，如 `socks`、`http`、`vless`、`direct`。 |

### 状态语义

| `status` | 含义 | 前端展示 |
|---|---|---|
| `ok` | 出站有 `server` 地址。 | 直接显示地址；tooltip 显示出站 tag 与 `地址:端口`；可复制。 |
| `direct` | 出站是 `direct`，没有代理服务器。 | 「直连」。 |
| `multi` | 出站是 `selector` / `urltest`，没有单一地址。 | 「多出口」。 |
| `none` | 出站类型没有 `server` 字段（例如 `wireguard`）。 | `—`。 |
| `split` | 当前规则可能让该用户走多个出站。 | 「按规则分流」。 |
| `unbound` | 没有可用出站、命中 reject、或规则指向已删除的 tag。 | 「未绑定」。 |

`success:false` 只出现在参数错误（`users` 不是 JSON 数组）或无法读取当前配置时。

## 4. 用户 → 出站解析

后端读取当前生效配置（`ConfigService.GetConfig("")`）中的 `route.rules` 和
`route.final`：

1. `auth_user` 是客户端身份条件；`user` 是进程属主，不使用。
2. 无 `auth_user` 限制、无其它匹配条件、动作为 `route` 的规则是确定性出口；
   第一条命中后停止。
3. 带目的地址、域名、协议、入站、端口、规则集等条件的规则只记为「可能出口」；
   与默认出站不同则返回 `split`。
4. 条件规则全部落空时用 `route.final`；没有 `final` 时用第一个出站。
5. 逻辑规则 `and`/`or`/`invert`、`reject`、`auth_user` 单值/数组均已处理。
6. 链式代理（`detour`）显示**该出站自己的 `server`**；`detour` 只描述连接路径，
   不是另一个可显示地址。
7. selector/urltest 不展开成员，返回 `multi`；不猜测当前选择。

## 5. 与旧「落地 IP」接口的区别

旧的 `GET {base}/api/egressIps` 和 `apiv2/egressIps` 已移除。它通过出站向外部
回显服务拨号测量公网 IP，需要手动刷新、缓存和探测并发控制。当前接口只读配置，
无探测、无缓存、无数据库表、无后台任务；页面打开即可显示。
