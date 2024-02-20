package postgres

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func CreateMockedStorage() (*PostgresStorage, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}
	logger, err := logger.NewLogger("debug")
	if err != nil {
		return nil, nil, err
	}
	mockedStorage := PostgresStorage{
		Conn:   db,
		Logger: logger,
	}
	return &mockedStorage, mock, nil
}

func TestPostgresStorage_Status(t *testing.T) {
	t.Run("mocked storage check status function", func(t *testing.T) {
		db, mock, err := CreateMockedStorage()
		require.NoError(t, err)

		mock.ExpectPing()

		err = db.Status(context.TODO())
		assert.NoError(t, err)

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("real storage check status function (invalid DSN)", func(t *testing.T) {
		log, err := logger.NewLogger("debug")
		require.NoError(t, err)

		db, err := NewPostgresStorage("dummy", log)
		require.NoError(t, err)
		defer db.Conn.Close()

		err = db.Status(context.TODO())
		var expectedErrString = "cannot parse `dummy`: failed to parse as DSN (invalid dsn)"
		assert.EqualError(t, err, expectedErrString)
	})

	t.Run("real storage check status function", func(t *testing.T) {
		log, err := logger.NewLogger("debug")
		require.NoError(t, err)

		db, err := NewPostgresStorage("postgres://postgres:password@localhost:5432/db", log)
		require.NoError(t, err)
		defer db.Conn.Close()

		err = db.Status(context.TODO())
		assert.Error(t, err)
	})
}

func TestNewPostgresStorage(t *testing.T) {
	logger, err := logger.NewLogger("info")
	require.NoError(t, err)
	t.Run("test new postgres storage", func(t *testing.T) {
		var got any
		var err error
		got, err = NewPostgresStorage("dummy connection string", logger)
		require.NoError(t, err)
		if _, ok := got.(*PostgresStorage); !ok {
			t.Fatalf("Resulting object have incorrect type (not equal *PostgresStorage struct)")
		}
	})
}

func TestPostgresStorage_Close(t *testing.T) {
	t.Run("test postgres storage close function", func(t *testing.T) {
		db, mock, err := CreateMockedStorage()
		require.NoError(t, err)

		mock.ExpectClose()

		db.Close()
		if err = mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}
