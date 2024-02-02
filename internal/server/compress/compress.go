package compress

import (
	"compress/gzip"
	"github.com/aksenk/go-yandex-metrics/internal/server/logger"
	"io"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	gzw io.Writer
}

func (g gzipResponseWriter) Write(b []byte) (int, error) {
	return g.gzw.Write(b)
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

//func NewGzipResponseWriter(w http.ResponseWriter) *gzipResponseWriter {
//	return &gzipResponseWriter{
//		w,
//		gzip.NewWriter(w),
//	}
//}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !strings.Contains(request.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(writer, request)
			return
		}
		logger.Log.Debugf("Using gzipped response")
		// TODO проверять и сжимать только определенные content type
		// TODO отказаться от gzip.NewWriterLevel() и использовать метод gzip.Reset()
		gzw, _ := gzip.NewWriterLevel(writer, gzip.BestSpeed)
		defer gzw.Close()
		grw := gzipResponseWriter{
			writer,
			gzw,
		}
		grw.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(grw, request)
	})
}
