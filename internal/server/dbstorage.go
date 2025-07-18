package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"sort"
	"time"

	"github.com/etoneja/go-metrics/internal/common"
	"github.com/etoneja/go-metrics/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const MaxConns = 10
const MinConns = 2
const MaxConnLifetime = time.Hour
const MaxConnIdleTime = time.Minute * 30

const (
	queryInsertGauge = `
		INSERT INTO metrics (id, value) 
		VALUES ($1, $2)
		ON CONFLICT (id)
		DO UPDATE SET
			value = $2
		RETURNING value;
	`
	queryInsertCounter = `
		INSERT INTO metrics (id, delta) 
		VALUES ($1, $2)
		ON CONFLICT (id)
		DO UPDATE SET
			delta = coalesce(metrics.delta, 0) + $2
		RETURNING delta;
	`
	querySelectCounter      = "select delta from metrics where id = $1;"
	querySelectGauge        = "select value from metrics where id = $1;"
	querySelectAllMetrics   = "select id, delta, value from metrics;"
	queryCreateMetricsTable = `
		CREATE TABLE IF NOT EXISTS metrics (
			id varchar(150) primary key,
			delta bigint null,
			value double precision null
	);`
)

type DBStorage struct {
	pool *pgxpool.Pool
}

func isDBRetryableError(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr)
}

func NewDBStorage(connString string) *DBStorage {
	ctx := context.Background()
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Fatalf("Unable to parse config: %v", err)
	}

	config.MaxConns = MaxConns
	config.MinConns = MinConns
	config.MaxConnLifetime = MaxConnLifetime
	config.MaxConnIdleTime = MaxConnIdleTime

	var pool *pgxpool.Pool
	var poolErr error

	backoffSchedule := common.DefaultBackoffSchedule
	backoffTicker := common.GetBackoffTicker(ctx, backoffSchedule)
	attemptNum := 0
	for range backoffTicker {
		attemptNum++
		attemptString := fmt.Sprintf("[%d/%d]", attemptNum, len(backoffSchedule)+1)

		pool, poolErr = pgxpool.NewWithConfig(ctx, config)
		if poolErr != nil {
			if isDBRetryableError(poolErr) {
				log.Printf("%s retryable error, will retry: %v", attemptString, poolErr)
				pool.Close()
				continue
			}
			log.Fatalf("%s non-retryable error: %v", attemptString, poolErr)
		}

		poolErr = pool.Ping(ctx)
		if poolErr != nil {
			if isDBRetryableError(poolErr) {
				log.Printf("%s retryable ping error, will retry: %v", attemptString, poolErr)
				pool.Close()
				continue
			}
			log.Fatalf("%s non-retryable ping error: %v", attemptString, poolErr)
		}

		break
	}

	if poolErr != nil {
		log.Fatalf("Failed to connect to DB after all attempts: %v", poolErr)
	}

	dbs := &DBStorage{pool: pool}

	err = dbs.runMigrations(ctx)
	if err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}
	return dbs
}

func (dbs *DBStorage) runMigrations(ctx context.Context) error {
	log.Println("runinning migrations")

	_, err := dbs.pool.Exec(ctx, queryCreateMetricsTable)
	if err != nil {
		return fmt.Errorf("failed to create metrics table: %w", err)
	}

	log.Println("migrations completed successfully")
	return err

}

func (dbs *DBStorage) GetGauge(ctx context.Context, key string) (float64, error) {
	row := dbs.pool.QueryRow(ctx, querySelectGauge, key)

	var value sql.NullFloat64
	err := row.Scan(&value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("%s %s: %w", common.MetricTypeGauge, key, ErrNotFound)
		}
		return 0, err
	}
	if !value.Valid {
		return 0, fmt.Errorf("%s %s: %w", common.MetricTypeGauge, key, ErrNotFound)
	}
	return value.Float64, nil
}

