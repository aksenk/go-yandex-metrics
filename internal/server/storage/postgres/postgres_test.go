package postgres

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/aksenk/go-yandex-metrics/internal/server/config"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func CreateMockedStorage() (*PostgresStorage, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
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
		cfg: &config.Config{
			Storage:  storage.PostgresStorage,
			LogLevel: "debug",
			PostgresStorage: config.PostgresConfig{
				DSN:  "dummy",
				Type: "postgres",
			},
		},
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

		cfg := config.Config{
			Storage: storage.PostgresStorage,
			PostgresStorage: config.PostgresConfig{
				DSN:  "dummy",
				Type: "postgres",
			},
		}
		db, err := NewPostgresStorage(&cfg, log)
		require.NoError(t, err)
		defer db.Conn.Close()

		err = db.Status(context.TODO())
		var expectedErrString = "cannot parse `dummy`: failed to parse as DSN (invalid dsn)"
		assert.EqualError(t, err, expectedErrString)
	})

	t.Run("real storage check status function", func(t *testing.T) {
		log, err := logger.NewLogger("debug")
		require.NoError(t, err)

		cfg := config.Config{
			Storage: storage.PostgresStorage,
			PostgresStorage: config.PostgresConfig{
				DSN:  "dummy",
				Type: "postgres",
			},
		}

		db, err := NewPostgresStorage(&cfg, log)
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
		cfg := config.Config{
			Storage: storage.PostgresStorage,
			PostgresStorage: config.PostgresConfig{
				DSN:  "dummy",
				Type: "postgres",
			},
		}
		got, err = NewPostgresStorage(&cfg, logger)
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

func TestPostgresStorage_GetMetric(t *testing.T) {
	type metric struct {
		Name  string
		Type  string
		Value any
	}
	emptyMetric := &models.Metric{
		ID:    "",
		MType: "",
		Value: nil,
		Delta: nil,
	}
	t.Run("metric exist", func(t *testing.T) {
		var m = metric{
			Name:  "test_metric",
			Type:  "counter",
			Value: 1,
		}
		checkMetric, err := models.NewMetric(m.Name, m.Type, m.Value)
		require.NoError(t, err)

		db, mock, err := CreateMockedStorage()
		require.NoError(t, err)

		mock.ExpectQuery("SELECT name, type, value, delta FROM server.metrics WHERE name = $1").WithArgs(checkMetric.ID).
			WillReturnRows(sqlmock.NewRows([]string{"name", "type", "value", "delta"}).AddRow(checkMetric.ID, checkMetric.MType, checkMetric.Value, checkMetric.Delta))

		gotMetric, err := db.GetMetric(context.TODO(), checkMetric.ID)
		require.NoError(t, err)
		assert.Equal(t, gotMetric.ID, checkMetric.ID)
		assert.Equal(t, gotMetric.MType, checkMetric.MType)
		assert.Equal(t, gotMetric.Value, checkMetric.Value)
		assert.Equal(t, gotMetric.Delta, checkMetric.Delta)

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}

	})

	t.Run("metric not exist", func(t *testing.T) {
		var m = metric{
			Name:  "test_metric",
			Type:  "gauge",
			Value: 123,
		}
		checkMetric, err := models.NewMetric(m.Name, m.Type, m.Value)
		require.NoError(t, err)

		db, mock, err := CreateMockedStorage()
		require.NoError(t, err)

		mock.ExpectQuery("SELECT name, type, value, delta FROM server.metrics WHERE name = $1").WithArgs(checkMetric.ID).WillReturnError(storage.ErrMetricNotExist)
		gotMetric, err := db.GetMetric(context.TODO(), checkMetric.ID)
		assert.ErrorIs(t, err, storage.ErrMetricNotExist)
		assert.Equal(t, gotMetric, emptyMetric)

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}

	})
}

func TestPostgresStorage_SaveMetric(t *testing.T) {
	type metric struct {
		Name  string
		Type  string
		Value any
	}
	t.Run("successfully", func(t *testing.T) {
		var m = metric{
			Name:  "test_metric",
			Type:  "counter",
			Value: 11,
		}
		checkMetric, err := models.NewMetric(m.Name, m.Type, m.Value)
		require.NoError(t, err)

		db, mock, err := CreateMockedStorage()
		require.NoError(t, err)

		mock.ExpectExec("INSERT INTO server.metrics (name, type, value, delta) VALUES ($1, $2, $3, $4) ON CONFLICT (name) DO UPDATE SET type=$2, value=$3, delta=$4").WithArgs(checkMetric.ID, checkMetric.MType, checkMetric.Value, checkMetric.Delta).WillReturnResult(sqlmock.NewResult(1, 1))
		err = db.SaveMetric(context.TODO(), checkMetric)
		assert.NoError(t, err)

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestPostgresStorage_GetAllMetrics(t *testing.T) {
	t.Run("get all metrics", func(t *testing.T) {
		db, mock, err := CreateMockedStorage()
		require.NoError(t, err)

		mock.ExpectQuery("SELECT name, type, value, delta FROM server.metrics").
			WillReturnRows(sqlmock.NewRows([]string{"name", "type", "value", "delta"}).
				AddRow("test1", "gauge", 1, nil).
				AddRow("test2", "counter", nil, 2))

		gotMetrics, err := db.GetAllMetrics(context.TODO())
		require.NoError(t, err)
		assert.Equal(t, gotMetrics["test1"].ID, "test1")
		assert.Equal(t, gotMetrics["test1"].MType, "gauge")
		assert.EqualValues(t, *gotMetrics["test1"].Value, 1)
		assert.Nil(t, gotMetrics["test1"].Delta)
		assert.Equal(t, gotMetrics["test2"].ID, "test2")
		assert.Equal(t, gotMetrics["test2"].MType, "counter")
		assert.Nil(t, gotMetrics["test2"].Value)
		assert.EqualValues(t, *gotMetrics["test2"].Delta, 2)

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestPostgresStorage_SaveBatchMetric(t *testing.T) {
	type metric struct {
		Name  string
		Type  string
		Value any
	}
	t.Run("successfully", func(t *testing.T) {
		var rawMetrics = []metric{
			{
				Name:  "test_metric",
				Type:  "counter",
				Value: 11,
			},
			{
				Name:  "test_metric2",
				Type:  "gauge",
				Value: 1,
			},
		}
		var checkMetrics []models.Metric

		db, mock, err := CreateMockedStorage()
		require.NoError(t, err)

		mock.ExpectBegin()

		for _, m := range rawMetrics {
			nm, err := models.NewMetric(m.Name, m.Type, m.Value)
			require.NoError(t, err)
			checkMetrics = append(checkMetrics, nm)

			mock.ExpectExec("INSERT INTO server.metrics (name, type, value, delta) VALUES ($1, $2, $3, $4) ON CONFLICT (name) DO UPDATE SET type=$2, value=$3, delta=$4").
				WithArgs(nm.ID, nm.MType, nm.Value, nm.Delta).
				WillReturnResult(sqlmock.NewResult(1, 1))
		}

		mock.ExpectCommit()

		err = db.SaveBatchMetrics(context.TODO(), checkMetrics)
		assert.NoError(t, err)

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}
