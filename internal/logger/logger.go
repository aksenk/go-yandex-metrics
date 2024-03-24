package logger

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

type contextKey string

var Log *zap.SugaredLogger

const KeyLogger contextKey = "logger"

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
	cfg := zap.NewProductionConfig()
	cfg.DisableStacktrace = true

	switch level {
	case "debug":
		atom.SetLevel(zap.DebugLevel)
		//cfg.DisableStacktrace = false
	case "info":
		atom.SetLevel(zap.InfoLevel)
	case "warn":
		atom.SetLevel(zap.WarnLevel)
	case "error":
		atom.SetLevel(zap.ErrorLevel)
	default:
		return nil, fmt.Errorf("unknown log level: %s", level)
	}

	cfg.Level = atom
	zl, _ := cfg.Build()
	return zl.Sugar(), nil
}

func Middleware(log *zap.SugaredLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			uri := r.RequestURI
			method := r.Method

			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Errorf("Error reading request body: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(body)) // reset body to its original state
			r.Body.Close()

			if r.Header.Get("Content-Encoding") == "gzip" {
				read, err := gzip.NewReader(bytes.NewReader(body))
				if err != nil {
					log.Errorf("Error reading gzip body: %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				ungz, err := io.ReadAll(read)
				if err != nil {
					log.Errorf("Error reading gzip body: %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				body = ungz
			}

			lrw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   &responseData{},
			}

			ctx := context.WithValue(r.Context(), KeyLogger, log)

			log.With("URI", uri,
				"method", method,
				"headers", r.Header,
				"body", string(body),
				"request_id", middleware.GetReqID(r.Context())).
				Info("Received request")

			next.ServeHTTP(&lrw, r.WithContext(ctx))

			duration := time.Since(start)

			log.With("statusCode", lrw.responseData.statusCode,
				"size", lrw.responseData.size,
				"duration", duration,
				"headers", lrw.Header(),
				"request_id", middleware.GetReqID(r.Context())).
				Info("Sent response")
		}
		return http.HandlerFunc(fn)
	}
}

func FromContext(ctx context.Context) (*zap.SugaredLogger, error) {
	if ctx == nil {
		return nil, fmt.Errorf("nil context")
	}
	logger := ctx.Value(KeyLogger)
	if logger == nil {
		log, err := NewLogger("info")
		if err != nil {
			return nil, err
		}
		return log, nil
	}
	if _, ok := logger.(*zap.SugaredLogger); !ok {
		return nil, fmt.Errorf("logger in context is not a *zap.SugaredLogger")
	}
	return logger.(*zap.SugaredLogger), nil
}
