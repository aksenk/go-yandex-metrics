package converter

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

//func Test_convertToFloat64(t *testing.T) {
//	tests := []struct {
//		name    string
//		value   any
//		wantErr bool
//	}{
//		{
//			name:    "test uint32",
//			value:   uint32(32),
//			wantErr: false,
//		},
//		{
//			name:    "test uint64",
//			value:   uint64(32),
//			wantErr: false,
//		},
//		{
//			name:    "test float64",
//			value:   float64(32),
//			wantErr: false,
//		},
//		{
//			name:    "test string",
//			value:   "kek",
//			wantErr: true,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			_, err := converter.AnyToFloat64(tt.value)
//			if tt.wantErr {
//				assert.Error(t, err)
//			} else {
//				assert.NoError(t, err)
//			}
//		})
//	}
//}

func TestAnyToFloat64(t *testing.T) {
	type args struct {
		value any
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{
			name: "test uint32",
			args: args{
				value: uint32(32),
			},
			want:    32,
			wantErr: false,
		},
		{
			name: "test uint64",
			args: args{
				value: uint64(32),
			},
			want:    32,
			wantErr: false,
		},
		{
			name: "test float64",
			args: args{
				value: float64(32),
			},
			want:    32,
			wantErr: false,
		},
		{
			name: "test string",
			args: args{
				value: "kek",
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "test nil",
			args: args{
				value: nil,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "test bool",
			args: args{
				value: true,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "test int",
			args: args{
				value: 11,
			},
			want:    11,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AnyToFloat64(tt.args.value)
			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAnyToInt64(t *testing.T) {
	type args struct {
		value any
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "test uint32",
			args: args{
				value: uint32(32),
			},
			want:    32,
			wantErr: false,
		},
		{
			name: "test uint64",
			args: args{
				value: uint64(32),
			},
			want:    32,
			wantErr: false,
		},
		{
			name: "test float64",
			args: args{
				value: float64(32),
			},
			want:    32,
			wantErr: false,
		},
		{
			name: "test string",
			args: args{
				value: "kek",
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "test nil",
			args: args{
				value: nil,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "test bool",
			args: args{
				value: true,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "test int",
			args: args{
				value: 11,
			},
			want:    11,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AnyToInt64(tt.args.value)
			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
