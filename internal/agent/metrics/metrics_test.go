package metrics

import (
	"github.com/aksenk/go-yandex-sprint1-metrics/internal/models"
	"reflect"
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
			if !reflect.DeepEqual(get1, tt.want1) {
				t.Errorf("PollCounter metric error. Objects is not equal. Get: %+v, want: %+v",
					get1, tt.want1)
			}
			if tt.want2.Name != get2.Name {
				t.Errorf("RandomValue metric error. Objects names is not equal. Get: %+v, want: %+v",
					get2.Name, tt.want2.Name)
			}
			if tt.want2.Type != get2.Type {
				t.Errorf("RandomValue metric error. Objects types is not equal. Get: %+v, want: %+v",
					get2.Type, tt.want2.Type)
			}
			_, get3 := generateCustomMetrics()
			if get3.Value == get2.Value {
				t.Errorf("RandomValue metric has the same values after second run. "+
					"It should be a random value. Get1: %v, get2: %v", get2.Value, get3.Value)
			}
		})
	}
}
