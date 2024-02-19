package postgres

import (
	"context"
	"database/sql"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"time"
)

type PostgresStorage struct {
	Conn *sql.DB
	Log  *zap.SugaredLogger
}

func NewPostgresStorage(connectionString string, log *zap.SugaredLogger) (*PostgresStorage, error) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, err
	}
	return &PostgresStorage{
		Conn: db,
		Log:  log,
	}, nil
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

func (p *PostgresStorage) Status(ctx context.Context) error {
	p.Log.Debugf("Checking postgres connection")
	timeout := 3 * time.Second
	DBCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	err := p.Conn.PingContext(DBCtx)
	if err != nil {
		p.Log.Errorf("Postgres connection is not OK: %v", err)
		return err
	}
	p.Log.Debugf("Postgres connection is OK")
	return nil
}

func (p *PostgresStorage) Close() error {
	p.Log.Debugf("Closing postgres connection")
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
		if err == migrate.ErrNoChange {
			return m.Version()
		}
		return 0, false, err
	}
	return m.Version()
}
