package models

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			name: "successfully test gauge",
			fields: fields{
				Name:  "kek",
				Type:  "gauge",
				Delta: 0,
				Value: 10,
			},
			want: "10",
		},
		{
			name: "successfully test counter",
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
				t.Errorf("String() = %v, metricRaw %v", got, tt.want)
			}
		})
	}
}

func TestNewMetric(t *testing.T) {
	type metricRaw struct {
		name  string
		mtype string
		value any
	}
	tests := []struct {
		name    string
		args    metricRaw
		want    metricRaw
		wantErr bool
	}{
		{
			name: "successfully test gauge",
			want: metricRaw{
				name:  "testGauge",
				mtype: "gauge",
				value: 10.123,
			},
			args: metricRaw{
				name:  "testGauge",
				mtype: "gauge",
				value: 10.123,
			},
			wantErr: false,
		},
		{
			name: "successfully test counter",
			want: metricRaw{
				name:  "testCounter",
				mtype: "counter",
				value: 11,
			},
			args: metricRaw{
				name:  "testCounter",
				mtype: "counter",
				value: 11,
			},
			wantErr: false,
		},
		{
			name: "unsuccessfully test",
			want: metricRaw{
				name:  "testGauge",
				mtype: "gauge",
				value: 10.123,
			},
			args: metricRaw{
				name:  "incorrect",
				mtype: "counter",
				value: 10,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got interface{}
			got, err := NewMetric(tt.args.name, tt.args.mtype, tt.args.value)
			require.NoError(t, err)
			if _, ok := got.(Metric); !ok {
				t.Errorf("NewMetric() got = %v, metricRaw Metric struct", got)
			}
			if tt.wantErr {
				assert.NotEqual(t, tt.want.name, got.(Metric).ID)
				assert.NotEqual(t, tt.want.mtype, got.(Metric).MType)
				assert.NotEqual(t, fmt.Sprintf("%v", tt.want.value), got.(Metric).String())
			} else {
				assert.Equal(t, tt.want.name, got.(Metric).ID)
				assert.Equal(t, tt.want.mtype, got.(Metric).MType)
				assert.Equal(t, fmt.Sprintf("%v", tt.want.value), got.(Metric).String())
			}
		})
	}
}
