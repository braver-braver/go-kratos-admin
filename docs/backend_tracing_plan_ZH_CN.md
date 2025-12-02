# 后端链路追踪方案（Jaeger + OpenTelemetry）

## 目标
- 全链路可观测：覆盖 REST/gRPC、GORM、Redis、Asynq、SSE 推送、外部 HTTP 调用。
- 日志与 Trace 关联：沿用 `X-Request-ID`，统一输出 `trace_id`/`span_id` 字段。
- 低侵入改造：优先使用 Kratos tracing 中间件与 OTel 插件，配置驱动。

## 头部约定（Trace vs Request ID）
- 入站：接受 W3C `traceparent`/`tracestate`；接受可选的 `X-Request-ID`（UUID 或任意字符串）用于日志关联。
- 服务端：若缺少 `X-Request-ID`，用当前 span 的 TraceID（或 UUID）填充，不覆盖客户端已有值。
- 出站响应：始终返回 `Trace-Id`（当前 span TraceID）和完整 `traceparent`（`00-<trace>-<span>-01`），便于在 Jaeger 直接查找；同时回显 `X-Request-ID` 做日志关联。`Trace-Id` 是 Jaeger 查询的权威值。
- 下游调用：Tracing 中间件自动注入 `traceparent`；`X-Request-ID` 原样透传以便日志串联。

## 目标架构
- **Tracer Provider**：OpenTelemetry SDK + Jaeger Exporter（Collector 优先，支持 OTLP/HTTP）。
- **传播协议**：W3C TraceContext + Baggage，`X-Request-ID` 写入 span attribute 和日志字段。
- **采样**：Head-based ratio，可配置；本地 0.1，调试 1.0，生产 0.01+。
- **资源属性**：`service.name=kratosAdmin-admin-service`、`service.version`、`deployment.environment`、`host.name`。
- **日志关联**：中间件从 ctx 提取 trace/span，写入 log context；Asynq 任务保留父 span。
- **请求 ID 回传**：如果前端提供 `X-Request-ID` 则复用；否则使用当前 `trace_id` 兜底，并在响应头回传 `X-Request-ID`，便于前端显示/报错对齐。

## 集成步骤
1) **Tracer 启动（公共包 `backend/pkg/tracing`，优先 OTLP）**
   ```go
   shutdown, _ := tracing.InitProvider(ctx, cfg, logger)
   ```
   配置/环境变量（可选，环境变量优先）：
   ```yaml
   trace:
     serviceName: kratosAdmin-admin-service
     otlpEndpoint: http://localhost:4318   # Jaeger all-in-one OTLP/HTTP
     sampleRate: 0.1
     environment: dev
   ```
   Env 兜底：`TRACE_ENABLED`、`TRACE_SERVICE_NAME`、`TRACE_SAMPLE_RATE`、`TRACE_ENVIRONMENT`、
   `OTEL_EXPORTER_OTLP_ENDPOINT`（host:port）、`OTEL_EXPORTER_OTLP_INSECURE`（默认 true）。

2) **服务端中间件**
   - REST/HTTP：在 `newRestMiddleware` 中加入 `tracing.Server()`，置于鉴权/日志前。
   - （未来 gRPC 同理）客户端调用使用 `tracing.Client()` 传播上下文。
   - 日志关联：
     ```go
     if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
         logger = log.With(logger,
             "trace_id", sc.TraceID().String(),
             "span_id", sc.SpanID().String(),
         )
     }
     ```
   - 将 `X-Request-ID` 写入 span attribute `request.id`，并落入日志。

3) **客户端与异步任务**
   - gRPC/HTTP 客户端启用 tracing middleware，自动注入 traceparent。
   - Asynq 入队传递 ctx；消费者 handler 内 `Tracer.Start` 创建子 span，日志同样带 trace_id/span_id。
   - SSE 推送可在事件 payload/metadata 附带 `trace_id`（便于前端调试）。

4) **数据层埋点**
   - GORM：在 `gormcli.NewGormClient` 使用 `otelgorm.NewPlugin(...)`（配置/环境开关）。
   - Redis：`redisotel.InstrumentTracing()`（配置/环境开关）。
   - 若有原生 SQL 驱动可选 `otelsql` 包装。

