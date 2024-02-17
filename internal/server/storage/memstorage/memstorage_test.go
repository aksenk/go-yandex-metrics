package memstorage

import (
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
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
			m, err := storage.GetMetric(tt.getMetricName)
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
		name    string
		metric  metric
		want    *metric
		wantErr bool
	}{
		{
			name: "successful test: counter",
			metric: metric{
				Name:  "test_counter",
				Type:  "counter",
				Value: "1",
			},
			want: &metric{
				Name:  "test_counter",
				Type:  "counter",
				Value: "1",
			},
			wantErr: false,
		},
		{
			name: "successful test: switch counter to gauge",
			metric: metric{
				Name:  "test_counter",
				Type:  "gauge",
				Value: "1",
			},
			want: &metric{
				Name:  "test_counter",
				Type:  "gauge",
				Value: "1",
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
			want: &metric{
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
				Value: "1.123",
			},
			want: &metric{
				Name:  "test_gauge",
				Type:  "gauge",
				Value: "1.123",
			},
			wantErr: false,
		},
		{
			name: "successful test: switch gauge to counter",
			metric: metric{
				Name:  "test_gauge",
				Type:  "counter",
				Value: "1",
			},
			want: &metric{
				Name:  "test_gauge",
				Type:  "counter",
				Value: "1",
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
			want: &metric{
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
			want: &metric{
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
			want: &metric{
				Name:  "test_custom",
				Type:  "gauge",
				Value: "kek",
			},
			wantErr: true,
		},
		{
			name: "unsuccessful test: incorrect value",
			metric: metric{
				Name:  "test_gauge3",
				Type:  "gauge",
				Value: "1",
			},
			want: &metric{
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
			want: &metric{
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
			want: &metric{
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
			m, err := models.NewMetric(tt.metric.Name, tt.metric.Type, tt.metric.Value)
			if !tt.wantErr {
				require.NoError(t, err)
			}
			err = s.SaveMetric(m)
			require.NoError(t, err)
			if !tt.wantErr {
				assert.Equal(t, tt.want.Name, m.ID)
				assert.Equal(t, tt.want.Type, m.MType)
				assert.Equal(t, tt.want.Value, m.String())
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
			got := storage.GetAllMetrics()
			for _, w := range tt.want {
				if _, ok := got[w]; !ok {
					t.Errorf("GetAllMetrics() = metric with name '%v' does not contains in result metrics: %v", w, got)
				}
			}
		})
	}
}
