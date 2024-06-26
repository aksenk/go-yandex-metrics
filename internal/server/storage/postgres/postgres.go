package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/aksenk/go-yandex-metrics/internal/retry"
	"github.com/aksenk/go-yandex-metrics/internal/server/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"net"
	"time"
)

type Migrator struct {
	conn          *sql.DB
	logger        *zap.SugaredLogger
	cfg           *config.Config
	migrationsDir string
	state         MigrationStatus
}

type MigrationStatus struct {
	version uint
	dirty   bool
	err     error
}

func NewMigrator(conn *sql.DB, cfg *config.Config, log *zap.SugaredLogger) *Migrator {
	return &Migrator{
		conn:          conn,
		logger:        log,
		cfg:           cfg,
		migrationsDir: cfg.PostgresStorage.MigrationsDir,
	}
}

func (m *Migrator) Err() error {
	return m.state.err
}

func (m *Migrator) Dirty() bool {
	return m.state.dirty
}

func (m *Migrator) Version() uint {
	return m.state.version
}

func (m *Migrator) Run() {
	driver, err := postgres.WithInstance(m.conn, &postgres.Config{})
	if err != nil {
		m.state = MigrationStatus{version: 0, dirty: false, err: err}
		return
	}
	migration, err := migrate.NewWithDatabaseInstance("file://"+m.migrationsDir, "postgres", driver)
	if err != nil {
		m.state = MigrationStatus{version: 0, dirty: false, err: err}
		return
	}
	if err = migration.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			m.state = MigrationStatus{version: 0, dirty: false, err: err}
			return
		}
	}
	v, d, e := migration.Version()
	m.state = MigrationStatus{version: v, dirty: d, err: e}
}

type PostgresStorage struct {
	Conn   *sql.DB
	Logger *zap.SugaredLogger
	cfg    *config.Config
}

func NewPostgresStorage(cfg *config.Config, log *zap.SugaredLogger) (*PostgresStorage, error) {
	db, err := sql.Open("pgx", cfg.PostgresStorage.DSN)
	if err != nil {
		return nil, err
	}
	return &PostgresStorage{
		Conn:   db,
		Logger: log,
		cfg:    cfg,
	}, nil
}

func (p *PostgresStorage) SaveMetric(ctx context.Context, metric models.Metric) error {
	retryer := retry.NewRetryer(p.Logger, p.cfg.RetryConfig.RetryAttempts, time.Duration(p.cfg.RetryConfig.RetryWaitTime), func(ctx context.Context) (bool, error) {
		_, err := p.Conn.ExecContext(ctx, "INSERT INTO server.metrics (name, type, value, delta) "+
			"VALUES ($1, $2, $3, $4) ON CONFLICT (name) DO UPDATE SET type=$2, value=$3, delta=$4",
			metric.ID, metric.MType, metric.Value, metric.Delta)
		return false, err
	})
	return retryer.Do(ctx)
}

func (p *PostgresStorage) SaveBatchMetrics(ctx context.Context, metrics []models.Metric) error {
	retryer := retry.NewRetryer(p.Logger, p.cfg.RetryConfig.RetryAttempts, time.Duration(p.cfg.RetryConfig.RetryWaitTime), func(ctx context.Context) (bool, error) {
		tx, err := p.Conn.BeginTx(ctx, nil)
		if err != nil {
			return false, err
		}
		defer tx.Rollback()
		for _, metric := range metrics {
			_, err = p.Conn.ExecContext(ctx, "INSERT INTO server.metrics (name, type, value, delta) "+
				"VALUES ($1, $2, $3, $4) ON CONFLICT (name) DO UPDATE SET type=$2, value=$3, delta=$4",
				metric.ID, metric.MType, metric.Value, metric.Delta)
			if err != nil {
				return false, err
			}
		}
		return false, tx.Commit()
	})
	return retryer.Do(ctx)
}

func (p *PostgresStorage) GetMetric(ctx context.Context, metricName string) (*models.Metric, error) {
	var metric models.Metric
	retryer := retry.NewRetryer(p.Logger, p.cfg.RetryConfig.RetryAttempts, time.Duration(p.cfg.RetryConfig.RetryWaitTime), func(ctx context.Context) (bool, error) {
		err := p.Conn.QueryRowContext(ctx, "SELECT name, type, value, delta FROM server.metrics WHERE name = $1",
			metricName).
			Scan(&metric.ID, &metric.MType, &metric.Value, &metric.Delta)
		if errors.Is(err, sql.ErrNoRows) {
			return true, nil
		}

		if err != nil {
			var netErr net.Error
			// возвращаем ошибку (для выполнения ретрая) только при сетевых ошибках
			if errors.As(err, &netErr) {
				p.Logger.Errorf("Connection error: %s", err)
				return false, err
			}
			return true, nil
		}
		return true, nil
	})
	return &metric, retryer.Do(ctx)
}

func (p *PostgresStorage) GetAllMetrics(ctx context.Context) (map[string]models.Metric, error) {
	allMetrics := make(map[string]models.Metric)
	var metric models.Metric

	retryer := retry.NewRetryer(p.Logger, p.cfg.RetryConfig.RetryAttempts, time.Duration(p.cfg.RetryConfig.RetryWaitTime), func(ctx context.Context) (bool, error) {
		rows, err := p.Conn.QueryContext(ctx, "SELECT name, type, value, delta FROM server.metrics")
		if err != nil {
			return false, err
		}
		defer rows.Close()

		for rows.Next() {
			rows.Scan(&metric.ID, &metric.MType, &metric.Value, &metric.Delta)
			allMetrics[metric.ID] = metric
		}
		if err = rows.Err(); err != nil {
			return false, err
		}
		return true, nil
	})
	return allMetrics, retryer.Do(ctx)
}

func (p *PostgresStorage) StartupRestore(ctx context.Context) error {
	return nil
}

func (p *PostgresStorage) FlushMetrics() error {
	return nil
}

func (p *PostgresStorage) Status(ctx context.Context) error {
	retryer := retry.NewRetryer(p.Logger, p.cfg.RetryConfig.RetryAttempts, time.Duration(p.cfg.RetryConfig.RetryWaitTime), func(ctx context.Context) (bool, error) {
		timeout := 3 * time.Second
		DBCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		err := p.Conn.PingContext(DBCtx)
		if err != nil {
			return false, err
		}
		return true, nil
	})
	return retryer.Do(ctx)
}

func (p *PostgresStorage) Close() error {
	p.Logger.Debug("Closing postgres connection")
	return p.Conn.Close()
}
