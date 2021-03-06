package log

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockLogger struct {
	LogPassedKV []interface{}
	LogError    error
}

func (m *mockLogger) Log(kv ...interface{}) error {
	m.LogPassedKV = kv
	return m.LogError
}

func TestNewVoidLogger(t *testing.T) {
	logger := NewVoidLogger()
	assert.NotNil(t, logger)
}

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		loggerName  string
		logLevel    string
	}{
		{"LevelDebug", "go-service", "handler", "debug"},
		{"LevelInfo", "grpc-service", "handler", "info"},
		{"LevelWarn", "auth-service", "handler", "warn"},
		{"LevelError", "tenant-service", "handler", "error"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			l := NewLogger(tc.serviceName, tc.loggerName, tc.logLevel)
			assert.NotNil(t, l)
		})
	}
}

func TestLogger(t *testing.T) {
	tests := []struct {
		name       string
		mockLogger mockLogger
		kv         []interface{}
		expectedKV []interface{}
	}{
		{
			"OK",
			mockLogger{},
			[]interface{}{"key", "value", "message", "content"},
			[]interface{}{"key", "value", "message", "content"},
		},
		{
			"Error",
			mockLogger{
				LogError: errors.New("error"),
			},
			[]interface{}{"key", "value", "message", "content"},
			[]interface{}{"key", "value", "message", "content"},
		},
	}

	for _, tc := range tests {
		t.Run("Debug"+tc.name, func(t *testing.T) {
			l := &Logger{Logger: &tc.mockLogger}
			l.Debug(tc.kv)
			assert.Equal(t, tc.expectedKV, tc.mockLogger.LogPassedKV[2])
		})

		t.Run("Info"+tc.name, func(t *testing.T) {
			l := &Logger{Logger: &tc.mockLogger}
			l.Info(tc.kv)
			assert.Equal(t, tc.expectedKV, tc.mockLogger.LogPassedKV[2])
		})

		t.Run("Warn"+tc.name, func(t *testing.T) {
			l := &Logger{Logger: &tc.mockLogger}
			l.Warn(tc.kv)
			assert.Equal(t, tc.expectedKV, tc.mockLogger.LogPassedKV[2])
		})

		t.Run("Error"+tc.name, func(t *testing.T) {
			l := &Logger{Logger: &tc.mockLogger}
			l.Error(tc.kv)
			assert.Equal(t, tc.expectedKV, tc.mockLogger.LogPassedKV[2])
		})
	}
}
