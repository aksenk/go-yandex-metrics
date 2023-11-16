package handlers

import (
	"github.com/aksenk/go-yandex-sprint1-metrics/internal/models"
	"testing"
)

func TestGenerateSendURL(t *testing.T) {
	tests := []struct {
		name   string
		metric models.Metric
		URL    string
		want   string
	}{
		{
			name: "basic logic test",
			metric: models.Metric{
				Name:  "PollInterval",
				Type:  "counter",
				Value: 1,
			},
			URL:  "http://localhost/update",
			want: "http://localhost/update/counter/PollInterval/1",
		},
		{
			name: "basic logic test dummy values",
			metric: models.Metric{
				Name:  "tratatatata",
				Type:  "akunamatata",
				Value: "keeeks",
			},
			URL:  "hophaylalaley",
			want: "hophaylalaley/akunamatata/tratatatata/keeeks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if get := generateSendURL(tt.metric, tt.URL); get != tt.want {
				t.Errorf("get: %v, want: %v", get, tt.want)
			}
		})
	}
}
