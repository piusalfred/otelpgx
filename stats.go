package otelpgx

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/piusalfred/otelpgx/internal"
	"go.opentelemetry.io/otel/metric"
)

// defaultMinimumReadDBStatsInterval is the default minimum interval between calls to db.Stats().
const defaultMinimumReadDBStatsInterval = time.Second

// RecordStats records database statistics for provided pgxpool.Pool at the provided interval.
func RecordStats(db *pgxpool.Pool, opts ...StatsOption) error {
	o := statsOptions{
		//meterProvider:              global.MeterProvider(),
		//minimumReadDBStatsInterval: defaultMinimumReadDBStatsInterval,
		//defaultAttributes: []attribute.KeyValue{
		//	semconv.DBSystemPostgreSQL,
		//},
	}

	for _, opt := range opts {
		opt.applyStatsOptions(&o)
	}

	meter := o.meterProvider.Meter(internal.MeterName)

	return recordStats(meter, db, o.minimumReadDBStatsInterval, o.defaultAttributes...)

}

func recordStats(
	meter metric.Meter,
	db *pgxpool.Pool,
	minimumReadDBStatsInterval time.Duration,
	attrs ...metric.ObserveOption,
) error {
	var (
		err error

		acquireCount                         metric.Int64ObservableCounter
		acquireDuration                      metric.Float64ObservableCounter
		acquiredConns                        metric.Int64ObservableUpDownCounter
		cancelledAcquires                    metric.Int64ObservableCounter
		constructingConns                    metric.Int64ObservableUpDownCounter
		emptyAcquires                        metric.Int64ObservableCounter
		idleConns                            metric.Int64ObservableUpDownCounter
		maxConns                             metric.Int64ObservableGauge
		maxIdleDestroyCount                  metric.Int64ObservableCounter
		maxLifetimeDestroyCountifetimeClosed metric.Int64ObservableCounter
		newConnsCount                        metric.Int64ObservableCounter
		totalConns                           metric.Int64ObservableUpDownCounter

		dbStats     *pgxpool.Stat
		lastDBStats time.Time

		// lock prevents a race between batch observer and instrument registration.
		lock sync.Mutex
	)

	lock.Lock()
	defer lock.Unlock()

	if acquireCount, err = meter.Int64ObservableCounter(
		pgxPoolAcquireCount,
		metric.WithUnit(UnitDimensionless),
		metric.WithDescription("Cumulative count of successful acquires from the pool."),
	); err != nil {
		return err
	}

	if acquireDuration, err = meter.Float64ObservableCounter(
		pgxpoolAcquireDuration,
		metric.WithUnit(UnitDimensionless),
		metric.WithDescription("Total duration of all successful acquires from the pool in nanoseconds."),
	); err != nil {
		return err
	}

	if acquiredConns, err = meter.Int64ObservableUpDownCounter(
		pgxpoolAcquiredConns,
		metric.WithUnit(UnitDimensionless),
		metric.WithDescription("Number of currently acquired connections in the pool."),
	); err != nil {
		return err
	}

	if cancelledAcquires, err = meter.Int64ObservableCounter(
		pgxpoolCancelledAcquires,
		metric.WithUnit(UnitDimensionless),
		metric.WithDescription("Cumulative count of acquires from the pool that were canceled by a context."),
	); err != nil {
		return err
	}

	if constructingConns, err = meter.Int64ObservableUpDownCounter(
		pgxpoolConstructingConns,
		metric.WithUnit(UnitMilliseconds),
		metric.WithDescription("Number of conns with construction in progress in the pool."),
	); err != nil {
		return err
	}

	if emptyAcquires, err = meter.Int64ObservableCounter(
		pgxpoolEmptyAcquire,
		metric.WithUnit(UnitDimensionless),
		metric.WithDescription("Cumulative count of successful acquires from the pool that waited for a resource to be released or constructed because the pool was empty."),
	); err != nil {
		return err
	}

	if idleConns, err = meter.Int64ObservableUpDownCounter(
		pgxpoolIdleConns,
		metric.WithUnit(UnitDimensionless),
		metric.WithDescription("Number of currently idle conns in the pool."),
	); err != nil {
		return err
	}

	if maxConns, err = meter.Int64ObservableGauge(
		pgxpoolMaxConns,
		metric.WithUnit(UnitDimensionless),
		metric.WithDescription("Maximum size of the pool."),
	); err != nil {
		return err
	}

	if maxIdleDestroyCount, err = meter.Int64ObservableCounter(
		pgxpoolMaxIdleDestroyCount,
		metric.WithUnit(UnitDimensionless),
		metric.WithDescription("Cumulative count of connections destroyed because they exceeded MaxConnIdleTime."),
	); err != nil {
		return err
	}

	if maxLifetimeDestroyCountifetimeClosed, err = meter.Int64ObservableCounter(
		pgxpoolMaxLifetimeDestroyCount,
		metric.WithUnit(UnitDimensionless),
		metric.WithDescription("Cumulative count of connections destroyed because they exceeded MaxConnLifetime."),
	); err != nil {
		return err
	}

	if newConnsCount, err = meter.Int64ObservableCounter(
		pgxpoolNewConnsCount,
		metric.WithUnit(UnitDimensionless),
		metric.WithDescription("Cumulative count of new connections opened."),
	); err != nil {
		return err
	}

	if totalConns, err = meter.Int64ObservableUpDownCounter(
		pgxpoolTotalConns,
		metric.WithUnit(UnitDimensionless),
		metric.WithDescription("Total number of resources currently in the pool. The value is the sum of ConstructingConns, AcquiredConns, and IdleConns."),
	); err != nil {
		return err
	}

	_, err = meter.RegisterCallback(
		func(ctx context.Context, o metric.Observer) error {
			lock.Lock()
			defer lock.Unlock()

			now := time.Now()
			if now.Sub(lastDBStats) >= minimumReadDBStatsInterval {
				dbStats = db.Stat()
				lastDBStats = now
			}

			o.ObserveInt64(acquireCount, dbStats.AcquireCount(), attrs...)
			o.ObserveFloat64(acquireDuration, float64(dbStats.AcquireDuration())/1e6, attrs...)
			o.ObserveInt64(acquiredConns, int64(dbStats.AcquiredConns()), attrs...)
			o.ObserveInt64(cancelledAcquires, dbStats.CanceledAcquireCount(), attrs...)
			o.ObserveInt64(constructingConns, int64(dbStats.ConstructingConns()), attrs...)
			o.ObserveInt64(emptyAcquires, dbStats.EmptyAcquireCount(), attrs...)
			o.ObserveInt64(idleConns, int64(dbStats.IdleConns()), attrs...)
			o.ObserveInt64(maxConns, int64(dbStats.MaxConns()), attrs...)
			o.ObserveInt64(maxIdleDestroyCount, dbStats.MaxIdleDestroyCount(), attrs...)
			o.ObserveInt64(maxLifetimeDestroyCountifetimeClosed, dbStats.MaxLifetimeDestroyCount(), attrs...)
			o.ObserveInt64(newConnsCount, dbStats.NewConnsCount(), attrs...)
			o.ObserveInt64(totalConns, int64(dbStats.TotalConns()), attrs...)

			return nil
		},
		acquireCount,
		acquireDuration,
		acquiredConns,
		cancelledAcquires,
		constructingConns,
		emptyAcquires,
		idleConns,
		maxConns,
		maxIdleDestroyCount,
		maxLifetimeDestroyCountifetimeClosed,
		newConnsCount,
		totalConns,
	)

	return err
}

//  - Use `"1"` instead of `unit.Dimensionless`
//  - Use `"By"` instead of `unit.Bytes`
//  - Use `"ms"` instead of `unit.Milliseconds`

const (
	UnitDimensionless string = "1"
	UnitBytes         string = "By"
	UnitMilliseconds  string = "ms"
)
