package metrics

import (
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func Test_GenerateCustomMetrics(t *testing.T) {
	type want struct {
		Name  string
		Type  string
		Value any
	}
	tests := []struct {
		name  string
		want1 want
		want2 want
	}{
		{
			name: "test custom metrics",
			want1: want{
				Name:  "PollCount",
				Type:  "counter",
				Value: 1,
			},
			want2: want{
				Name:  "RandomValue",
				Type:  "gauge",
				Value: 1.123,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var counter int64 = 1
			want1, err := models.NewMetric(tt.want1.Name, tt.want1.Type, tt.want1.Value)
			require.NoError(t, err)

			want2, err := models.NewMetric(tt.want2.Name, tt.want2.Type, tt.want2.Value)
			require.NoError(t, err)

			pollMetric, randMetric := GenerateCustomMetrics(counter)
			if !reflect.DeepEqual(want1, pollMetric) {
				t.Error("RuntimeRequiredMetrics are not equals")
			}

			assert.Equal(t, want2.ID, randMetric.ID)
			assert.Equal(t, want2.MType, randMetric.MType)

			oldRandValue := randMetric.Value
			requiredNewValue := *pollMetric.Delta + 1

			counter++

			pollMetric, randMetric = GenerateCustomMetrics(counter)
			assert.Equal(t, requiredNewValue, *pollMetric.Delta, "Value of the PollCount metric "+
				"should be incremented to 1")
			assert.NotEqualf(t, oldRandValue, randMetric.Value, "Value of the RandomValue metric "+
				"should be a random values")
		})
	}
}

func Test_GetSystemMetrics(t *testing.T) {
	metrics := GetSystemMetrics()
	assert.Contains(t, metrics, "Alloc", "The system handlers is not contains 'Alloc' metric")
}

func Test_RemoveUnnecessaryMetrics(t *testing.T) {
	type args struct {
		systemMetrics   map[string]interface{}
		requiredMetrics []string
	}
	type want struct {
		Name  string
		Type  string
		Delta int64
		Value float64
	}
	tests := []struct {
		name    string
		args    args
		want    []want
		wantErr bool
	}{
		{
			name: "successful test",
			args: args{
				systemMetrics: map[string]interface{}{
					"Metric1": uint32(123),
					"Metric2": float64(321),
					"Metric3": uint64(11),
				},
				requiredMetrics: []string{"Metric1", "Metric3"},
			},
			wantErr: false,
			want: []want{
				{
					Name:  "Metric1",
					Type:  "gauge",
					Value: 123,
				},
				{
					Name:  "Metric3",
					Type:  "gauge",
					Value: 11,
				},
			},
		},
		{
			name: "unsuccessful test",
			args: args{
				systemMetrics: map[string]interface{}{
					"Metric3": uint64(0),
				},
				requiredMetrics: []string{"Metric1", "Metric3"},
			},
			wantErr: true,
			want: []want{
				{
					Name:  "Metric1",
					Type:  "gauge",
					Value: float64(123),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultMetrics, err := RemoveUnnecessaryMetrics(tt.args.systemMetrics, tt.args.requiredMetrics)
			require.NoError(t, err)
			// reflect.DeepEqual работает некорректно когда мы работаем со слайсом карт, поэтому проверяем сами
			// если длина не совпадает - сразу выдаём ошибку
			// далее перебираем объекты из первого слайса и проверяем есть ли они во втором слайсе
			if len(resultMetrics) != len(tt.want) {
				require.Equal(t, tt.want, resultMetrics)
			}
			var wantedMetrics []models.Metric
			for _, v := range tt.want {
				m, err := models.NewMetric(v.Name, v.Type, v.Value)
				require.NoError(t, err)
				wantedMetrics = append(wantedMetrics, m)
			}
			isEq := false
			for _, rm := range resultMetrics {
				for _, vm := range wantedMetrics {
					if reflect.DeepEqual(rm, vm) {
						isEq = true
						break
					}
				}
			}
			if !tt.wantErr {
				assert.Truef(t, isEq, "Incorrect result handlers map\n"+
					"The maps must be equal\nWant: %+v\nGot: %+v",
					tt.want, resultMetrics)
			} else {
				assert.Falsef(t, isEq, "Incorrect result handlers map\n"+
					"The maps don't have to be equal\nWant: %+v\nGot: %+v",
					tt.want, resultMetrics)
			}
		})
	}
}
