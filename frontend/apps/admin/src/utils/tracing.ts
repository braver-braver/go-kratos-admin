/**
 * Minimal trace header injector for frontend → backend propagation.
 * - Generates W3C `traceparent` if missing.
 * - Adds `X-Request-ID` for log correlation.
 * - Optional static baggage via env.
 *
 * Designed to be backward compatible: headers are additive and safe
 * even if the backend has not yet enabled tracing.
 */

type HeaderRecord = Record<string, string>;

const TRACE_ENABLED =
  (import.meta.env.VITE_TRACE_ENABLED ?? 'true').toString() !== 'false';
const STATIC_BAGGAGE = (import.meta.env.VITE_TRACE_BAGGAGE ?? '')
  .toString()
  .trim();

let sessionTraceId: null | string = null;

function randomHex(bytes: number) {
  const buf = new Uint8Array(bytes);
  if (typeof crypto !== 'undefined' && crypto.getRandomValues) {
    crypto.getRandomValues(buf);
  } else {
    for (let i = 0; i < bytes; i++) {
      buf[i] = Math.floor(Math.random() * 256);
    }
  }
  return [...buf].map((b) => b.toString(16).padStart(2, '0')).join('');
}

function generateTraceId() {
  return randomHex(16); // 16 bytes => 32 hex chars
}

function generateSpanId() {
  return randomHex(8); // 8 bytes => 16 hex chars
}

function ensureSessionTraceId() {
  if (!sessionTraceId) {
    sessionTraceId = generateTraceId();
  }
  return sessionTraceId;
}

function ensureRequestId(headers: HeaderRecord) {
  if (headers['X-Request-ID']) return;
  headers['X-Request-ID'] =
    typeof crypto !== 'undefined' && 'randomUUID' in crypto
      ? crypto.randomUUID()
      : generateTraceId();
}

function ensureTraceparent(headers: HeaderRecord) {
  if (headers.traceparent) return;
  const traceId = ensureSessionTraceId();
  const spanId = generateSpanId();
  headers.traceparent = `00-${traceId}-${spanId}-01`;
}

function ensureBaggage(headers: HeaderRecord) {
  if (headers.baggage || !STATIC_BAGGAGE) return;
  headers.baggage = STATIC_BAGGAGE;
}

export function applyTracingHeaders(config: { headers?: HeaderRecord }) {
  if (!TRACE_ENABLED) return config;
  const headers: HeaderRecord = (config.headers ||= {});
  ensureTraceparent(headers);
  ensureBaggage(headers);
  ensureRequestId(headers);
  return config;
}
