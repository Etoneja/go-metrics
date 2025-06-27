package server

import (
	"context"
	"log"
	"time"

	"github.com/etoneja/go-metrics/internal/models"
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

	return &DBStorage{pool: pool}
}

func (dbs *DBStorage) GetGauge(key string) (float64, bool, error) {
	panic("not implemented")
}

func (dbs *DBStorage) SetGauge(key string, value float64) (float64, error) {
	panic("not implemented") // TODO: Implement
}

func (dbs *DBStorage) GetCounter(key string) (int64, bool, error) {
	panic("not implemented") // TODO: Implement
}

func (dbs *DBStorage) IncrementCounter(key string, value int64) (int64, error) {
	panic("not implemented") // TODO: Implement
}

func (dbs *DBStorage) GetAll() *[]models.MetricModel {
	panic("not implemented") // TODO: Implement
}

func (dbs *DBStorage) ShutDown() {
	dbs.pool.Close()
}

func (dbs *DBStorage) Ping(ctx context.Context) error {
	err := dbs.pool.Ping(ctx)
	return err
}
