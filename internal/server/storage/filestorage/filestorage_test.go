package filestorage

import (
	"bufio"
	"encoding/json"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestNewFileStorage(t *testing.T) {
	type args struct {
		filename         string
		synchronousFlush bool
	}
	tests := []struct {
		name    string
		args    args
		want    args
		wantErr bool
	}{
		{
			name: "successfully test",
			args: args{
				filename:         "./test-storage",
				synchronousFlush: false,
			},
			want: args{
				filename:         "./test-storage",
				synchronousFlush: false,
			},
			wantErr: false,
		},
		{
			name: "failed test: empty filename",
			args: args{
				filename:         "",
				synchronousFlush: false,
			},
			want: args{
				filename:         "./test-storage",
				synchronousFlush: false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got any
			var err error
			defer os.RemoveAll(tt.args.filename)
			got, err = NewFileStorage(tt.args.filename, tt.args.synchronousFlush)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if _, ok := got.(*FileStorage); ok {
					assert.Equal(t, tt.want.filename, tt.args.filename)
					assert.Equal(t, tt.want.synchronousFlush, tt.args.synchronousFlush)
				} else {
					t.Fatalf("Resulting object have incorrect type (not equal Filestorage struct)")
				}
				err = os.RemoveAll(tt.args.filename)
				assert.NoError(t, err)
			}
		})
	}
}

