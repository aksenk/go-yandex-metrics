package postgres

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"github.com/stretchr/testify/require"
	"testing"
)

func createMockedStorage() (*PostgresStorage, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}
	logger, err := logger.NewLogger("debug")
	if err != nil {
		return nil, nil, err
	}
	mockedStorage := PostgresStorage{
		db:  db,
		log: logger,
	}
	return &mockedStorage, mock, nil
}

func TestPostgresStorage_Status(t *testing.T) {
	t.Run("check status function", func(t *testing.T) {
		db, mock, err := createMockedStorage()
		require.NoError(t, err)

		mock.ExpectPing()

		db.Status(context.TODO())

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestNewPostgresStorage(t *testing.T) {
	t.Run("test new postgres storage", func(t *testing.T) {
		log := logger.Log
		var got any
		var err error
		got, err = NewPostgresStorage("dummy connection string", log)
		require.NoError(t, err)
		if _, ok := got.(*PostgresStorage); !ok {
			t.Fatalf("Resulting object have incorrect type (not equal *PostgresStorage struct)")
		}
	})
}

func TestPostgresStorage_Close(t *testing.T) {
	t.Run("test postgres storage close function", func(t *testing.T) {
		db, mock, err := createMockedStorage()
		require.NoError(t, err)

		mock.ExpectClose()

		db.Close()
		if err = mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}