func (dbs *DBStorage) SetGauge(ctx context.Context, key string, value float64) (float64, error) {
	row := dbs.pool.QueryRow(ctx, queryInsertGauge, key, value)

	var newvalue float64
	err := row.Scan(&newvalue)
	if err != nil {
		return 0, err
	}
	return newvalue, nil
}

func (dbs *DBStorage) GetCounter(ctx context.Context, key string) (int64, error) {
	row := dbs.pool.QueryRow(ctx, querySelectCounter, key)

	var value sql.NullInt64
	err := row.Scan(&value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("%s %s: %w", common.MetricTypeCounter, key, ErrNotFound)
		}
		return 0, err
	}
	if !value.Valid {
		return 0, fmt.Errorf("%s %s: %w", common.MetricTypeCounter, key, ErrNotFound)
	}
	return value.Int64, nil
}

func (dbs *DBStorage) IncrementCounter(ctx context.Context, key string, value int64) (int64, error) {
	row := dbs.pool.QueryRow(ctx, queryInsertCounter, key, value)

	var newvalue int64
	err := row.Scan(&newvalue)
	if err != nil {
		return 0, err
	}
	return newvalue, nil
}

func (dbs *DBStorage) GetAll(ctx context.Context) ([]models.MetricModel, error) {
	rows, err := dbs.pool.Query(ctx, querySelectAllMetrics)

	if err != nil {
		return nil, err
	}

	metrics := []models.MetricModel{}

	var id string
	var delta sql.NullInt64
	var value sql.NullFloat64

	for rows.Next() {

		err = rows.Scan(&id, &delta, &value)
		if err != nil {
			return nil, err
		}

		if delta.Valid {
			metrics = append(metrics, *models.NewMetricModel(id, common.MetricTypeCounter, delta.Int64, 0))
		}

		if value.Valid {
			metrics = append(metrics, *models.NewMetricModel(id, common.MetricTypeGauge, 0, value.Float64))
		}

	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].ID < metrics[j].ID
	})

	return metrics, nil

}

func (dbs *DBStorage) ShutDown() {
	log.Println("Shutdowning db storage")
	dbs.pool.Close()
}

func (dbs *DBStorage) Ping(ctx context.Context) error {
	err := dbs.pool.Ping(ctx)
	return err
}

func (dbs *DBStorage) BatchUpdate(ctx context.Context, metrics []models.MetricModel) ([]models.MetricModel, error) {
	newMetrics := make([]models.MetricModel, 0, len(metrics))

	metricsCopy := make([]models.MetricModel, len(metrics))
	copy(metricsCopy, metrics)
	sort.Slice(metricsCopy, func(i, j int) bool {
		return metricsCopy[i].ID < metricsCopy[j].ID
	})

	tx, err := dbs.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	gaugeStmt, err := tx.Prepare(ctx, "insert-gauge", queryInsertGauge)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare gauge statement: %w", err)
	}

	counterStmt, err := tx.Prepare(ctx, "insert-counter", queryInsertCounter)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare counter statement: %w", err)
	}

	var row pgx.Row
	for _, m := range metricsCopy {
		switch m.MType {
		case common.MetricTypeCounter:
			row = tx.QueryRow(ctx, counterStmt.Name, m.ID, *m.Delta)

			var newDelta int64
			err := row.Scan(&newDelta)
			if err != nil {
				return nil, err
			}
			newMetrics = append(newMetrics, *models.NewMetricModel(m.ID, m.MType, newDelta, 0))
		case common.MetricTypeGauge:
			row := tx.QueryRow(ctx, gaugeStmt.Name, m.ID, *m.Value)

			var newValue float64
			err := row.Scan(&newValue)
			if err != nil {
				return nil, err
			}
			newMetrics = append(newMetrics, *models.NewMetricModel(m.ID, m.MType, 0, newValue))
		default:
			return nil, fmt.Errorf("unknown metric type %s", m.MType)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return newMetrics, nil
}
