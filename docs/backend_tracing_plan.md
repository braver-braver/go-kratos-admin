# Backend Tracing Plan (Jaeger + OpenTelemetry)

## Objectives
- End-to-end request visibility across REST/gRPC, GORM, Redis, Asynq, SSE pushes, and outbound HTTP clients.
- Correlate traces with existing structured logs (reuse `X-Request-ID` when present, always emit `trace_id`/`span_id`).
- Minimal code churn: rely on Kratos middleware + OTel auto-instrumentation helpers; keep configs driven by `bootstrap` settings.

## Header contract (Trace vs Request ID)
- Inbound: accept W3C `traceparent`/`tracestate`; accept optional `X-Request-ID` (UUID or any string) for log correlation.
- Server behavior: if `X-Request-ID` is missing, reuse the current span trace ID (or UUID) for that header; never override a client-supplied value.
- Outbound response: always return `Trace-Id` (active span trace ID) and full `traceparent` (`00-<trace>-<span>-01`) so Jaeger can search directly; echo `X-Request-ID` for log correlation. `Trace-Id` is the authoritative value for Jaeger lookup.
- Downstream calls: tracing middleware injects `traceparent`; clients should forward `X-Request-ID` unchanged for log joins.

## Target Architecture
- **Tracer provider**: OpenTelemetry SDK with Jaeger exporter (collector endpoint preferred; OTLP/HTTP fallback).
- **Propagation**: W3C TraceContext + Baggage; continue honoring `X-Request-ID` by mapping it to span attributes/log fields.
- **Sampling**: Head-based ratio (env/configurable), default `0.1` in dev, `1.0` in staging when debugging, `0.01`+ in prod.
- **Resource attributes**: `service.name=kratosAdmin-admin-service`, `service.version`, `deployment.environment`, `host.name`.
- **Log correlation**: logging middleware injects `trace_id`, `span_id`, `request_id` into log context; Asynq jobs copy parent span.
- **Request ID echo**: Server middleware reuses incoming `X-Request-ID` if present; otherwise falls back to current `trace_id`, and always returns `X-Request-ID` in the response header for frontend display.

```
Client -> REST/gRPC -> Kratos middleware (tracing + logging) -> Service -> Repo (GORM/Redis) -> Asynq/HTTP out
             |                                                                       ^
             v                                                                       |
           Jaeger <--------------- OTLP (trace spans, db/redis attrs) ---------------+
```

## Integration Plan
1) **Tracer bootstrap (shared pkg `backend/pkg/tracing`, OTLP-first)**
   - Initialize once during bootstrap:
     ```go
     shutdown, _ := tracing.InitProvider(ctx, cfg, logger)
     ```
   - Config and env (both optional, env wins when present):
     ```yaml
     trace:
       serviceName: kratosAdmin-admin-service
       otlpEndpoint: http://localhost:4318   # Jaeger all-in-one OTLP/HTTP
       sampleRate: 0.1
       environment: dev
     ```
     Env fallbacks: `TRACE_ENABLED`, `TRACE_SERVICE_NAME`, `TRACE_SAMPLE_RATE`, `TRACE_ENVIRONMENT`,
     `OTEL_EXPORTER_OTLP_ENDPOINT` (host:port), `OTEL_EXPORTER_OTLP_INSECURE` (default true).

2) **Server middleware**
   - REST/HTTP: add `tracing.Server()` before auth/logging middleware in `newRestMiddleware`.
   - gRPC (if added later): same middleware stack.
   - Inject trace correlation into logs:
     ```go
     if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
         logger = log.With(logger,
             "trace_id", sc.TraceID().String(),
             "span_id", sc.SpanID().String(),
         )
     }
     ```
   - Map `X-Request-ID` -> span attribute `request.id` and add to log context.

3) **Clients**
   - gRPC/HTTP clients: enable `tracing.Client()` so outbound calls carry context.
   - Asynq producers/consumers: keep `context.Context` when enqueueing; wrap handlers with `otel.Tracer.Start` to restore parent span (leveraging existing “carry ctx” comment).
   - SSE push: include `trace_id` in event metadata when originating from traced handlers.

4) **Data layer instrumentation**
   - GORM: `gorm.Use(otelgorm.NewPlugin(...))` behind config/env flag.
   - Redis: `redisotel.InstrumentTracing()` after client creation (env/flag guarded).
   - SQL drivers (if using raw DB): optionally `otelsql` wrapper.

