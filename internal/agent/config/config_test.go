package config

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	var want = map[string]string{
		"ServerUseHTTPS": "true",
		"ServerURL":      "https://1.1.1.1:111/updates",
		"PollInterval":   "1",
		"ReportInterval": "11",
	}
	var overrideEnv = map[string]string{
		"USE_HTTPS":       "true",
		"ADDRESS":         "1.1.1.1:111",
		"POLL_INTERVAL":   "1",
		"REPORT_INTERVAL": "11",
	}
	t.Run("test NewConfig()", func(t *testing.T) {
		var gotConfig interface{}
		var err error
		for k, v := range overrideEnv {
			t.Setenv(k, v)
		}
		gotConfig, err = NewConfig()
		require.NoError(t, err)
		if _, ok := gotConfig.(*Config); !ok {
			t.Fatalf("Resulting object have incorrect type (not equal Config struct)")
		}
		ServerUseHTTPSBool, err := strconv.ParseBool(want["ServerUseHTTPS"])
		require.NoError(t, err)
		PollIntervalInt, err := strconv.Atoi(want["PollInterval"])
		require.NoError(t, err)
		ReportIntervalInt, err := strconv.Atoi(want["ReportInterval"])
		require.NoError(t, err)

		assert.Equal(t, ServerUseHTTPSBool, gotConfig.(*Config).ServerUseHTTPS)
		assert.Equal(t, want["ServerURL"], gotConfig.(*Config).ServerURL)
		assert.Equal(t, time.Second*time.Duration(PollIntervalInt), gotConfig.(*Config).PollInterval)
		assert.Equal(t, time.Second*time.Duration(ReportIntervalInt), gotConfig.(*Config).ReportInterval)
	})
}