5) **部署与导出**
   - 开发：`docker run -p 16686:16686 -p 14268:14268 -p 4317:4317 -p 4318:4318 jaegertracing/all-in-one`（使用 OTLP 4318 HTTP）。
   - 生产：指向集中 OTLP Collector（Jaeger/Tempo/OTel Collector），按需开启 TLS/认证。

## 在 Jaeger 中查看单个请求的完整 Trace
1) 打开 UI：`http://localhost:16686`。
2) 选择 Service：`kratosAdmin-admin-gateway`（或你的 serviceName 配置）。
3) （可选）将响应头中的 `X-Request-ID` / `trace_id` 填入 “Trace ID” 或 “Tags” 搜索（例如 `request.id=<id>`）。
4) 点击 “Find Traces”，打开最新的 Trace，可查看 HTTP -> GORM -> Redis -> Asynq 等 span 细节。

6) **验证与自测**
   - CI/单测：`TRACE_SAMPLE_RATE=1` + stdout exporter，保证不依赖外部网络。
   - 手工：调用 `/admin/v1/login`，在 Jaeger 中应看到 HTTP -> GORM -> Redis -> Asynq 的 span 链。

## 时序图
```plantuml
@startuml
actor Client
participant REST as REST Server
participant Service as AuthService
participant Repo as GORM Repo
participant Redis as Redis
participant Asynq as AsynqWorker

Client -> REST: HTTP 请求 (traceparent + X-Request-ID)
REST -> REST: tracing.Server() 创建入口 span\n日志中加入 trace_id/span_id
REST -> Service: ctx 携带 trace/span
Service -> Repo: DB 调用 (otelgorm span)
Repo --> Service: rows
Service -> Redis: redisotel span (token cache)
Service -> Asynq: 入队，携带 ctx
Asynq -> Asynq: handler Start span\n日志同样输出 trace_id/span_id
Service --> REST: 响应
REST --> Client: 响应（可回传 trace 信息）
@enduml
```

## 前端注入与传播（Vben/Vue）
- **traceparent 头**：前端在每个请求带上 W3C `traceparent`、`baggage`。若无现成 Tracer，可用 `@opentelemetry/api` WebTracer 生成，也可轻量生成符合格式的 header。
- **请求 ID**：继续发送 `X-Request-ID`（UUID v4），后端将其映射到 span attribute `request.id` 并写日志。
- **CORS**：确保 REST 允许的头里包含 `traceparent`、`baggage`、`X-Request-ID`。
- **页面级父 span**：单页应用可在每次页面进入时创建 root span，页面内所有 API 请求以其为父；可安全地把 `user.id`、`tenant.id` 等放入 baggage（避免敏感数据）。
- **错误提示**：后端可选回传 `X-Trace-ID`，前端在错误提示中展示，便于用户与运维对齐链路。

## 任务清单
- [ ] 新增 tracer 初始化与配置块。
- [ ] REST/gRPC 启用 tracing middleware，日志写入 trace_id/span_id。
- [ ] Asynq/SSE 保留并继续传播 span。
- [ ] 启用 `otelgorm` 与 `redisotel`。
- [ ] 提供本地 Jaeger 启动示例与 README 说明。

## OrbStack 本地 Jaeger 测试
OrbStack 内置 Docker 体验，直接使用 Jaeger all-in-one。
1) 启动 Jaeger：
   ```sh
   docker run --rm -p 16686:16686 -p 14268:14268 jaegertracing/all-in-one:1.57
   ```
   - UI: http://localhost:16686
   - Collector HTTP（兼容 OTLP/HTTP）: http://localhost:14268/api/traces
2) `bootstrap.yaml` 中配置：
   ```yaml
   trace:
     serviceName: kratosAdmin-admin-service
     jaegerEndpoint: http://host.docker.internal:14268/api/traces
     sampleRate: 1.0   # 本地全采样
     environment: dev
   ```
   若服务直接在宿主机跑，`http://localhost:14268/api/traces` 即可；容器内访问宿主需用 `host.docker.internal`。
3) 运行服务：
   ```sh
   TRACE_SAMPLE_RATE=1 go run ./app/admin/service/cmd/server
   ```
4) 调用接口（如 login），在 Jaeger UI 中确认链路：HTTP 入口 + GORM + Redis + Asynq。