func TestFileStorageSaveMetric(t *testing.T) {
	type args struct {
		filename         string
		synchronousFlush bool
	}
	type metric struct {
		Name  string
		Type  string
		Value any
	}

	tests := []struct {
		name    string
		args    args
		metric  metric
		want    *metric
		wantErr bool
	}{
		{
			name: "successful test: counter",
			args: args{
				filename:         "./test-storage",
				synchronousFlush: false,
			},
			metric: metric{
				Name:  "test_counter",
				Type:  "counter",
				Value: "1",
			},
			want: &metric{
				Name:  "test_counter",
				Type:  "counter",
				Value: "1",
			},
			wantErr: false,
		},
		{
			name: "successful test: switch counter to gauge",
			args: args{
				filename:         "./test-storage",
				synchronousFlush: false,
			},
			metric: metric{
				Name:  "test_counter",
				Type:  "gauge",
				Value: "1",
			},
			want: &metric{
				Name:  "test_counter",
				Type:  "gauge",
				Value: "1",
			},
			wantErr: false,
		},
		{
			name: "successful test: counter",
			args: args{
				filename:         "./test-storage",
				synchronousFlush: false,
			},
			metric: metric{
				Name:  "test_counter2",
				Type:  "counter",
				Value: "1",
			},
			want: &metric{
				Name:  "test_counter2",
				Type:  "counter",
				Value: "1",
			},
			wantErr: false,
		},
		{
			name: "successful test: gauge",
			args: args{
				filename:         "./test-storage",
				synchronousFlush: false,
			},
			metric: metric{
				Name:  "test_gauge",
				Type:  "gauge",
				Value: "1.123",
			},
			want: &metric{
				Name:  "test_gauge",
				Type:  "gauge",
				Value: "1.123",
			},
			wantErr: false,
		},
		{
			name: "successful test: switch gauge to counter",
			args: args{
				filename:         "./test-storage",
				synchronousFlush: false,
			},
			metric: metric{
				Name:  "test_gauge",
				Type:  "counter",
				Value: "1",
			},
			want: &metric{
				Name:  "test_gauge",
				Type:  "counter",
				Value: "1",
			},
			wantErr: false,
		},
		{
			name: "successful test: gauge",
			args: args{
				filename:         "./test-storage",
				synchronousFlush: false,
			},
			metric: metric{
				Name:  "test_gauge2",
				Type:  "gauge",
				Value: "1.123",
			},
			want: &metric{
				Name:  "test_gauge2",
				Type:  "gauge",
				Value: "1.123",
			},
			wantErr: false,
		},
		{
			name: "successful test: custom type",
			args: args{
				filename:         "./test-storage",
				synchronousFlush: false,
			},
			metric: metric{
				Name:  "test_custom",
				Type:  "gauge",
				Value: "1.123",
			},
			want: &metric{
				Name:  "test_custom",
				Type:  "gauge",
				Value: "1.123",
			},
			wantErr: false,
		},
		{
			name: "successful test: custom type 2",
			args: args{
				filename:         "./test-storage",
				synchronousFlush: false,
			},
			metric: metric{
				Name:  "test_custom",
				Type:  "gauge",
				Value: "kek",
			},
			want: &metric{
				Name:  "test_custom",
				Type:  "gauge",
				Value: "kek",
			},
			wantErr: true,
		},
		{
			name: "unsuccessful test: incorrect value",
			args: args{
				filename:         "./test-storage",
				synchronousFlush: false,
			},
			metric: metric{
				Name:  "test_gauge3",
				Type:  "gauge",
				Value: "1",
			},
			want: &metric{
				Name:  "test_gauge3",
				Type:  "gauge",
				Value: "2",
			},
			wantErr: true,
		},
		{
			name: "unsuccessful test: incorrect type",
			args: args{
				filename:         "./test-storage",
				synchronousFlush: false,
			},
			metric: metric{
				Name:  "test_gauge3",
				Type:  "gauge",
				Value: "1",
			},
			want: &metric{
				Name:  "test_gauge3",
				Type:  "counter",
				Value: "1",
			},
			wantErr: true,
		},
		{
			name: "unsuccessful test: incorrect metric name",
			args: args{
				filename:         "./test-storage",
				synchronousFlush: false,
			},
			metric: metric{
				Name:  "test_gauge3",
				Type:  "gauge",
				Value: "1",
			},
			want: &metric{
				Name:  "test_gauge4",
				Type:  "gauge",
				Value: "1",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		s, err := NewFileStorage(tt.args.filename, tt.args.synchronousFlush)
		require.NoError(t, err)
		t.Run(tt.name, func(t *testing.T) {
			m, err := models.NewMetric(tt.metric.Name, tt.metric.Type, tt.metric.Value)
			if !tt.wantErr {
				require.NoError(t, err)
			}
			err = s.SaveMetric(m)
			require.NoError(t, err)
			if !tt.wantErr {
				assert.Equal(t, tt.want.Name, m.ID)
				assert.Equal(t, tt.want.Type, m.MType)
				assert.Equal(t, tt.want.Value, m.String())
			} else {
				assert.NotEqual(t, tt.want, m)
			}
		})
	}
}

func TestFileStorage_FlushMetrics(t *testing.T) {
	type metric struct {
		Name  string
		Type  string
		Value any
	}
	type args struct {
		filename         string
		synchronousFlush bool
	}
	tests := []struct {
		name    string
		args    args
		metrics map[string]metric
		wantErr bool
	}{
		{
			name: "successful test",
			args: args{
				filename:         "./test-storage",
				synchronousFlush: false,
			},
			metrics: map[string]metric{
				"test_counter": {
					Name:  "test_counter",
					Type:  "counter",
					Value: 10,
				},

				"test_gauge": {
					Name:  "test_gauge",
					Type:  "gauge",
					Value: 1.123,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer os.RemoveAll(tt.args.filename)
			var err error

			s, err := NewFileStorage(tt.args.filename, tt.args.synchronousFlush)
			require.NoError(t, err)

			for _, m := range tt.metrics {
				s.Metrics[m.Name], err = models.NewMetric(m.Name, m.Type, m.Value)
				require.NoError(t, err)
			}

			err = s.FlushMetrics()
			assert.NoError(t, err)

			f, err := os.OpenFile(tt.args.filename, os.O_RDONLY, 0644)
			require.NoError(t, err)
			defer f.Close()

			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				var metric models.Metric
				err = json.Unmarshal(scanner.Bytes(), &metric)
				require.NoError(t, err)
				assert.Equal(t, s.Metrics[metric.ID], metric)
			}

			err = scanner.Err()
			require.NoError(t, err)

			err = os.RemoveAll(tt.args.filename)
			require.NoError(t, err)
		})
	}
}

func TestFileStorage_StartupRestore(t *testing.T) {
	type metric struct {
		Name  string
		Type  string
		Value any
	}
	type args struct {
		filename         string
		synchronousFlush bool
	}
	tests := []struct {
		name        string
		args        args
		fileContent string
		wantMetrics []metric
		wantErr     bool
	}{
		{
			name: "successful test",
			args: args{
				filename:         "./test-storage",
				synchronousFlush: false,
			},
			fileContent: `{"id":"test_counter","type":"counter","delta":10}
{"id":"test_gauge","type":"gauge","value":1.123}`,
			wantMetrics: []metric{
				{
					Name:  "test_counter",
					Type:  "counter",
					Value: 10,
				},
				{
					Name:  "test_gauge",
					Type:  "gauge",
					Value: 1.123,
				},
			},
			wantErr: false,
		},
		{
			name: "successful test",
			args: args{
				filename:         "./test-storage",
				synchronousFlush: false,
			},
			fileContent: `{"id":"test_counter","type":"counter","delta":10}
{"id":"test_gauge","type":"gauge","value":1.123}`,
			wantMetrics: []metric{
				{
					Name:  "test_counter2",
					Type:  "counter",
					Value: 10,
				},
				{
					Name:  "test_gauge3",
					Type:  "gauge",
					Value: 1.123,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.RemoveAll(tt.args.filename)
			require.NoError(t, err)
			defer os.RemoveAll(tt.args.filename)

			f, err := os.OpenFile(tt.args.filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			require.NoError(t, err)
			_, err = f.Write([]byte(tt.fileContent))
			require.NoError(t, err)
			err = f.Close()
			require.NoError(t, err)

			s, err := NewFileStorage(tt.args.filename, tt.args.synchronousFlush)
			require.NoError(t, err)
			err = s.StartupRestore()
			require.NoError(t, err)

			err = os.RemoveAll(tt.args.filename)
			require.NoError(t, err)

			for _, m := range tt.wantMetrics {
				if _, ok := s.Metrics[m.Name]; !ok {
					if !tt.wantErr {
						t.Errorf("Metric %s not found", m.Name)
					}
				}
			}
		})
	}
}
