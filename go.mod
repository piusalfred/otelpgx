module github.com/exaring/otelpgx

go 1.22

toolchain go1.22.3

require (
	github.com/jackc/pgx/v5 v5.6.0
	go.opentelemetry.io/contrib/bridges/otelslog v0.2.0
	go.opentelemetry.io/otel v1.27.0
	go.opentelemetry.io/otel/exporters/stdout/stdoutlog v0.3.0
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.27.0
	go.opentelemetry.io/otel/metric v1.27.0
	go.opentelemetry.io/otel/sdk v1.27.0
	go.opentelemetry.io/otel/sdk/log v0.3.0
	go.opentelemetry.io/otel/sdk/metric v1.27.0
	go.opentelemetry.io/otel/trace v1.27.0
)

require (
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20231201235250-de7065d80cb9 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	go.opentelemetry.io/otel/log v0.3.0 // indirect
	golang.org/x/crypto v0.23.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	golang.org/x/text v0.15.0 // indirect
)