5) **Export/Deployment**
   - Local dev: `docker run -p 16686:16686 -p 14268:14268 -p 4317:4317 -p 4318:4318 jaegertracing/all-in-one` (use OTLP port 4318 HTTP).
   - Prod: point OTLP/HTTP to central collector (Jaeger/Tempo/OTel Collector); enable TLS/auth if provided by infra.

## How to view a single trace in Jaeger
1) Open the UI: `http://localhost:16686`.
2) Select Service: `kratosAdmin-admin-gateway` (or the service name you configured).
3) (Optional) Enter the `X-Request-ID`/`trace_id` from response headers into “Trace ID” or “Tags” search (e.g., `request.id=<id>`).
4) Click “Find Traces”, then open the latest trace to see spans (HTTP -> GORM -> Redis -> Asynq).

6) **Verification**
   - Run `go test ./...` with `TRACE_SAMPLE_RATE=1` to ensure spans created without network errors (use stdout exporter in CI).
   - Manual check: call `/admin/v1/login`; confirm spans in Jaeger show HTTP -> GORM -> Redis -> Asynq.

## Sequence (PlantUML)
```plantuml
@startuml
actor Client
participant REST as REST Server
participant Service as AuthService
participant Repo as GORM Repo
participant Redis as Redis
participant Asynq as AsynqWorker

Client -> REST: HTTP request (traceparent + X-Request-ID)
REST -> REST: tracing.Server() creates span\nlog middleware adds trace_id/span_id
REST -> Service: ctx carries trace/span
Service -> Repo: DB call (otelgorm span)
Repo --> Service: rows
Service -> Redis: redisotel span (token cache)
Service -> Asynq: enqueue with ctx\n(parent span context)
Asynq -> Asynq: handler starts child span\nlogs include trace_id/span_id
Service --> REST: response
REST --> Client: response headers with trace info (optional)
@enduml
```

## Frontend Propagation (Vben/Vue)
- **Traceparent injection**: browser client adds W3C `traceparent` + `baggage` headers per request. If no parent, frontend generates one via `@opentelemetry/api` WebTracer (or lightweight header generator) so backend receives a valid context.
- **Request ID**: keep sending `X-Request-ID` (UUID v4) from frontend; backend logs map it to span attributes `request.id`.
- **CORS/Headers**: ensure `traceparent`, `baggage`, `X-Request-ID` are whitelisted in CORS allow-headers (REST server config).
- **SPA navigation**: create a root span per page view (optional) and set it as parent for all API calls triggered on that page; attach user/tenant info as `baggage` keys (e.g., `user.id`, `tenant.id`) if safe.
- **Error surfacing**: when backend returns a trace-aware header (optional `X-Trace-ID`), show it in UI error to help support correlate logs/traces.

## Work Items Checklist
- [ ] Add tracer initializer + config block.
- [ ] Wire `tracing.Server()` / `tracing.Client()` into server/client creation.
- [ ] Inject `trace_id`/`span_id` into logging middleware (REST + Asynq).
- [ ] Enable `otelgorm` and `redisotel`.
- [ ] Provide dev `docker-compose` snippet for Jaeger and update README/docs.

## Local Jaeger with OrbStack
OrbStack ships a lightweight Docker runtime; Jaeger “all-in-one” works out of the box.
1) Start Jaeger:
   ```sh
   docker run --rm -p 16686:16686 -p 14268:14268 jaegertracing/all-in-one:1.57
   ```
   - UI: http://localhost:16686
   - Collector HTTP (OTLP HTTP compatible): http://localhost:14268/api/traces
2) Point the service to the collector in `bootstrap.yaml`:
   ```yaml
   trace:
     serviceName: kratosAdmin-admin-service
     jaegerEndpoint: http://host.docker.internal:14268/api/traces
     sampleRate: 1.0   # for local debugging
     environment: dev
   ```
   Use `host.docker.internal` so the containerized app can reach the host’s Jaeger; if running the app on the host, `http://localhost:14268/api/traces` is fine.
3) Run the service (env):
   ```sh
   TRACE_SAMPLE_RATE=1 go run ./app/admin/service/cmd/server
   ```
4) Exercise an endpoint (e.g., login) and open the Jaeger UI to verify spans (should show HTTP entry + GORM + Redis + Asynq).
