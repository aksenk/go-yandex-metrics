package handlers

import (
	"compress/gzip"
	"encoding/json"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func Test_sendMetrics(t *testing.T) {
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
			name: "test counter metric",
			metrics: []metric{
				{
					Name:  "TestMetric",
					Type:  "counter",
					Value: 1,
				},
			},
		},
		{
			name: "test gauge metric",
			metrics: []metric{
				{
					Name:  "TestMetric",
					Type:  "gauge",
					Value: 1.123,
				},
			},
		},
		// отправку некорректных метрик не делаем, потому что некорректные метрики
		// отсекаются функцией models.NewMetric()
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requiredURL := "/update"
				// проверяем корректность http запроса
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "gzip", r.Header.Get("Accept-Encoding"))
				assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))
				assert.Equal(t, requiredURL, r.URL.RequestURI())
				// распаковываем gzip данные
				gz, err := gzip.NewReader(r.Body)
				assert.NoError(t, err)
				body, err := io.ReadAll(gz)
				assert.NoError(t, err)
				// пробуем распарсить данные в структуру метрики
				var m models.Metric
				err = json.Unmarshal(body, &m)
				assert.NoError(t, err)
				t.Log("Received request with correct gzipped json metric")
				w.Write(body)
			}))
			defer s.Close()
			var metrics []models.Metric
			for _, m := range tt.metrics {
				nm, err := models.NewMetric(m.Name, m.Type, m.Value)
				require.NoError(t, err)
				metrics = append(metrics, nm)
			}
			err := sendMetrics(metrics, s.URL+"/update")
			assert.NoError(t, err)
		})
	}
}

func TestHandleMetrics(t *testing.T) {
	type rawMetric struct {
		Name  string
		Type  string
		Value any
	}
	type args struct {
		handleAfter int
		checkAfter  int
		metrics     []rawMetric
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "successful test",
			wantErr: false,
			args: args{
				handleAfter: 1,
				checkAfter:  2,
				metrics: []rawMetric{
					{
						Name:  "FirstMetric",
						Type:  "gauge",
						Value: 1.123,
					},
					{
						Name:  "SecondMetric",
						Type:  "counter",
						Value: 10,
					},
				},
			},
		},
		{
			name:    "unsuccessful test",
			wantErr: true,
			args: args{
				handleAfter: 3,
				checkAfter:  2,
				metrics: []rawMetric{
					{
						Name:  "FirstMetric",
						Type:  "gauge",
						Value: 1.123,
					},
					{
						Name:  "SecondMetric",
						Type:  "counter",
						Value: 10,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		c := make(chan []models.Metric, 1)
		t.Run(tt.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("ok"))
			}))
			checkAfter := time.Second * time.Duration(tt.args.checkAfter)
			ticker := time.NewTicker(time.Second * time.Duration(tt.args.handleAfter))
			var metrics []models.Metric
			for _, m := range tt.args.metrics {
				nm, err := models.NewMetric(m.Name, m.Type, m.Value)
				require.NoError(t, err)
				metrics = append(metrics, nm)
			}
			c <- metrics
			go HandleMetrics(c, ticker, s.URL)
			time.Sleep(checkAfter)
			var data []models.Metric
			select {
			case data = <-c:
				if !tt.wantErr {
					assert.Nil(t, data, "The channel is not empty")
				}
			default:
				if tt.wantErr {
					assert.NotNil(t, data, "The channel is empty")
				}
			}
		})
	}
}
