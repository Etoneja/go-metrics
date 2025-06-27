package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/etoneja/go-metrics/internal/common"
	"github.com/etoneja/go-metrics/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const MaxConns = 10
const MinConns = 2
const MaxConnLifetime = time.Hour
const MaxConnIdleTime = time.Minute * 30

type DBStorage struct {
	pool *pgxpool.Pool
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

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
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

	_, err := dbs.pool.Exec(ctx, `
		CREATE TABLE metrics (
			id varchar(150) primary key,
			delta bigint null,
			value double precision null
		);`)
	if err != nil {
		return fmt.Errorf("failed to create metrics table: %w", err)
	}

	log.Println("migrations completed successfully")
	return err

}

func (dbs *DBStorage) GetGauge(key string) (float64, bool, error) {
	ctx := context.Background()

	row := dbs.pool.QueryRow(ctx, "select value from metrics where id = $1;", key)

	var value sql.NullFloat64
	err := row.Scan(&value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, err
	}
	if !value.Valid {
		return 0, false, nil
	}
	return value.Float64, true, nil
}

func (dbs *DBStorage) SetGauge(key string, value float64) (float64, error) {
	ctx := context.Background()

	row := dbs.pool.QueryRow(ctx, `
		INSERT INTO metrics (id, value) 
		VALUES ($1, $2)
		ON CONFLICT (id)
		DO UPDATE SET
			value = $2
		RETURNING value;
	`, key, value)

	var newvalue float64
	err := row.Scan(&newvalue)
	if err != nil {
		return 0, err
	}
	return newvalue, nil
}

func (dbs *DBStorage) GetCounter(key string) (int64, bool, error) {
	ctx := context.Background()

	row := dbs.pool.QueryRow(ctx, "select delta from metrics where id = $1;", key)

	var value sql.NullInt64
	err := row.Scan(&value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, err
	}
	if !value.Valid {
		return 0, false, nil
	}
	return value.Int64, true, nil
}

func (dbs *DBStorage) IncrementCounter(key string, value int64) (int64, error) {
	ctx := context.Background()

	row := dbs.pool.QueryRow(ctx, `
		INSERT INTO metrics (id, delta) 
		VALUES ($1, $2)
		ON CONFLICT (id)
		DO UPDATE SET
			delta = coalesce(metrics.delta, 0) + $2
		RETURNING delta;
	`, key, value)

	var newvalue int64
	err := row.Scan(&newvalue)
	if err != nil {
		return 0, err
	}
	return newvalue, nil
}

func (dbs *DBStorage) GetAll() (*[]models.MetricModel, error) {
	ctx := context.Background()

	rows, err := dbs.pool.Query(ctx, "select id, delta, value from metrics;")

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
			metrics = append(metrics, models.MetricModel{ID: id, MType: common.MetricTypeCounter, Delta: &delta.Int64})
		}

		if value.Valid {
			metrics = append(metrics, models.MetricModel{ID: id, MType: common.MetricTypeGauge, Value: &value.Float64})
		}

	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return &metrics, nil

}

func (dbs *DBStorage) ShutDown() {
	log.Println("Shutdowning db storage")
	dbs.pool.Close()
}

func (dbs *DBStorage) Ping(ctx context.Context) error {
	err := dbs.pool.Ping(ctx)
	return err
}
