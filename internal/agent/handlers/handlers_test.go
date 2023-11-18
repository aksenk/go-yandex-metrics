package handlers

import (
	"github.com/aksenk/go-yandex-sprint1-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// вроде как имена тестов нужно писать через CamelCase, но если это приватная функция
// то генератор в IDE даёт ей название через snake_case. это ок?
func Test_generateSendURL(t *testing.T) {
	tests := []struct {
		name    string
		metric  models.Metric
		wantErr bool
		want    string
	}{
		{
			name: "basic logic test",
			metric: models.Metric{
				Name:  "PollInterval",
				Type:  "counter",
				Value: 1,
			},
			wantErr: false,
			want:    "/counter/PollInterval/1",
		},
		{
			name: "basic logic test dummy values",
			metric: models.Metric{
				Name:  "tratatatata",
				Type:  "akunamatata",
				Value: "keeeks",
			},
			wantErr: false,
			want:    "/akunamatata/tratatatata/keeeks",
		},
		{
			name: "error test",
			metric: models.Metric{
				Name:  "PollInterval",
				Type:  "counter",
				Value: 1,
			},
			wantErr: true,
			want:    "/gauge/PollInterval/1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			get := generateSendURL(tt.metric, "")
			if tt.wantErr {
				assert.NotEqual(t, tt.want, get)
			} else {
				assert.Equal(t, tt.want, get)
			}
		})
	}
}

func Test_sendMetrics(t *testing.T) {
	type args struct {
		metrics []models.Metric
		path    string
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
				metrics: []models.Metric{
					{
						Name:  "TestMetric",
						Type:  "counter",
						Value: 1,
					},
				},
				path: "/counter/TestMetric/1",
			},
		},
		{
			name:    "unsuccessful test",
			wantErr: true,
			args: args{
				metrics: []models.Metric{
					{
						Name:  "TestMetric",
						Type:  "counter",
						Value: 1,
					},
				},
				path: "/gauge/TestMetric/1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.wantErr {
					assert.NotEqual(t, tt.args.path, r.URL.RequestURI())

				} else {
					assert.Equal(t, tt.args.path, r.URL.RequestURI())
				}
				w.Write([]byte(r.URL.RequestURI()))
			}))
			defer s.Close()
			err := sendMetrics(tt.args.metrics, s.URL)
			assert.Equal(t, nil, err)
		})
	}
}

func TestHandleMetrics(t *testing.T) {
	type args struct {
		handleAfter int
		checkAfter  int
		metrics     []models.Metric
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
				metrics: []models.Metric{
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
				metrics: []models.Metric{
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
			c <- tt.args.metrics
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
