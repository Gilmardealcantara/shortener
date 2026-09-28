package tel

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/Gilmardealcantara/shortener/pkg/config"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

// Setup initializes the OpenTelemetry providers and configures slog.
// The returned function shuts down all providers created by Setup.
func Setup(ctx context.Context, cfg *config.Config) (func(context.Context) error, error) {
	if !cfg.EnebleOTel {
		slog.InfoContext(ctx, "OpenTelemetry disabled")
		return func(context.Context) error { return nil }, nil
	}

	var shutdownFuncs []func(context.Context) error

	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	var err error
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	res, err := newResource()
	if err != nil {
		return shutdown, err
	}

	headers := map[string]string{"api-key": cfg.NewRelicAPIKey}
	otel.SetTextMapPropagator(newPropagator())

	tracerProvider, err := newTracerProvider(ctx, res, cfg.OTelEndpoint, headers)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	meterProvider, err := newMeterProvider(ctx, res, cfg.OTelEndpoint, headers)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	loggerProvider, err := newLoggerProvider(ctx, res, cfg.OTelEndpoint, headers)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)
	slog.SetDefault(otelslog.NewLogger("github.com/Gilmardealcantara/shortener"))

	ProbeConnectivity(ctx, cfg.OTelEndpoint)

	return shutdown, nil
}

// HTTPHandler adds server-side tracing to an HTTP handler.
func HTTPHandler(handler http.Handler) http.Handler {
	return otelhttp.NewHandler(handler, "/")
}

// ProbeConnectivity creates a span and forces it to be exported immediately.
func ProbeConnectivity(ctx context.Context, endpoint string) {
	tracer := otel.Tracer("startup-probe")
	spanCtx, span := tracer.Start(ctx, "otel.connectivity.probe")
	span.End()

	type forceFlush interface {
		ForceFlush(context.Context) error
	}
	if ff, ok := otel.GetTracerProvider().(forceFlush); ok {
		if err := ff.ForceFlush(ctx); err != nil {
			slog.ErrorContext(spanCtx, "OTel connectivity probe failed — traces may not reach the backend",
				"endpoint", endpoint,
				"error", err,
			)
			return
		}
	}
	slog.InfoContext(spanCtx, "OTel connectivity probe OK", "endpoint", endpoint)
}

func newResource() (*resource.Resource, error) {
	return resource.Merge(
		resource.Default(),
		resource.NewSchemaless(
			attribute.String("service.name", "shortener"),
			attribute.String("service.version", "0.1.0"),
		),
	)
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTracerProvider(ctx context.Context, res *resource.Resource, endpoint string, headers map[string]string) (*trace.TracerProvider, error) {
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithHeaders(headers),
	)
	if err != nil {
		return nil, err
	}

	return trace.NewTracerProvider(
		trace.WithBatcher(exporter, trace.WithBatchTimeout(time.Second)),
		trace.WithResource(res),
	), nil
}

func newMeterProvider(ctx context.Context, res *resource.Resource, endpoint string, headers map[string]string) (*metric.MeterProvider, error) {
	exporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(endpoint),
		otlpmetrichttp.WithHeaders(headers),
		otlpmetrichttp.WithTemporalitySelector(metric.DefaultTemporalitySelector),
	)
	if err != nil {
		return nil, err
	}

	return metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter, metric.WithInterval(10*time.Second))),
		metric.WithResource(res),
	), nil
}

func newLoggerProvider(ctx context.Context, res *resource.Resource, endpoint string, headers map[string]string) (*log.LoggerProvider, error) {
	stdoutExporter, err := stdoutlog.New(stdoutlog.WithPrettyPrint(), stdoutlog.WithoutTimestamps())
	if err != nil {
		return nil, err
	}

	otlpExporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpoint(endpoint),
		otlploghttp.WithHeaders(headers),
	)
	if err != nil {
		return nil, err
	}

	return log.NewLoggerProvider(
		log.WithProcessor(log.NewSimpleProcessor(stdoutExporter)),
		log.WithProcessor(log.NewBatchProcessor(otlpExporter)),
		log.WithResource(res),
	), nil
}
