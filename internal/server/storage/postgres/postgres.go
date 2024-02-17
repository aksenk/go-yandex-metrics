package postgres

import (
	"context"
	"database/sql"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"time"
)

type PostgresStorage struct {
	db  *sql.DB
	log *zap.SugaredLogger
}

func NewPostgresStorage(ctx context.Context, connectionString string, timeout time.Duration, log *zap.SugaredLogger) (*PostgresStorage, error) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, err
	}
	DBCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	err = db.PingContext(DBCtx)
	if err != nil {
		return nil, err
	}
	return &PostgresStorage{
		db:  db,
		log: log,
	}, nil
}

func (p *PostgresStorage) Close() error {
	return p.db.Close()
}

func (p *PostgresStorage) SaveMetric(metric models.Metric) error {
	return nil
}

func (p *PostgresStorage) GetMetric(name string) (*models.Metric, error) {
	return nil, nil
}

func (p *PostgresStorage) GetAllMetrics() map[string]models.Metric {
	return nil
}

func (p *PostgresStorage) StartupRestore() error {
	return nil
}

func (p *PostgresStorage) FlushMetrics() error {
	return nil
}
