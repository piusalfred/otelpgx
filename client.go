package otelpgx

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
)

const (
	DBMaxConnLifetimeKey          = "db.max_conn_lifetime"
	DBMaxConnIdleTimeKey          = "db.max_conn_idle_time"
	DBMaxConnsKey                 = "db.max_conns"
	DBMinConnsKey                 = "db.min_conns"
	DBHealthCheckPeriodKey        = "db.health_check_period"
	DBStatementCacheCapacityKey   = "db.statement_cache_capacity"
	DBDescriptionCacheCapacityKey = "db.description_cache_capacity"
	DBHostKey                     = "db.host"
	DBPortKey                     = "db.port"
	DBUserKey                     = "db.user"
	DBConnectTimeoutKey           = "db.connect_timeout"
	DBKerberosSrvNameKey          = "db.kerberos_srv_name"
	DBKerberosSpnKey              = "db.kerberos_spn"
)

type ResourceConfig struct {
	ServiceName    string
	ServiceVersion string
	ServiceEnv     string
}

// parsePgxConfig parses the pgxpool.Config and returns attributes for the resource.
func parsePgxConfig(config *pgxpool.Config) []attribute.KeyValue {
	cc := config.ConnConfig

	attrs := []attribute.KeyValue{
		semconv.DBSystemPostgreSQL,
		semconv.DBName(cc.Database),
		semconv.DBConnectionString(cc.ConnString()),
		attribute.String(DBMaxConnLifetimeKey, config.MaxConnLifetime.String()),
		attribute.String(DBMaxConnIdleTimeKey, config.MaxConnIdleTime.String()),
		attribute.Int64(DBMaxConnsKey, int64(config.MaxConns)),
		attribute.Int64(DBMinConnsKey, int64(config.MinConns)),
		attribute.String(DBHealthCheckPeriodKey, config.HealthCheckPeriod.String()),
		attribute.Int64(DBStatementCacheCapacityKey, int64(cc.StatementCacheCapacity)),
		attribute.Int64(DBDescriptionCacheCapacityKey, int64(cc.DescriptionCacheCapacity)),
		attribute.String(DBHostKey, cc.Host),
		attribute.Int64(DBPortKey, int64(cc.Port)),
		attribute.String(DBUserKey, cc.User),
		attribute.String(DBConnectTimeoutKey, cc.ConnectTimeout.String()),
		attribute.String(DBKerberosSrvNameKey, cc.KerberosSrvName),
		attribute.String(DBKerberosSpnKey, cc.KerberosSpn),
	}

	for k, v := range cc.RuntimeParams {
		keyValue := fmt.Sprintf("db.runtime_param.%s", strings.ToLower(k))
		attrs = append(attrs, attribute.String(keyValue, v))
	}

	return attrs
}

func createOTelResource(config *ResourceConfig, pc *pgxpool.Config, attrs ...attribute.KeyValue) (*resource.Resource, error) {
	initialAttrs := []attribute.KeyValue{
		semconv.ServiceName(config.ServiceName),
		semconv.ServiceVersion(config.ServiceVersion),
		semconv.DeploymentEnvironment(config.ServiceEnv),
	}

	finalAttrs := append(initialAttrs, parsePgxConfig(pc)...)

	finalAttrs = append(initialAttrs, attrs...)

	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			finalAttrs...,
		),
	)
	if err != nil {
		return nil, err
	}

	return r, nil
}
