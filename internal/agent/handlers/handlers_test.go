package handlers

import (
	"github.com/aksenk/go-yandex-sprint1-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateSendURL(t *testing.T) {
	tests := []struct {
		name   string
		metric models.Metric
		want   string
	}{
		{
			name: "basic logic test",
			metric: models.Metric{
				Name:  "PollInterval",
				Type:  "counter",
				Value: 1,
			},
			want: "/counter/PollInterval/1",
		},
		{
			name: "basic logic test dummy values",
			metric: models.Metric{
				Name:  "tratatatata",
				Type:  "akunamatata",
				Value: "keeeks",
			},
			want: "/akunamatata/tratatatata/keeeks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			get := generateSendURL(tt.metric, "")
			assert.Equal(t, tt.want, get)
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
			name:    "correct test",
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
			name:    "incorrect test",
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
				t.Log(tt.args.path)
				t.Log(r.URL.RequestURI())
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
