package handlers

import (
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func Test_sendMetrics(t *testing.T) {
	type rawMetric struct {
		Name  string
		Type  string
		Value any
	}
	type args struct {
		metrics []rawMetric
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
				metrics: []rawMetric{
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
				metrics: []rawMetric{
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
				if r.Method != http.MethodPost {
					t.Error("HTTP method is not POST")
				}
				if tt.wantErr {
					assert.NotEqual(t, tt.args.path, r.URL.RequestURI())

				} else {
					assert.Equal(t, tt.args.path, r.URL.RequestURI())
				}
				w.Write([]byte(r.URL.RequestURI()))
			}))
			defer s.Close()
			var metrics []models.Metric
			for _, m := range tt.args.metrics {
				nm, err := models.NewMetric(m.Name, m.Type, m.Value)
				require.NoError(t, err)
				metrics = append(metrics, nm)
			}
			err := sendMetrics(metrics, s.URL)
			assert.Equal(t, nil, err)
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
