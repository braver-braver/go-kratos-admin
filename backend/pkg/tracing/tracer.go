package tracing

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	conf "github.com/tx7do/kratos-bootstrap/api/gen/go/conf/v1"

	"kratos-admin/pkg/service"
)

// Config is a minimal tracing configuration with safe defaults.
type Config struct {
	Enabled     bool
	ServiceName string
	Environment string
	Endpoint    string
	Insecure    bool
	SampleRate  float64
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envFloat(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

// LoadConfig builds tracing config from bootstrap plus environment fallbacks.
// Unknown/missing fields keep tracing optional and backward compatible.
func LoadConfig(cfg *conf.Bootstrap) Config {
	c := Config{
		Enabled:     true,
		ServiceName: envOrDefault("TRACE_SERVICE_NAME", service.AdminService),
		Environment: envOrDefault("TRACE_ENVIRONMENT", ""),
		Endpoint:    envOrDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"),
		Insecure:    envOrDefault("OTEL_EXPORTER_OTLP_INSECURE", "true") != "false",
		SampleRate:  envFloat("TRACE_SAMPLE_RATE", 0.1),
	}

	if cfg != nil && cfg.Server != nil && cfg.Server.Rest != nil && cfg.Server.Rest.Middleware != nil {
		c.Enabled = cfg.Server.Rest.Middleware.GetEnableTracing()
	}

	return c
}

// EnvTracingEnabled exposes a simple env-based switch for components
// that want to avoid hard dependency on config structure.
func EnvTracingEnabled() bool {
	return envOrDefault("TRACE_ENABLED", "true") != "false"
}

// InitProvider installs the global tracer provider and propagator.
// Returns a shutdown function; caller should invoke it during graceful stop.
func InitProvider(ctx context.Context, cfg *conf.Bootstrap, logger log.Logger) (func(context.Context) error, error) {
	c := LoadConfig(cfg)
	if !c.Enabled || !EnvTracingEnabled() {
		return func(context.Context) error { return nil }, nil
	}

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpointURL(c.Endpoint),
	}
	if c.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	exp, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(c.ServiceName),
			attribute.String("deployment.environment", c.Environment),
		),
		resource.WithProcess(),
		resource.WithHost(),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp, sdktrace.WithBatchTimeout(5*time.Second)),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(c.SampleRate)),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	helper := log.NewHelper(log.With(logger, "module", "tracing"))
	helper.Infof(
		"tracing enabled: endpoint=%s sample_rate=%.2f service=%s env=%s", c.Endpoint, c.SampleRate,
		c.ServiceName, c.Environment,
	)

	return tp.Shutdown, nil
}
