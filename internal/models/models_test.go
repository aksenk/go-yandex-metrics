package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMetric(t *testing.T) {
	type args struct {
		metricName  string
		metricType  string
		metricValue string
	}
	tests := []struct {
		name    string
		args    args
		want    *Metric
		wantErr bool
	}{
		{
			name: "successful test: gauge metric",
			args: args{
				metricName:  "test_metric",
				metricType:  "gauge",
				metricValue: "1",
			},
			want: &Metric{
				Name:  "test_metric",
				Type:  "gauge",
				Value: float64(1),
			},
			wantErr: false,
		},
		{
			name: "successful test: gauge metric 2",
			args: args{
				metricName:  "test_metric",
				metricType:  "gauge",
				metricValue: "1.123",
			},
			want: &Metric{
				Name:  "test_metric",
				Type:  "gauge",
				Value: float64(1.123),
			},
			wantErr: false,
		},
		{
			name: "unsuccessful test: gauge metric with incorrect value",
			args: args{
				metricName:  "test_metric",
				metricType:  "gauge",
				metricValue: "asd",
			},
			want:    &Metric{},
			wantErr: true,
		},
		{
			name: "successful test: counter metric",
			args: args{
				metricName:  "test_metric",
				metricType:  "counter",
				metricValue: "1",
			},
			want: &Metric{
				Name:  "test_metric",
				Type:  "counter",
				Value: int64(1),
			},
			wantErr: false,
		},
		{
			name: "unsuccessful test: counter metric with incorrect value",
			args: args{
				metricName:  "test_metric",
				metricType:  "counter",
				metricValue: "asd",
			},
			want:    &Metric{},
			wantErr: true,
		},
		{
			name: "unsuccessful test: counter metric with incorrect float value",
			args: args{
				metricName:  "test_metric",
				metricType:  "counter",
				metricValue: "1.123",
			},
			want:    &Metric{},
			wantErr: true,
		},
		{
			name: "unsuccessful test: incorrect metric type",
			args: args{
				metricName:  "test_metric",
				metricType:  "kek",
				metricValue: "1",
			},
			want:    &Metric{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewMetric(tt.args.metricName, tt.args.metricType, tt.args.metricValue)
			if !tt.wantErr {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, m)
			} else {
				assert.Error(t, err)
			}

		})
	}
}
