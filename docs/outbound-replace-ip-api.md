# 出站管理「替换 IP」API

日期：2026-11-28
实现：`service/outbounds.go`、`api/apiService.go`、
`frontend/src/layouts/modals/OutboundReplaceIp.vue`

## 1. 用途

管理员在出站管理表格中按任意顺序勾选多个 SOCKS/HTTP 出站，粘贴等量的
`服务器地址:端口:用户名:密码` 记录，一次提交整批替换。每个出站保留原
`id`、`tag`、`type` 及所有其他配置（SOCKS 版本、HTTP TLS、拨号设置等），只更新
`server`、`server_port`、`username`、`password`。

- 会话：`POST /api/replaceOutbounds`
- API Token：`POST /apiv2/replaceOutbounds`，请求头 `Token: <token>`
- Content-Type：`application/json`

## 2. 请求

```json
{
  "items": [
    {
      "tag": "socks-c",
      "server": "192.0.2.66",
      "server_port": 9873,
      "username": "user-c",
      "password": "pass:c"
    }
  ]
}
```

| 字段 | 必填 | 说明 |
|---|---|---|
| `items` | 是 | 至少一项；顺序即勾选顺序，前端用它对应粘贴行。 |
| `items[].tag` | 是 | 目标出站标签，必须存在且不能重复。 |
| `items[].server` | 是 | 新服务器地址；IPv6 传不带方括号的地址。 |
| `items[].server_port` | 是 | 1–65535 的整数。 |
| `items[].username` | 否 | 新用户名；空字符串表示删除认证用户名。 |
| `items[].password` | 否 | 新密码；空字符串表示删除认证密码。密码可包含冒号。 |

请求体上限 1 MiB，必须是单个 JSON 对象，未知字段会被拒绝。

## 3. 响应

```json
{
  "success": true,
  "msg": "",
  "obj": {
    "saved": 1,
    "applied": true,
    "outbounds": [
      {
        "id": 3,
        "tag": "socks-c",
        "type": "socks",
        "server": "192.0.2.66",
        "server_port": 9873,
        "username": "user-c",
        "password": "pass:c",
        "version": "5"
      }
    ]
  }
}
```

| 字段 | 说明 |
|---|---|
| `saved` | 本次提交成功的出站数。 |
| `applied` | 是否已成功重启核心并生效。 |
| `warning` | 数据库已提交但运行时生效失败时的说明；此时 `applied=false`。 |
| `outbounds` | 替换后重新读取的完整出站列表，供前端立即刷新。 |

## 4. 事务与生效

1. 后端先校验全部目标：非空、无重复、存在、类型为 SOCKS/HTTP、端口合法。
2. 同一事务内逐项读取原出站，更新 `Options` 的四个字段；其他字段原样保留。
3. 任一项校验或写入失败，整批回滚，不产生部分更新。
4. 提交成功后只调用一次核心重启流程；启动失败时数据库仍已保存，接口返回
   `success:true`，`applied:false` 并在 `warning` 中说明。前端必须展示“已保存
   但未生效”，不能当成完全失败。
5. 路由和用户仍引用原 tag，不重建出站、不改 tag/type/id。

## 5. 错误

`success:false` 且 `msg` 带目标标签或原因，例如：

- `outbound "x": record not found`
- `outbound "x" appears more than once`
- `outbound "x": type "vless" is not supported`
- `outbound "x": port must be 1-65535`
- `invalid JSON body: ...`

前端保留已输入的文本和目标映射，错误显示在弹窗内；可修正后重试。
