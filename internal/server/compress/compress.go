package compress

import (
	"compress/gzip"
	"github.com/aksenk/go-yandex-metrics/internal/server/logger"
	"io"
	"net/http"
	"strings"
)

type gzipReadCloser struct {
	io.ReadCloser
	gr io.Reader
}

func (g gzipReadCloser) Read(b []byte) (int, error) {
	return g.gr.Read(b)
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gzw io.Writer
}

func (g gzipResponseWriter) Write(b []byte) (int, error) {
	return g.gzw.Write(b)
}

//func NewGzipResponseWriter(w http.ResponseWriter) *gzipResponseWriter {
//	return &gzipResponseWriter{
//		w,
//		gzip.NewWriter(w),
//	}
//}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var nextResponseWriter http.ResponseWriter
		if strings.Contains(request.Header.Get("Content-Encoding"), "gzip") {
			logger.Log.Debugf("Using gzipped request")
			gr, err := gzip.NewReader(request.Body)
			if err != nil {
				// TODO
				panic(err)
			}
			grc := gzipReadCloser{
				request.Body,
				gr,
			}
			request.Body = grc
		}

		if strings.Contains(request.Header.Get("Accept-Encoding"), "gzip") {
			logger.Log.Debugf("Using gzipped response")
			// TODO проверять и сжимать только определенные content type
			// TODO отказаться от gzip.NewWriterLevel() и использовать метод gzip.Reset()
			gzw, _ := gzip.NewWriterLevel(writer, gzip.BestSpeed)
			defer gzw.Close()
			nextResponseWriter = gzipResponseWriter{
				writer,
				gzw,
			}
			nextResponseWriter.Header().Set("Content-Encoding", "gzip")
		} else {
			nextResponseWriter = writer
		}

		next.ServeHTTP(nextResponseWriter, request)
	})
}
