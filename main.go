package main

import (
	"context"
	"log"
	"os"
	"runtime"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"google.golang.org/grpc/credentials"

	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	apis "github.com/felixgao/pdf_to_png/api"
	"github.com/felixgao/pdf_to_png/telemetry"
)

var (
	serviceName = os.Getenv("SERVICE_NAME")
	endpoint    = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	insecure    = os.Getenv("INSECURE_MODE")
)

func initTracer() func(context.Context) error {

	secureOption := otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if len(insecure) > 0 {
		secureOption = otlptracegrpc.WithInsecure()
	}

	if endpoint == "" {
		endpoint = "0.0.0.0:4317"
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(endpoint),
		),
	)

	if err != nil {
		log.Fatal("Could not set exporter: ", err)
	}
	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("host.arch", runtime.GOARCH),
			attribute.String("application", "PDF2IMG-app"),
		),
	)
	if err != nil {
		log.Fatal("Could not set resources: ", err)
	}
	// set up the global trace provider
	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter)),
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(resources),
		),
	)
	// set up the global metrics provider

	return exporter.Shutdown
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func setupWebServer() {
	r := gin.Default()
	// gin OpenTelemetry middleware
	r.Use(otelgin.Middleware("otel-otlp-go-service"))
	r.Use(gzip.Gzip(gzip.BestSpeed))
	r.Use(corsMiddleware())
	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())

	// setup end points
	apis.RegisterHealthCheckHandlers(r)
	apis.RegisterConvertHandlers(r)

	// start the server
	_ = r.Run(":8080")

}

func main() {
	// initializing the tracer and metric
	// cleanup := initTracer()
	// defer cleanup(context.Background())
	telemetry.SetupFromEnvs()
	defer telemetry.Cleanup()

	r := gin.Default()
	r.Use(otelgin.Middleware("otel-otlp-go-service"))
	log.SetOutput(os.Stdout)
	log.SetFlags(0)

	// setup for vips
	vips.LoggingSettings(nil, vips.LogLevelError)
	vips.Startup(nil)
	defer vips.Shutdown()

	// setup web server
	setupWebServer()
}
