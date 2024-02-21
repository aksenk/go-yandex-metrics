package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"time"
)

type PostgresStorage struct {
	Conn   *sql.DB
	Logger *zap.SugaredLogger
}

func NewPostgresStorage(connectionString string, log *zap.SugaredLogger) (*PostgresStorage, error) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, err
	}
	return &PostgresStorage{
		Conn:   db,
		Logger: log,
	}, nil
}

func (p *PostgresStorage) SaveMetric(ctx context.Context, metric models.Metric) error {
	_, err := p.GetMetric(ctx, metric.ID)
	if errors.Is(err, storage.ErrMetricNotExist) {
		_, err = p.Conn.ExecContext(ctx, "INSERT INTO server.metrics (name, type, value, delta) VALUES ($1, $2, $3, $4)",
			metric.ID, metric.MType, metric.Value, metric.Delta)
		return err
	}
	if err != nil {
		return err
	}
	_, err = p.Conn.ExecContext(ctx, "UPDATE server.metrics SET type=$1, value=$2, delta=$3 WHERE name=$4",
		metric.MType, metric.Value, metric.Delta, metric.ID)
	return err
}

func (p *PostgresStorage) GetMetric(ctx context.Context, metricName string) (*models.Metric, error) {
	var metric models.Metric
	err := p.Conn.QueryRowContext(ctx, "SELECT name, type, value, delta FROM server.metrics WHERE name = $1",
		metricName).
		Scan(&metric.ID, &metric.MType, &metric.Value, &metric.Delta)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, storage.ErrMetricNotExist
	}
	if err != nil {
		return nil, err
	}
	return &metric, nil
}

func (p *PostgresStorage) GetAllMetrics(ctx context.Context) (map[string]models.Metric, error) {
	allMetrics := make(map[string]models.Metric)
	var metric models.Metric

	rows, err := p.Conn.QueryContext(ctx, "SELECT name, type, value, delta FROM server.metrics")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&metric.ID, &metric.MType, &metric.Value, &metric.Delta)
		allMetrics[metric.ID] = metric
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return allMetrics, nil
}

func (p *PostgresStorage) StartupRestore(ctx context.Context) error {
	return nil
}

func (p *PostgresStorage) FlushMetrics() error {
	return nil
}

func (p *PostgresStorage) Status(ctx context.Context) error {
	p.Logger.Debugf("Checking postgres connection")
	timeout := 3 * time.Second
	DBCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	err := p.Conn.PingContext(DBCtx)
	if err != nil {
		p.Logger.Errorf("Postgres connection is not OK: %v", err)
		return err
	}
	p.Logger.Debugf("Postgres connection is OK")
	return nil
}

func (p *PostgresStorage) Close() error {
	p.Logger.Debugf("Closing postgres connection")
	return p.Conn.Close()
}

func RunMigrations(migrationsDir string, conn *sql.DB) (version uint, dirty bool, err error) {
	driver, err := postgres.WithInstance(conn, &postgres.Config{})
	if err != nil {
		return 0, false, err
	}
	m, err := migrate.NewWithDatabaseInstance("file://"+migrationsDir, "postgres", driver)
	if err != nil {
		return 0, false, err
	}
	if err = m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return m.Version()
		}
		return 0, false, err
	}
	return m.Version()
}
