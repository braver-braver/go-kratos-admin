# 前端链路追踪与请求关联

## 新增内容
- 在 `apps/admin` 的所有 API 请求自动注入 `traceparent`、可选的 `baggage`、以及 `X-Request-ID`。
- 默认开启；即便后端尚未启用追踪也不会影响现有行为，属于附加型头部。
- 每个页面会生成稳定的 `traceId`，每个请求生成独立的 `spanId`。

## 配置
- `VITE_TRACE_ENABLED`（默认 `true`）：设为 `false` 可关闭头部注入。
- `VITE_TRACE_BAGGAGE`（可选）：逗号分隔的 baggage 项，例如 `user.id=123,tenant.id=abc`（避免敏感信息）。

## 行为说明
- 若请求已带 `traceparent`，拦截器不会覆盖。
- `X-Request-ID` 在缺省时使用 `crypto.randomUUID()` 生成（降级为随机 hex）。
- 仅当提供 `VITE_TRACE_BAGGAGE` 时才写入 `baggage`。
- 当前后端未启用追踪也能正常工作；启用 Jaeger/OTel 后可实现端到端关联。

## 头部约定（前后端）
- 发送：`traceparent` + 可选 `baggage`（W3C），再加 `X-Request-ID`（建议 UUID v4）。
- 接收：后端总是返回 `Trace-Id`（用于 Jaeger 查找的权威 TraceID）、完整 `traceparent`（`00-<trace>-<span>-01`），并回显 `X-Request-ID`。
- 在 Jaeger 搜索请使用 `Trace-Id`/`traceparent`；`X-Request-ID` 用于日志关联（客户端自带时可能与 TraceID 不同）。

## 使用示例
无需修改业务代码。按需在环境变量中控制：
```sh
# 关闭追踪头
VITE_TRACE_ENABLED=false pnpm dev

# 添加静态 baggage
VITE_TRACE_BAGGAGE="user.id=demo,tenant.id=demo-tenant" pnpm dev
```

## CORS 提示
如果后端对请求头做白名单校验，请包含：
`traceparent`、`baggage`、`X-Request-ID`（参考 `backend/app/admin/service/configs/server.yaml: server.rest.cors.headers`）。
