package models

import (
	"testing"
)

func TestMetric_String(t *testing.T) {
	type fields struct {
		Name  string
		Type  string
		Delta int64
		Value float64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "successfull test gauge",
			fields: fields{
				Name:  "kek",
				Type:  "gauge",
				Delta: 0,
				Value: 10,
			},
			want: "10",
		},
		{
			name: "successfull test counter",
			fields: fields{
				Name:  "fek",
				Type:  "counter",
				Delta: 11,
				Value: 0,
			},
			want: "11",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m Metric
			if tt.fields.Delta != 0 {
				m = Metric{
					ID:    tt.fields.Name,
					MType: tt.fields.Type,
					Delta: &tt.fields.Delta,
				}
			} else if tt.fields.Value != 0 {
				m = Metric{
					ID:    tt.fields.Name,
					MType: tt.fields.Type,
					Value: &tt.fields.Value,
				}
			} else {
				t.Errorf("Delta and value: both values cannot be non-zero at the same time")
			}
			if got := m.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
