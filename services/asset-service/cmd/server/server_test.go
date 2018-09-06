package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"syscall"
	"testing"
	"time"

	"github.com/moorara/microservices-demo/services/asset-service/pkg/log"
	"github.com/moorara/microservices-demo/services/asset-service/pkg/metrics"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
)

// mockHTTPServer is a mock implementation of HTTPServer
type mockHTTPServer struct {
	ListenAndServeCalled   bool
	ListenAndServeOutError error

	ShutdownCalled    bool
	ShutdownInContext context.Context
	ShutdownOutError  error
}

func (m *mockHTTPServer) ListenAndServe() error {
	m.ListenAndServeCalled = true
	return m.ListenAndServeOutError
}

func (m *mockHTTPServer) Shutdown(ctx context.Context) error {
	m.ShutdownCalled = true
	m.ShutdownInContext = ctx
	return m.ShutdownOutError
}

func TestNotFound(t *testing.T) {
	tests := []struct {
		port           string
		method         string
		url            string
		expectedStatus int
	}{
		{
			port:           ":9999",
			method:         "GET",
			url:            "/invalid",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		logger := log.NewNopLogger()
		metrics := metrics.New("test-service")
		tracer := mocktracer.New()
		server := New(tc.port, logger, metrics, tracer)

		r := httptest.NewRequest(tc.method, tc.url, nil)
		w := httptest.NewRecorder()
		server.notFound(w, r)

		assert.Equal(t, tc.expectedStatus, w.Result().StatusCode)
	}
}

func TestLiveness(t *testing.T) {
	tests := []struct {
		port           string
		method         string
		url            string
		expectedStatus int
	}{
		{
			port:           ":9999",
			method:         "GET",
			url:            "/liveness",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		logger := log.NewNopLogger()
		metrics := metrics.New("test-service")
		tracer := mocktracer.New()
		server := New(tc.port, logger, metrics, tracer)

		r := httptest.NewRequest(tc.method, tc.url, nil)
		w := httptest.NewRecorder()
		server.liveness(w, r)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	}
}

func TestReadiness(t *testing.T) {
	tests := []struct {
		port           string
		method         string
		url            string
		expectedStatus int
	}{
		{
			port:           ":9999",
			method:         "GET",
			url:            "/readiness",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		logger := log.NewNopLogger()
		metrics := metrics.New("test-service")
		tracer := mocktracer.New()
		server := New(tc.port, logger, metrics, tracer)

		r := httptest.NewRequest(tc.method, tc.url, nil)
		w := httptest.NewRecorder()
		server.readiness(w, r)

		assert.Equal(t, tc.expectedStatus, w.Result().StatusCode)
	}
}

func TestStart(t *testing.T) {
	tests := []struct {
		name          string
		signal        syscall.Signal
		httpServer    *mockHTTPServer
		expectedError error
	}{
		{
			"IntSignal",
			syscall.SIGINT,
			&mockHTTPServer{},
			errors.New("interrupt"),
		},
		{
			"TermSignal",
			syscall.SIGTERM,
			&mockHTTPServer{},
			errors.New("terminated"),
		},
		{
			"HTTPServerError",
			0,
			&mockHTTPServer{
				ListenAndServeOutError: errors.New("server error"),
			},
			errors.New("server error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger := log.NewNopLogger()
			server := &Server{
				logger:     logger,
				httpServer: tc.httpServer,
			}

			if tc.signal > 0 {
				sig := tc.signal // prevent data race
				go func() {
					time.Sleep(50 * time.Millisecond)
					syscall.Kill(syscall.Getpid(), sig)
				}()
			}

			err := server.Start()
			assert.IsType(t, tc.expectedError, err)
		})
	}
}

func TestStop(t *testing.T) {
	tests := []struct {
		name       string
		httpServer *mockHTTPServer
	}{
		{
			"HTTPServerError",
			&mockHTTPServer{
				ShutdownOutError: errors.New("server error"),
			},
		},
		{
			"NoError",
			&mockHTTPServer{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger := log.NewNopLogger()
			server := &Server{
				logger:     logger,
				httpServer: tc.httpServer,
			}

			server.Stop()
			assert.True(t, tc.httpServer.ShutdownCalled)
		})
	}
}