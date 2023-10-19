package telemetry

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"google.golang.org/grpc/credentials"
)

var meterProvider *sdkmetric.MeterProvider

func newMeterProvider() {
	// The context passed in to the exporter is only passed to the client and used when connecting to the endpoint
	ctx := context.Background()

	exporter, err := getMetricsClient(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to initialize OLTP metric exporter")
		return
	}

	reader := sdkmetric.NewPeriodicReader(exporter)

	meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(newResource()),
		sdkmetric.WithReader(reader),
	)

	otel.SetMeterProvider(meterProvider)
}

func getMetricsClient(ctx context.Context) (client sdkmetric.Exporter, err error) {
	protocol := otlpProtocolGrpc
	if v := os.Getenv(otlpProtocol); v != "" {
		protocol = v
	}
	if v := os.Getenv(otlpMetricsProtocol); v != "" {
		protocol = v
	}
	switch protocol {
	case otlpProtocolHTTP:
		client, err = otlpmetrichttp.New(ctx)
	case otlpProtocolGrpc:
		secureOption := otlpmetricgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
		if v := os.Getenv("INSECURE_MODE"); v != "" {
			secureOption = otlpmetricgrpc.WithInsecure()
		}
		client, err = otlpmetricgrpc.New(ctx, secureOption)
	default:
		err = fmt.Errorf("unknown or unsupported OLTP protocol: %s. No metrics will be exported", protocol)
	}
	return
}

func cleanupMeterProvider() (err error) {
	if meterProvider != nil {
		ctx := context.Background()
		err = meterProvider.ForceFlush(ctx)
		if err != nil {
			err = meterProvider.Shutdown(ctx)
		}
	}
	return err
}
