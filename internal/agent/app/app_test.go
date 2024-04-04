package app

import (
	"context"
	"github.com/aksenk/go-yandex-metrics/internal/agent/config"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

func createTestApp(t *testing.T, cfg *config.Config) *App {
	logLevel := "debug"
	logger, err := logger.NewLogger(logLevel)
	require.NoError(t, err)
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	app, err := NewApp(&client, logger, cfg)
	require.NoError(t, err)
	return app
}

/*
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
			cfg := config.Config{
				ServerURL:      "localhost:8080",
				ReportInterval: 10,
				PollInterval:   10,
				LogLevel:       "debug",
			}
			cfg.ServerURL = s.URL + "/update"
			app := createTestApp(t, &cfg)
			err := app.sendBatchMetrics(metrics)
			assert.NoError(t, err)
		})
	}
}
*/

func TestWaitMetrics(t *testing.T) {
	type rawMetric struct {
		Name  string
		Type  string
		Value any
	}
	tests := []struct {
		name         string
		metrics      []rawMetric
		sendInterval int
		checkAfter   int
		wantErr      bool
	}{
		{
			name:         "successful test",
			wantErr:      false,
			sendInterval: 1,
			checkAfter:   2,
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
		{
			name:         "unsuccessful test",
			wantErr:      true,
			checkAfter:   1,
			sendInterval: 10,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//_ = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//	w.Write([]byte("ok"))
			//}))
			var metrics []models.Metric
			for _, m := range tt.metrics {
				nm, err := models.NewMetric(m.Name, m.Type, m.Value)
				require.NoError(t, err)
				metrics = append(metrics, nm)
			}

			cfg := config.Config{
				ServerURL:      "localhost:8080",
				ReportInterval: time.Second * time.Duration(tt.sendInterval),
				PollInterval:   time.Second*time.Duration(tt.checkAfter) + 1,
				LogLevel:       "debug",
			}
			app := createTestApp(t, &cfg)
			app.ReadyMetrics <- metrics

			go app.WaitMetrics(context.TODO())

			time.Sleep(time.Second * time.Duration(tt.checkAfter))

			var data []models.Metric
			select {
			case data = <-app.ReadyMetrics:
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

func Test_GetMetrics(t *testing.T) {
	type want struct {
		Name  string
		Type  string
		Value any
	}
	tests := []struct {
		name                   string
		sleepTime              time.Duration
		pollInterval           time.Duration
		runtimeRequiredMetrics []string
		wantErr                bool
		want                   want
	}{
		{
			name:                   "successful test with custom metric",
			sleepTime:              time.Millisecond * 750,
			pollInterval:           time.Millisecond * 500,
			runtimeRequiredMetrics: []string{},
			wantErr:                false,
			want: want{
				Name:  "PollCount",
				Type:  "counter",
				Value: 2,
			},
		},
		{
			name:                   "successful test with system metric",
			sleepTime:              time.Millisecond * 750,
			pollInterval:           time.Millisecond * 500,
			runtimeRequiredMetrics: []string{"LastGC"},
			wantErr:                false,
			want: want{
				Name:  "LastGC",
				Type:  "gauge",
				Value: 0,
			},
		},
		{
			name:                   "unsuccessful test",
			sleepTime:              time.Millisecond * 750,
			pollInterval:           time.Millisecond * 500,
			runtimeRequiredMetrics: []string{"LastGC"},
			wantErr:                true,
			want: want{
				Name:  "Kek",
				Type:  "gauge",
				Value: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{
				ServerURL:      "localhost:8080",
				ReportInterval: tt.pollInterval + 10,
				PollInterval:   tt.pollInterval,
				LogLevel:       "debug",
			}
			app := createTestApp(t, &cfg)
			app.RuntimeRequiredMetrics = tt.runtimeRequiredMetrics

			wantMetric, err := models.NewMetric(tt.want.Name, tt.want.Type, tt.want.Value)
			require.NoError(t, err)

			go app.GetRuntimeMetrics(context.TODO())
			time.Sleep(tt.sleepTime)
			var data []models.Metric
			select {
			case data = <-app.ReadyMetrics:
				//t.Logf("received %+v", data)
			default:
				//t.Log("empty")
			}
			isContains := false
			for _, m := range data {
				if m.ID == wantMetric.ID {
					isContains = true
				}
			}
			if !tt.wantErr {
				assert.True(t, isContains)
			} else {
				assert.False(t, isContains)
			}
		})
	}
}
