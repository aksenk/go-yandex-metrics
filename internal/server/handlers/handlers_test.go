package handlers

import (
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MemStorageDummy struct {
	Dummy string
}

func (m MemStorageDummy) SaveMetric(metric *models.Metric) error {
	return nil
}

func (m MemStorageDummy) GetMetric(name string) (*models.Metric, error) {
	return &models.Metric{}, nil
}

func TestUpdateMetric(t *testing.T) {
	type args struct {
		method string
		path   string
	}
	tests := []struct {
		name     string
		wantCode int
		args     args
	}{
		{
			name:     "GET request unsuccessful",
			wantCode: 405,
			args: args{
				method: "GET",
				path:   "/update/",
			},
		},
		{
			name:     "POST request unsuccessful: no URL path",
			wantCode: 404,
			args: args{
				method: "POST",
				path:   "/",
			},
		},
		{
			name:     "POST request unsuccessful: only update path",
			wantCode: 400,
			args: args{
				method: "POST",
				path:   "/update/",
			},
		},
		{
			name:     "POST request unsuccessful: metric type path",
			wantCode: 404,
			args: args{
				method: "POST",
				path:   "/update/gauge/",
			},
		},
		{
			name:     "POST request unsuccessful: metric type + metric name path",
			wantCode: 400,
			args: args{
				method: "POST",
				path:   "/update/gauge/kek/",
			},
		},
		{
			name:     "POST request successful: gauge value with dot",
			wantCode: 200,
			args: args{
				method: "POST",
				path:   "/update/gauge/name/1.1",
			},
		},
		{
			name:     "POST request successful: gauge value without dot",
			wantCode: 200,
			args: args{
				method: "POST",
				path:   "/update/gauge/name/1",
			},
		},
		{
			name:     "POST request successful: switch metric type from gauge to counter",
			wantCode: 200,
			args: args{
				method: "POST",
				path:   "/update/counter/name/1",
			},
		},
		{
			name:     "POST request unsuccessful: counter float",
			wantCode: 400,
			args: args{
				method: "POST",
				path:   "/update/counter/name/1.1",
			},
		},
		{
			name:     "POST request unsuccessful: incorrect metric type",
			wantCode: 400,
			args: args{
				method: "POST",
				path:   "/update/gauges/name/1",
			},
		},
		{
			name:     "POST request successful: switch metric type from counter to gauge",
			wantCode: 200,
			args: args{
				method: "POST",
				path:   "/update/gauge/name/111.123",
			},
		},
	}
	for _, tt := range tests {
		storage := &MemStorageDummy{}
		t.Run(tt.name, func(t *testing.T) {
			handler := NewRouter(storage)
			server := httptest.NewServer(handler)
			request, err := http.NewRequest(tt.args.method, server.URL+tt.args.path, nil)
			require.NoError(t, err)
			result, err := server.Client().Do(request)
			require.NoError(t, err)
			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			result.Body.Close()
			assert.Equalf(t, tt.wantCode, result.StatusCode, "Response: %v", string(body))
		})
	}
}
