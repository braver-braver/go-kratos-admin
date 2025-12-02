# Frontend Tracing & Request Correlation

## What was added
- Automatic injection of `traceparent`, `baggage` (optional), and `X-Request-ID` headers for all API calls in `apps/admin`.
- Default is **enabled**; headers are additive and safe even if the backend has not yet enabled tracing.
- A stable per-page `traceId` is generated once; each request gets its own `spanId`.

## Configuration
- `VITE_TRACE_ENABLED` (default `true`): set to `false` to disable header injection.
- `VITE_TRACE_BAGGAGE` (optional): comma-separated baggage entries, e.g. `user.id=123,tenant.id=abc` (keep values non-sensitive).

## Behavior
- If `traceparent` already exists on the request, the interceptor leaves it untouched.
- `X-Request-ID` is generated via `crypto.randomUUID()` when absent (falls back to random hex).
- `baggage` is only set when `VITE_TRACE_BAGGAGE` is provided.
- Works with current backend as-is; when backend tracing is enabled (Jaeger/OTel), these headers allow end-to-end correlation.

## Header contract (frontend ↔ backend)
- Send: `traceparent` + optional `baggage` (W3C), plus `X-Request-ID` (UUID v4 recommended).
- Receive: backend always returns `Trace-Id` (authoritative trace ID for Jaeger lookup), full `traceparent` (`00-<trace>-<span>-01`), and echoes `X-Request-ID`.
- Use `Trace-Id`/`traceparent` for Jaeger searches; keep `X-Request-ID` for log correlation (it may differ from trace ID when provided by the client).

## Usage
No code changes are required in feature modules. To disable per environment:
```sh
VITE_TRACE_ENABLED=false pnpm dev
```
To add static baggage:
```sh
VITE_TRACE_BAGGAGE="user.id=demo,tenant.id=demo-tenant" pnpm dev
```

## CORS note
If the backend enforces an allowlist of headers, ensure it includes:
`traceparent`, `baggage`, `X-Request-ID` (see `backend/app/admin/service/configs/server.yaml: server.rest.cors.headers`).
