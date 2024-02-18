package logger

import (
	"fmt"
	"go.uber.org/zap"
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

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size = size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.statusCode = statusCode
}

func init() {
	atom := zap.NewAtomicLevel()
	//atom.SetLevel(zap.InfoLevel)
	// TODO убрать
	atom.SetLevel(zap.DebugLevel)
	cfg := zap.NewProductionConfig()
	cfg.Level = atom
	zl, _ := cfg.Build()
	Log = zl.Sugar()
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
		start := time.Now()
		uri := r.RequestURI
		method := r.Method
		lrw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   &responseData{},
		}
		next.ServeHTTP(&lrw, r)
		duration := time.Since(start)
		Log.Infof("Request URI=%v method=%v duration=%v", uri, method, duration)
		Log.Infof("Response statusCode=%v size=%v", lrw.responseData.statusCode, lrw.responseData.size)
	}
	return http.HandlerFunc(fn)
}
