package logger

import (
	"bytes"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

var Log *zap.SugaredLogger

type responseData struct {
	statusCode int
	size       int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func init() {
	atom := zap.NewAtomicLevel()
	atom.SetLevel(zap.InfoLevel)
	cfg := zap.NewProductionConfig()
	cfg.Level = atom
	zl, _ := cfg.Build()
	Log = zl.Sugar()
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size = size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.statusCode = statusCode
}

func NewLogger(level string) (*zap.SugaredLogger, error) {
	atom := zap.NewAtomicLevel()

	switch level {
	case "debug":
		atom.SetLevel(zap.DebugLevel)
	case "info":
		atom.SetLevel(zap.InfoLevel)
	case "warn":
		atom.SetLevel(zap.WarnLevel)
	case "error":
		atom.SetLevel(zap.ErrorLevel)
	default:
		return nil, fmt.Errorf("unknown log level: %s", level)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = atom
	zl, _ := cfg.Build()
	return zl.Sugar(), nil
}

func Middleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		logger, err := NewLogger("debug")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		start := time.Now()
		uri := r.RequestURI
		method := r.Method

		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Errorf("Error reading request body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(body)) // reset body to its original state

		lrw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   &responseData{},
		}
		next.ServeHTTP(&lrw, r)
		duration := time.Since(start)

		logger.Infof("Request URI=%v method=%v duration=%v headers=%v body=%v", uri, method, duration, r.Header, string(body))
		logger.Infof("Response statusCode=%v size=%v", lrw.responseData.statusCode, lrw.responseData.size)
	}
	return http.HandlerFunc(fn)
}
