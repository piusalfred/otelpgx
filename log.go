package otelpgx

import (
	"log/slog"

	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

type LoggerProviderConfig struct {
	ScopeName      string
	ScopeVersion   string
	ScopeSchemaURL string
}

type LoggerProvider struct {
	Provider *log.LoggerProvider
	Logger   *slog.Logger
	Handler  slog.Handler
}

func NewLoggerProvider(cfg LoggerProviderConfig, exporter log.Exporter) (*LoggerProvider, error) {
	logExporter, err := stdoutlog.New()
	if err != nil {
		return nil, err
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
		log.WithAttributeCountLimit(8),
		log.WithResource(&resource.Resource{}),
	)

	//instrumentationScope := instrumentation.Scope{
	//	Name:      cfg.ScopeName,
	//	Version:   cfg.ScopeVersion,
	//	SchemaURL: cfg.ScopeSchemaURL,
	//}

	opts := []otelslog.Option{
		otelslog.WithLoggerProvider(loggerProvider),
	}

	logger := otelslog.NewLogger(
		"logger", opts...,
	)

	handler := otelslog.NewHandler("logger", opts...)

	return &LoggerProvider{
		Provider: loggerProvider,
		Logger:   logger,
		Handler:  handler,
	}, nil
}
