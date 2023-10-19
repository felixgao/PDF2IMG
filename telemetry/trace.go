package telemetry

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"
)

// ----------------------------------------
// Tracer Setup and Teardown
// ----------------------------------------
func newTraceProvider() {
	// The context passed in to the exporter is only passed to the client and used when connecting to the endpoint
	ctx := context.Background()

	client, err := getTraceClient()
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to initialize OLTP trace client")
		return
	}

	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to initialize OLTP trace exporter")
		return
	}

	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter)),
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(newResource()),
		),
	)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)
}

func getTraceClient() (client otlptrace.Client, err error) {
	protocol := otlpProtocolGrpc
	if v := os.Getenv(otlpProtocol); v != "" {
		protocol = v
	}
	if v := os.Getenv(otlpTracesProtocol); v != "" {
		protocol = v
	}

	endpoint := "0.0.0.0:4317"
	if v := os.Getenv(otlpEndpoint); v != "" {
		endpoint = v
	}

	switch protocol {
	case otlpProtocolHTTP:
		// TODO: figure out how to do secure options
		client = otlptracehttp.NewClient(otlptracehttp.WithEndpoint(endpoint))
	case otlpProtocolGrpc:
		secureOption := otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
		if v := os.Getenv("INSECURE_MODE"); v != "" {
			secureOption = otlptracegrpc.WithInsecure()
		}
		client = otlptracegrpc.NewClient(secureOption,
			otlptracegrpc.WithEndpoint(endpoint))
	default:
		err = fmt.Errorf("unknown or unsupported OLTP protocol: %s. No traces will be exported", protocol)
	}
	return
}

func cleanupTraceProvider() error {
	tracer, ok := otel.GetTracerProvider().(shutdownTracerProvider)
	if ok {
		return tracer.Shutdown(context.Background())
	}
	return nil
}

type shutdownTracerProvider interface {
	oteltrace.TracerProvider
	Shutdown(ctx context.Context) error
}
