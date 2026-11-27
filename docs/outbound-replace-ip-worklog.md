# 出站管理「替换 IP」工作记录

日期：2026-11-28
任务书：`../s-ui-outbound-replace-ip-prompt.md`

## 1. 开工前阅读

- `AGENTS.md`：代码变更后必须增量更新 `.ua` 知识图谱、回写工作记录、构建并发版按
  `../s-ui-custom/AGENTS.md`。
- `.ua/knowledge-graph.json` 出站保存链：
  - `OutboundService.GetAll` 把 `Options` 展平供前端；
  - `OutboundService.Save` 在事务中处理 new/edit/del，edit 先按 id 取旧 tag，
    运行中先 `RemoveOutbound(oldTag)` 再 `AddOutbound(newConfig)`，最后写库；
  - `ConfigService.Save` 负责事务提交、审计 `changes`、刷新 `LoadPartialData`；
  - `api.Save` 成功后返回受影响对象，前端 `Data().save` 用 `setNewData` 刷新。
- 历史工作记录：`docs/session-worklog-2026-11-28.md`、`docs/egress-ip-worklog.md`、
  `docs/api-coverage.md`。
- 前端现有页面：`frontend/src/views/Outbounds.vue`（搜索、分页、单条编辑/删除、
  批量新增）、`frontend/src/layouts/modals/OutboundBulk.vue`（现有批量新增，仅
  创建新出站，不可复用为原地替换）、`frontend/src/views/Clients.vue`（多选、
  跨页保留、页脚、工具栏布局）。

## 2. 可复用实现

1. `OutboundService.Save` 的 `Options` 读写方式和事务边界；替换只需更新四个字段，
   保留 id/tag/type 与其余 `Options`。
2. `ConfigService.Save` 的事务提交后动作模式，以及 `RestartCore`/`restartCoreWithConfig`
   的一次性生效路径；批量替换应在提交后只重启一次核心。
3. 前端 `Data().save`/`setNewData` 的刷新方式；批量替换接口可直接返回
   `success/msg/obj`，成功后调用 `Data().loadData()` 或 `setNewData`。
4. `Outbounds.vue` 的搜索、分页、`matchesOutboundSearch`、`clampPage`、
   `paginationConfig`，以及 `Clients.vue` 的多选、跨页 `preserveSelectedRowKeys`、
   页脚/工具栏样式。
5. 现有协议组件 `Socks.vue`/`Http.vue`：用户名/密码为空时删除字段，非空时保留；
   后端替换需遵循同样的可空语义。
6. 现有 i18n 结构（en/zhcn/zhtw 完整，ru/vi/fa 回退 en）。

## 3. 尚待验证/风险

- 批量事务中任一目标不存在、重复、类型不支持或字段非法时必须全部回滚。
- 提交成功但运行时 `RestartCore` 失败：数据库已更新，界面必须明确“已保存但未
  生效”，不能显示成完全失败。
- 打开弹窗后目标被删除，或核心/数据库写入失败时，输入和标签映射不能丢失。
- 选中顺序：跨页、搜索、排序、取消后重选、当前页全选都必须保持追加语义。
- 输入解析：IPv4/域名/方括号 IPv6、密码含冒号、空用户名/密码、CRLF、空白行、
  非法行不能导致后续行错配，端口 1–65535。
- 现有单条编辑、批量新增、核心热更新不能回归。
- 运行时真实核心替换（SOCKS/HTTP）和事务回滚需要隔离实例验证；单测只能覆盖
  事务和字段保留，不能完全替代真实拨号。

## 4. 实施计划

1. 后端新增 `POST /{api,apiv2}/replaceOutbounds`，严格解析目标 tag + 四个字段；
   事务内校验/更新 `Options`，提交后重启核心一次，返回 `saved/applied/warning`。
2. 前端新增 `OutboundReplaceIp.vue`：有序多选、输入解析、逐行预览、行数/字段
   校验、一次提交、失败保留。
3. `Outbounds.vue` 增加仅 SOCKS/HTTP 可用的勾选列与「替换 IP（N）」工具栏按钮，
   保持跨页/搜索/排序顺序。
4. 补后端事务/字段保留/回滚测试、前端解析测试；隔离实例做任务书四组验收。
5. 交付 API 文档、覆盖表、逐项实测结果；更新 AGENTS/工作记录与知识图谱；按
   发布规则推新版本。

## 5. 实现摘要

