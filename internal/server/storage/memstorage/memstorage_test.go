package memstorage

import (
	"context"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestMemStorage_GetMetric(t *testing.T) {
	type metric struct {
		Name  string
		Type  string
		Value any
	}
	tests := []struct {
		name            string
		getMetricName   string
		existingMetrics map[string]metric
		want            *metric
		wantErr         bool
	}{
		{
			name:          "successful test: get existing metric",
			getMetricName: "test_metric",
			existingMetrics: map[string]metric{
				"test_metric": {
					Name:  "test_metric",
					Type:  "gauge",
					Value: "1",
				},
			},
			want: &metric{
				Name:  "test_metric",
				Type:  "gauge",
				Value: "1",
			},
			wantErr: false,
		},
		{
			name:            "unsuccessful test: get not existing metric",
			getMetricName:   "test_metric",
			existingMetrics: map[string]metric{},
			want: &metric{
				Name:  "test_metric",
				Type:  "gauge",
				Value: "1",
			},
			wantErr: true,
		},
		{
			name:            "unsuccessful test: get unexpected metric with empty name",
			getMetricName:   "",
			existingMetrics: map[string]metric{},
			want: &metric{
				Name:  "test_metric",
				Type:  "gauge",
				Value: "1",
			},
			wantErr: true,
		},
		{
			name:            "unsuccessful test: get expected metric with empty name",
			getMetricName:   "",
			existingMetrics: map[string]metric{},
			want:            &metric{},
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			existingMetrics := make(map[string]models.Metric)
			for _, m := range tt.existingMetrics {
				nm, err := models.NewMetric(m.Name, m.Type, m.Value)
				require.NoError(t, err)
				existingMetrics[m.Name] = nm
			}
			storage := MemStorage{
				Metrics: existingMetrics,
				mu:      sync.Mutex{},
			}
			m, err := storage.GetMetric(context.TODO(), tt.getMetricName)
			if !tt.wantErr {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.Name, m.ID)
				assert.Equal(t, tt.want.Type, m.MType)
				assert.Equal(t, tt.want.Value, m.String())
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestMemStorage_SaveMetric(t *testing.T) {
	type metric struct {
		Name  string
		Type  string
		Value any
	}

	tests := []struct {
		name   string
		metric metric
	}{
		{
			name: "counter",
			metric: metric{
				Name:  "test_counter",
				Type:  "counter",
				Value: "1",
			},
		},
		{
			name: "switch counter to gauge",
			metric: metric{
				Name:  "test_counter",
				Type:  "gauge",
				Value: "1",
			},
		},
		{
			name: "gauge",
			metric: metric{
				Name:  "test_gauge",
				Type:  "gauge",
				Value: "1.123",
			},
		},
		{
			name: "switch gauge to counter",
			metric: metric{
				Name:  "test_gauge",
				Type:  "counter",
				Value: "1",
			},
		},
	}
	for _, tt := range tests {
		log, err := logger.NewLogger("debug")
		require.NoError(t, err)

		s := NewMemStorage(log)

		t.Run(tt.name, func(t *testing.T) {
			m, err := models.NewMetric(tt.metric.Name, tt.metric.Type, tt.metric.Value)
			require.NoError(t, err)

			err = s.SaveMetric(context.TODO(), m)
			require.NoError(t, err)

			assert.Equal(t, s.Metrics[tt.metric.Name].ID, tt.metric.Name)
			assert.Equal(t, s.Metrics[tt.metric.Name].MType, tt.metric.Type)
		})
	}
}

func TestMemStorage_SaveBatchMetrics(t *testing.T) {
	type metric struct {
		Name  string
		Type  string
		Value any
	}

	tests := []struct {
		name    string
		metrics []metric
	}{
		{
			name: "insert new metrics",
			metrics: []metric{
				{
					Name:  "test_counter",
					Type:  "counter",
					Value: "1",
				},
				{
					Name:  "test_gauge",
					Type:  "gauge",
					Value: "12",
				},
			},
		},
		{
			name: "update existing metrics",
			metrics: []metric{
				{
					Name:  "test_counter",
					Type:  "counter",
					Value: "1",
				},
				{
					Name:  "test_gauge",
					Type:  "gauge",
					Value: "12",
				},
			},
		},
	}

	log, err := logger.NewLogger("debug")
	require.NoError(t, err)
	s := NewMemStorage(log)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var metrics []models.Metric
			for _, m := range tt.metrics {
				mm, err := models.NewMetric(m.Name, m.Type, m.Value)
				require.NoError(t, err)
				metrics = append(metrics, mm)
			}

			err = s.SaveBatchMetrics(context.TODO(), metrics)
			require.NoError(t, err)

			for _, m := range metrics {
				assert.Equal(t, s.Metrics[m.ID].ID, m.ID)
				assert.Equal(t, s.Metrics[m.ID].MType, m.MType)
			}
		})
	}
}

func TestNewMemStorage(t *testing.T) {
	tests := []struct {
		name string
		want *MemStorage
	}{
		{
			name: "return correct MemStorage",
			want: &MemStorage{
				Metrics: map[string]models.Metric{},
			},
		},
	}
	logger, err := logger.NewLogger("info")
	require.NoError(t, err)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got any = NewMemStorage(logger)
			if _, ok := got.(*MemStorage); !ok {
				t.Fatalf("Resulting object have incorrect type (not equal *MemStorage struct)")
			}
		})
	}
}

func TestMemStorage_GetAllMetrics(t *testing.T) {
	type metric struct {
		Name  string
		Type  string
		Value any
	}
	tests := []struct {
		name            string
		existingMetrics map[string]metric
		want            []string
	}{
		{
			name: "successful test: get existing metric",
			want: []string{"test_metric1", "test_metric2"},
			existingMetrics: map[string]metric{
				"test_metric1": {
					Name:  "test_metric1",
					Type:  "gauge",
					Value: "1.23",
				},
				"test_metric2": {
					Name:  "test_metric2",
					Type:  "counter",
					Value: "123",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			existingMetrics := make(map[string]models.Metric)
			for _, m := range tt.existingMetrics {
				nm, err := models.NewMetric(m.Name, m.Type, m.Value)
				require.NoError(t, err)
				existingMetrics[m.Name] = nm
			}
			storage := MemStorage{
				Metrics: existingMetrics,
				mu:      sync.Mutex{},
			}
			got, err := storage.GetAllMetrics(context.TODO())
			require.NoError(t, err)
			for _, w := range tt.want {
				if _, ok := got[w]; !ok {
					t.Errorf("GetAllMetrics() = metric with name '%v' does not contains in result metrics: %v", w, got)
				}
			}
		})
	}
}
