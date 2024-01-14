package handlers

import (
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage/memstorage"
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

func (m MemStorageDummy) SaveMetric(metric models.Metric) error {
	return nil
}

func (m MemStorageDummy) GetMetric(name string) (*models.Metric, error) {
	return &models.Metric{}, nil
}

func (m MemStorageDummy) GetAllMetrics() map[string]models.Metric {
	return make(map[string]models.Metric)
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
			wantCode: 500,
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

func TestGetMetric(t *testing.T) {
	type metric struct {
		Name  string
		Type  string
		Value any
	}
	tests := []struct {
		name           string
		storageMetrics map[string]metric
		requestURL     string
		wantCode       int
		wantBody       string
	}{
		{
			name: "successful test: gauge",
			storageMetrics: map[string]metric{
				"test": {
					Name:  "test",
					Type:  "gauge",
					Value: 10.123,
				},
			},
			requestURL: "/value/gauge/test",
			wantCode:   200,
			wantBody:   "10.123\n",
		},
		{
			name: "successful test: counter",
			storageMetrics: map[string]metric{
				"test": {
					Name:  "test",
					Type:  "counter",
					Value: 5,
				},
			},
			requestURL: "/value/counter/test",
			wantCode:   200,
			wantBody:   "5\n",
		},
		{
			name: "unsuccessful test: gauge",
			storageMetrics: map[string]metric{
				"test": {
					Name:  "test",
					Type:  "gauge",
					Value: 10.123,
				},
			},
			requestURL: "/value/counter/test",
			wantCode:   404,
			wantBody:   "Error receiving metric: metric not found\n",
		},
		{
			name: "unsuccessful test: counter",
			storageMetrics: map[string]metric{
				"test": {
					Name:  "test",
					Type:  "counter",
					Value: 5,
				},
			},
			requestURL: "/value/gauge/test",
			wantCode:   404,
			wantBody:   "Error receiving metric: metric not found\n",
		},
		{
			name:           "unsuccessful test: without metric type",
			storageMetrics: map[string]metric{},
			requestURL:     "/value/",
			wantCode:       400,
			wantBody:       "Missing metric type\n",
		},
		{
			name:           "unsuccessful test: without metric name",
			storageMetrics: map[string]metric{},
			requestURL:     "/value/gauge/",
			wantCode:       404,
			wantBody:       "Missing metric name\n",
		},
		{
			name:           "unsuccessful test: with a slash at the end",
			storageMetrics: map[string]metric{},
			requestURL:     "/value/gauge/test/",
			wantCode:       404,
			wantBody:       "404 page not found\n",
		},
		{
			name:           "unsuccessful test: missing gauge metric",
			storageMetrics: map[string]metric{},
			requestURL:     "/value/gauge/test",
			wantCode:       404,
			wantBody:       "Error receiving metric: metric not found\n",
		},
		{
			name:           "unsuccessful test: missing counter metric",
			storageMetrics: map[string]metric{},
			requestURL:     "/value/counter/test",
			wantCode:       404,
			wantBody:       "Error receiving metric: metric not found\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storageMetrics := make(map[string]models.Metric)
			for _, m := range tt.storageMetrics {
				nm, err := models.NewMetric(m.Name, m.Type, m.Value)
				require.NoError(t, err)
				storageMetrics[m.Name] = nm
			}
			storage := &memstorage.MemStorage{
				Metrics: storageMetrics,
			}
			server := httptest.NewServer(NewRouter(storage))
			response, err := server.Client().Get(server.URL + tt.requestURL)
			require.NoError(t, err)
			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			err = response.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, tt.wantCode, response.StatusCode)
			assert.Equal(t, tt.wantBody, string(body))
		})
	}
}

func TestListAllMetrics(t *testing.T) {
	type metric struct {
		Name  string
		Type  string
		Value any
	}
	tests := []struct {
		name     string
		metrics  map[string]metric
		wantCode int
		wantBody string
	}{
		{
			name:     "successful test: no metrics",
			metrics:  map[string]metric{},
			wantCode: 200,
			wantBody: "<html><head><title>all metrics</title></head>" +
				"<body><h1>List of all metrics</h1><p></p></body></html>\n",
		},
		{
			name: "successful test: one metric",
			metrics: map[string]metric{
				"first": {
					Name:  "first",
					Type:  "gauge",
					Value: 1,
				},
			},
			wantCode: 200,
			wantBody: "<html><head><title>all metrics</title></head>" +
				"<body><h1>List of all metrics</h1><p>first=1</p></body></html>\n",
		},
		{
			name: "successful test: two metric",
			metrics: map[string]metric{
				"first": {
					Name:  "first",
					Type:  "gauge",
					Value: 1,
				},
				"second": {
					Name:  "second",
					Type:  "counter",
					Value: 2,
				},
			},
			wantCode: 200,
			wantBody: "<html><head><title>all metrics</title></head>" +
				"<body><h1>List of all metrics</h1><p>first=1</p><p>second=2</p></body></html>\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := make(map[string]models.Metric)
			for _, m := range tt.metrics {
				nm, err := models.NewMetric(m.Name, m.Type, m.Value)
				require.NoError(t, err)
				metrics[m.Name] = nm
			}
			storage := &memstorage.MemStorage{
				Metrics: metrics,
			}
			server := httptest.NewServer(ListAllMetrics(storage))
			response, err := server.Client().Get(server.URL + "/")
			require.NoError(t, err)
			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			stringBody := string(body)
			err = response.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, tt.wantCode, response.StatusCode)
			assert.Equal(t, tt.wantBody, stringBody)
		})
	}
}
