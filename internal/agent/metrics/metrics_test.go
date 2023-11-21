package metrics

import (
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"slices"
	"testing"
	"time"
)

func Test_generateCustomMetrics(t *testing.T) {
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
				Value: int64(1),
			},
			want2: models.Metric{
				Name:  "RandomValue",
				Type:  "gauge",
				Value: int64(1),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pollMetric, randMetric models.Metric
			var counter int64
			generateCustomMetrics(&pollMetric, &randMetric, &counter)
			assert.Equal(t, tt.want1, pollMetric)
			assert.Equal(t, tt.want2.Name, randMetric.Name)
			assert.Equal(t, tt.want2.Type, randMetric.Type)
			oldRandValue := randMetric.Value
			requiredNewValue := pollMetric.Value.(int64) + 1
			generateCustomMetrics(&pollMetric, &randMetric, &counter)
			assert.Equal(t, requiredNewValue, pollMetric.Value, "Value of the PollCount metric "+
				"should be incremented to 1")
			assert.NotEqualf(t, oldRandValue, randMetric.Value, "Value of the RandomValue metric "+
				"should be a random values")
		})
	}
}

func Test_getSystemMetrics(t *testing.T) {
	metrics := getSystemMetrics()
	assert.Contains(t, metrics, "Alloc", "The system metrics is not contains 'Alloc' metric")
}

func Test_convertToFloat64(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{
			name:    "test uint32",
			value:   uint32(32),
			wantErr: false,
		},
		{
			name:    "test uint64",
			value:   uint64(32),
			wantErr: false,
		},
		{
			name:    "test float64",
			value:   float64(32),
			wantErr: false,
		},
		{
			name:    "test string",
			value:   "kek",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := convertToFloat64(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_getRequiredSystemMetrics(t *testing.T) {
	type args struct {
		m map[string]interface{}
		r []string
	}
	tests := []struct {
		name    string
		args    args
		want    []models.Metric
		wantErr bool
	}{
		{
			name: "successful test",
			args: args{
				m: map[string]interface{}{
					"Metric1": uint32(123),
					"Metric2": float64(321),
					"Metric3": uint64(0),
				},
				r: []string{"Metric1", "Metric3"},
			},
			wantErr: false,
			want: []models.Metric{
				{
					Name:  "Metric1",
					Type:  "gauge",
					Value: float64(123),
				},
				{
					Name:  "Metric3",
					Type:  "gauge",
					Value: float64(0),
				},
			},
		},
		{
			name: "unsuccessful test",
			args: args{
				m: map[string]interface{}{
					"Metric3": uint64(0),
				},
				r: []string{"Metric1", "Metric3"},
			},
			wantErr: true,
			want: []models.Metric{
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
			resultMetrics := getRequiredSystemMetrics(tt.args.m, tt.args.r)
			// reflect.DeepEqual работает некорректно когда мы работаем со слайсом карт, поэтому проверяем сами
			// если длина не совпадает - сразу выдаём ошибку
			// далее перебираем объекты из первого слайса и проверяем есть ли они во втором слайсе
			if len(resultMetrics) != len(tt.want) {
				require.Equal(t, tt.want, resultMetrics)
			}
			isEq := true
			for _, v1 := range resultMetrics {
				isContains := slices.Contains(tt.want, v1)
				if !isContains {
					isEq = false
					break
				}
			}
			if !tt.wantErr {
				assert.Truef(t, isEq, "Incorrect result metrics map\n"+
					"The maps must be equal\nWant: %+v\nGot: %+v",
					tt.want, resultMetrics)
			} else {
				assert.Falsef(t, isEq, "Incorrect result metrics map\n"+
					"The maps don't have to be equal\nWant: %+v\nGot: %+v",
					tt.want, resultMetrics)
			}
		})
	}
}

func TestGetMetrics(t *testing.T) {
	type args struct {
		c chan []models.Metric
		s time.Duration
		r []string
	}
	tests := []struct {
		name       string
		args       args
		checkAfter time.Duration
		wantErr    bool
		want       models.Metric
	}{
		{
			name: "successful test with custom metric",
			args: args{
				s: time.Duration(time.Millisecond * 500),
				r: []string{},
				c: make(chan []models.Metric, 1),
			},
			checkAfter: time.Duration(time.Millisecond * 750),
			wantErr:    false,
			want: models.Metric{
				Name:  "PollCount",
				Type:  "counter",
				Value: int64(2),
			},
		},
		{
			name: "successful test with system metric",
			args: args{
				s: time.Duration(time.Millisecond * 500),
				r: []string{"LastGC"},
				c: make(chan []models.Metric, 1),
			},
			checkAfter: time.Duration(time.Millisecond * 750),
			wantErr:    false,
			want: models.Metric{
				Name:  "LastGC",
				Type:  "gauge",
				Value: float64(0),
			},
		},
		{
			name: "unsuccessful test",
			args: args{
				s: time.Duration(time.Millisecond * 500),
				r: []string{"LastGC"},
				c: make(chan []models.Metric, 1),
			},
			checkAfter: time.Duration(time.Millisecond * 750),
			wantErr:    true,
			want: models.Metric{
				Name:  "Kek",
				Type:  "gauge",
				Value: float64(0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go GetMetrics(tt.args.c, tt.args.s, tt.args.r)
			time.Sleep(tt.checkAfter)
			var data []models.Metric
			select {
			case data = <-tt.args.c:
				//t.Logf("received %+v", data)
			default:
				//t.Log("empty")
			}

			if !tt.wantErr {
				assert.Contains(t, data, tt.want)
				//assert.Equal(t, tt.want, data)
			} else {
				assert.NotContains(t, data, tt.want)
			}
		})
	}
}
