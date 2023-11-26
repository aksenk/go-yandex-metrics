package handlers

import (
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http/httptest"
	"testing"
)

// кажется где-то тут кроются сакральные знания о интерфейсах
// я теперь могу замокать хранилище и прогнать тесты независимо
// но ряд кейсов (где валидируются входящие данные) я теперь проверить не могу
// а может это и не там совсем должно проверяться?
// в целом валидация данных это не проблемы сохраняющей функции

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
		storage MemStorageDummy
		method  string
		url     string
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
				storage: MemStorageDummy{},
				method:  "GET",
				url:     "http://localhost/kek",
			},
		},
		{
			name:     "POST request unsuccessful: no URL path",
			wantCode: 404,
			args: args{
				storage: MemStorageDummy{},
				method:  "POST",
				url:     "http://localhost/",
			},
		},
		{
			name:     "POST request unsuccessful: only update path",
			wantCode: 404,
			args: args{
				storage: MemStorageDummy{},
				method:  "POST",
				url:     "http://localhost/update/",
			},
		},
		{
			name:     "POST request unsuccessful: metric type path",
			wantCode: 404,
			args: args{
				storage: MemStorageDummy{},
				method:  "POST",
				url:     "http://localhost/update/gauge/",
			},
		},
		{
			name:     "POST request unsuccessful: metric type + metric name path",
			wantCode: 400,
			args: args{
				storage: MemStorageDummy{},
				method:  "POST",
				url:     "http://localhost/update/gauge/kek/",
			},
		},
		{
			name:     "POST request successful: gauge value with dot",
			wantCode: 200,
			args: args{
				storage: MemStorageDummy{},
				method:  "POST",
				url:     "http://localhost/update/gauge/name/1.1",
			},
		},
		{
			name:     "POST request successful: gauge value without dot",
			wantCode: 200,
			args: args{
				storage: MemStorageDummy{},
				method:  "POST",
				url:     "http://localhost/update/gauge/name/1",
			},
		},
		{
			name:     "POST request successful: switch metric type from gauge to counter",
			wantCode: 200,
			args: args{
				storage: MemStorageDummy{},
				method:  "POST",
				url:     "http://localhost/update/counter/name/1",
			},
		},
		{
			name:     "POST request unsuccessful: counter float",
			wantCode: 400,
			args: args{
				storage: MemStorageDummy{},
				method:  "POST",
				url:     "http://localhost/update/counter/name/1.1",
			},
		},
		{
			name:     "POST request unsuccessful: incorrect metric type",
			wantCode: 400,
			args: args{
				storage: MemStorageDummy{},
				method:  "POST",
				url:     "http://localhost/update/gauges/name/1",
			},
		},
		{
			name:     "POST request successful: switch metric type from counter to gauge",
			wantCode: 200,
			args: args{
				storage: MemStorageDummy{},
				method:  "POST",
				url:     "http://localhost/update/gauge/name/111.123",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(tt.args.method, tt.args.url, nil)
			handler := UpdateMetric(tt.args.storage)
			handler(recorder, request)
			result := recorder.Result()
			body, err := io.ReadAll(result.Body)
			result.Body.Close()
			if err != nil {
				t.Errorf("Error reading response body: %v", err)
			}
			assert.Equalf(t, tt.wantCode, result.StatusCode, "Response: %v", string(body))
			//if got := UpdateMetric(tt.args.storage); !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("UpdateMetric() = %v, want %v", got, tt.want)
			//}
		})
	}
}
