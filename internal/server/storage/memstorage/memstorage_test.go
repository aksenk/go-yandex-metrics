package memstorage

import (
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"reflect"
	"sync"
	"testing"
)

func TestMemStorage_GetMetric(t *testing.T) {
	tests := []struct {
		name          string
		getMetricName string
		storage       MemStorage
		want          *models.Metric
		wantErr       bool
	}{
		{
			name:          "successful test: get existing metric",
			getMetricName: "test_metric",
			storage: MemStorage{
				Metrics: map[string]models.Metric{
					"test_metric": models.Metric{
						Name:  "test_metric",
						Type:  "gauge",
						Value: "1",
					},
				},
			},
			want: &models.Metric{
				Name:  "test_metric",
				Type:  "gauge",
				Value: "1",
			},
			wantErr: false,
		},
		{
			name:          "unsuccessful test: get not existing metric",
			getMetricName: "test_metric",
			storage: MemStorage{
				Metrics: map[string]models.Metric{},
			},
			want: &models.Metric{
				Name:  "test_metric",
				Type:  "gauge",
				Value: "1",
			},
			wantErr: true,
		},
		{
			name:          "unsuccessful test: get unexpected metric with empty name",
			getMetricName: "",
			storage: MemStorage{
				Metrics: map[string]models.Metric{},
			},
			want: &models.Metric{
				Name:  "test_metric",
				Type:  "gauge",
				Value: "1",
			},
			wantErr: true,
		},
		{
			name:          "unsuccessful test: get expected metric with empty name",
			getMetricName: "",
			storage: MemStorage{
				Metrics: map[string]models.Metric{},
			},
			want:    &models.Metric{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := tt.storage.GetMetric(tt.getMetricName)
			if !tt.wantErr {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, m)
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
		name    string
		metric  metric
		want    *models.Metric
		wantErr bool
	}{
		{
			name: "successful test: counter",
			metric: metric{
				Name:  "test_counter",
				Type:  "counter",
				Value: int64(1),
			},
			want: &models.Metric{
				Name:  "test_counter",
				Type:  "counter",
				Value: int64(1),
			},
			wantErr: false,
		},
		{
			name: "successful test: switch counter to gauge",
			metric: metric{
				Name:  "test_counter",
				Type:  "gauge",
				Value: float64(1),
			},
			want: &models.Metric{
				Name:  "test_counter",
				Type:  "gauge",
				Value: float64(1),
			},
			wantErr: false,
		},
		{
			name: "successful test: counter",
			metric: metric{
				Name:  "test_counter2",
				Type:  "counter",
				Value: "1",
			},
			want: &models.Metric{
				Name:  "test_counter2",
				Type:  "counter",
				Value: "1",
			},
			wantErr: false,
		},
		{
			name: "successful test: gauge",
			metric: metric{
				Name:  "test_gauge",
				Type:  "gauge",
				Value: float64(1.123),
			},
			want: &models.Metric{
				Name:  "test_gauge",
				Type:  "gauge",
				Value: float64(1.123),
			},
			wantErr: false,
		},
		{
			name: "successful test: switch gauge to counter",
			metric: metric{
				Name:  "test_gauge",
				Type:  "counter",
				Value: int64(1),
			},
			want: &models.Metric{
				Name:  "test_gauge",
				Type:  "counter",
				Value: int64(1),
			},
			wantErr: false,
		},
		{
			name: "successful test: gauge",
			metric: metric{
				Name:  "test_gauge2",
				Type:  "gauge",
				Value: "1.123",
			},
			want: &models.Metric{
				Name:  "test_gauge2",
				Type:  "gauge",
				Value: "1.123",
			},
			wantErr: false,
		},
		{
			name: "successful test: custom type",
			metric: metric{
				Name:  "test_custom",
				Type:  "gauge",
				Value: "1.123",
			},
			want: &models.Metric{
				Name:  "test_custom",
				Type:  "gauge",
				Value: "1.123",
			},
			wantErr: false,
		},
		{
			name: "successful test: custom type 2",
			metric: metric{
				Name:  "test_custom",
				Type:  "gauge",
				Value: "kek",
			},
			want: &models.Metric{
				Name:  "test_custom",
				Type:  "gauge",
				Value: "kek",
			},
			wantErr: false,
		},
		{
			name: "unsuccessful test: incorrect value",
			metric: metric{
				Name:  "test_gauge3",
				Type:  "gauge",
				Value: "1",
			},
			want: &models.Metric{
				Name:  "test_gauge3",
				Type:  "gauge",
				Value: "2",
			},
			wantErr: true,
		},
		{
			name: "unsuccessful test: incorrect type",
			metric: metric{
				Name:  "test_gauge3",
				Type:  "gauge",
				Value: "1",
			},
			want: &models.Metric{
				Name:  "test_gauge3",
				Type:  "counter",
				Value: "1",
			},
			wantErr: true,
		},
		{
			name: "unsuccessful test: incorrect metric name",
			metric: metric{
				Name:  "test_gauge3",
				Type:  "gauge",
				Value: "1",
			},
			want: &models.Metric{
				Name:  "test_gauge4",
				Type:  "gauge",
				Value: "1",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		s := &MemStorage{
			Metrics: map[string]models.Metric{},
			mu:      sync.Mutex{},
		}
		t.Run(tt.name, func(t *testing.T) {
			m := &models.Metric{
				Name:  tt.metric.Name,
				Type:  tt.metric.Type,
				Value: tt.metric.Value,
			}
			err := s.SaveMetric(m)
			assert.NoError(t, err)
			if !tt.wantErr {
				assert.Equal(t, tt.want, m)
			} else {
				assert.NotEqual(t, tt.want, m)
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMemStorage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMemStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}