- 后端：`OutboundService.ReplaceServers` 对选中的 SOCKS/HTTP 出站在一个事务内
  校验并更新 `Options` 的 `server`、`server_port`、`username`、`password`；空
  用户名/密码删除对应键，其他字段原样保留。提交成功后只重启核心一次；核心不可用
  或重启失败时返回 `saved:true, applied:false, warning`。
- API：`POST /{api,apiv2}/replaceOutbounds`，严格 JSON、1 MiB 上限；响应带
  `saved/applied/warning/outbounds`。
- 前端：`Outbounds.vue` 增加仅 SOCKS/HTTP 可勾选的多选列与「替换 IP（N）」；
  选中顺序独立维护，跨页/搜索/全选按追加规则；`OutboundReplaceIp.vue` 实时解析、
  逐行预览（含完整明文用户名/密码）、行数与字段校验、一次提交、失败保留。
- 解析：`utils/outboundReplace.ts`，支持 LF/CRLF/空白行、IPv4、域名、方括号
  IPv6、密码含冒号、空认证；非法行保留位置。
- 文档：[outbound-replace-ip-api.md](outbound-replace-ip-api.md)。

## 6. 逐项实测

| 场景 | 操作 | 实际结果 |
|---|---|---|
| 顺序与协议 | A/C 为 SOCKS、B 为 HTTP，按 C→B→A 勾选，粘贴 66/61/68 | ✅ 预览 C→66、B→61、A→68；保存值一致；A/C `version` 4/5、B `tls.enabled/server_name` 完整保留 |
| 单选 | 只勾选 A | ✅ 仅 A 出现在预览，确认按钮为 1 |
| 当前页全选 | 14 个出站，页 1 含 A/B/C/D/E01–E06 | ✅ 选中 9 个，vless D 禁用且未选中；预览按表格顺序 B,C,A,E01…E06 |
| 跨页 | 页 1 选 A，页 2 选 E-10，回页 1 | ✅ 按钮 `(2)`，预览 A→E-10，顺序保持 |
| 搜索 | 已选 A/E-10，搜索 E-10 | ✅ 已选不清空，按钮仍 `(2)`，预览仍 A→E-10 |
| 取消 | 打开后取消 | ✅ 原记录不变，已选顺序保留 |
| 空白行/CRLF | 三行中夹空行、CRLF | ✅ 空白跳过，三条正确映射 |
| IPv6/密码含冒号 | `[2001:db8::10]:1080:alice:p:a` | ✅ 新地址 `[2001:db8::10]:1080`，用户名 alice，密码 `p:a` |
| 行数不等 | 选 3 个、输入 2 条 | ✅ 显示「已选 3 个出站，已输入 2 条」，确认禁用 |
| 非法行 | 中间一条 `bad-line` | ✅ 该行保留位置并标红，后续行不错配，确认禁用 |
| 端口越界/0 | `host:70000`、`host:0` | ✅ 前端标红；后端端口 0/重复/不支持类型测试拒绝 |
| 目标被删除 | 打开弹窗后删除 A，再提交 | ✅ 后端返回 `outbound "A-socks": record not found`；弹窗保留输入/映射；B/C 原值未变 |
| 事务回滚 | 批内一个不存在目标 | ✅ service 测试：原出站字段不变 |
| 运行时失败 | 开启维护模式后提交 | ✅ `success:true, saved:1, applied:false, warning:"...core is stopped for maintenance"`；配置已保存 |
| 单条编辑回归 | 用原值保存 B | ✅ `api/save` 成功 |
| 批量新增入口回归 | 打开批量添加弹窗 | ✅ 弹窗正常打开，未修改其逻辑 |

## 7. 自动化与验证命令

- `go test ./... -count=1`：全部通过。
- 新增 `TestReplaceServersPreservesOtherOptions`、
  `TestReplaceServersRollsBackOnMissingTarget`、
  `TestReplaceServersRejectsDuplicateUnsupportedAndEmptyAuth`、
  `TestReplaceOutboundsAPI`。
- 前端 `npm test`：13 文件 / **165** 测试通过；`vue-tsc --noEmit`、`npm run build`
  通过。
- 隔离实例（22095）：核心运行中替换返回 `applied:true`；维护模式返回
  `applied:false + warning`。

## 8. 未验证/边界

- 没有注入真实 SQLite 写失败（如磁盘满/约束冲突）；原子性由“不存在目标”与
  事务代码覆盖，未做故障注入。
- 出站表格当前没有可排序列，因此“排序后选择”未实测；搜索和跨页已实测。
- 未在 386 运行本功能；未对真实 SOCKS/HTTP 流量做替换后的拨号回归。
- 前端选择顺序由浏览器实测，未做组件级自动化测试。
