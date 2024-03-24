package compress

import (
	"compress/gzip"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"io"
	"net/http"
	"strings"
)

type gzipReadCloser struct {
	io.ReadCloser
	gzipReader io.Reader
}

func (g gzipReadCloser) Read(b []byte) (int, error) {
	return g.gzipReader.Read(b)
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gzipWriter io.Writer
}

func (g gzipResponseWriter) Write(b []byte) (int, error) {
	return g.gzipWriter.Write(b)
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		log := logger.Log
		var nextResponseWriter http.ResponseWriter
		if strings.Contains(request.Header.Get("Content-Encoding"), "gzip") {
			gr, err := gzip.NewReader(request.Body)
			if err != nil {
				log.Errorf("Can not ungzip incoming reeuest data: %v", err)
				http.Error(writer, "can not ungzip request data", http.StatusInternalServerError)
				return
			}
			grc := gzipReadCloser{
				request.Body,
				gr,
			}
			request.Body = grc
		}

		if strings.Contains(request.Header.Get("Accept-Encoding"), "gzip") {
			allowedGzipContentTypes := []string{"application/json", "text/html", ""}
			requestContentType := request.Header.Get("Content-Type")
			isAllowedContentType := false
			for _, ct := range allowedGzipContentTypes {
				if requestContentType == ct {
					isAllowedContentType = true
				}
			}
			if isAllowedContentType {
				log.Debug("The client requested compressed data. The response will be gzipped")
				// TODO отказаться от gzip.NewWriterLevel() и использовать метод gzip.Reset()
				gzw, _ := gzip.NewWriterLevel(writer, gzip.BestSpeed)
				defer gzw.Close()
				nextResponseWriter = gzipResponseWriter{
					writer,
					gzw,
				}
				nextResponseWriter.Header().Set("Content-Encoding", "gzip")
			} else {
				log.Debugf("Header 'Accept-Encoding: gzip' is exist, but this content-type is not allowed for gzipping: %v", requestContentType)
				nextResponseWriter = writer
			}
		} else {
			nextResponseWriter = writer
		}
		next.ServeHTTP(nextResponseWriter, request)
	})
}
