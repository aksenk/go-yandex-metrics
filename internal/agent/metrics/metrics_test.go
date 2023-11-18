package metrics

import (
	"github.com/aksenk/go-yandex-sprint1-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateCustomMetrics(t *testing.T) {
	tests := []struct {
		name  string
		want1 models.Metric
		want2 models.Metric
	}{
		{
			name: "test basic logic",
			want1: models.Metric{
				Name:  "PollCount",
				Type:  "counter",
				Value: 1,
			},
			want2: models.Metric{
				Name:  "RandomValue",
				Type:  "gauge",
				Value: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			get1, get2 := generateCustomMetrics()
			assert.Equal(t, get1, tt.want1)
			assert.Equal(t, tt.want2.Name, get2.Name)
			assert.Equal(t, tt.want2.Type, get2.Type)
			_, get3 := generateCustomMetrics()
			assert.NotEqualf(t, get2.Value, get3.Value, "Value should be a random values")
		})
	}
}
