package signature

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"go.uber.org/zap"
	"io"
	"net/http"
)

func Middleware(cryptKey string, log *zap.SugaredLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			reqSignHeader := r.Header.Get("HashSHA256")
			if reqSignHeader == "" {
				log.Errorf("Header HashSHA256 is empty")
				http.Error(w, "Header HashSHA256 is empty", http.StatusBadRequest)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Errorf("Error reading request body: %v", err)
				http.Error(w, "error reading body", http.StatusInternalServerError)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(body)) // reset body to its original state

			err = r.Body.Close()
			if err != nil {
				log.Errorf("Error closing request body: %v", err)
				http.Error(w, "error closing body", http.StatusInternalServerError)
				return
			}

			h := hmac.New(sha256.New, []byte(cryptKey))
			h.Write(body)
			sign := h.Sum(nil)

			strSign := hex.EncodeToString(sign[:])
			if strSign != reqSignHeader {
				log.Errorf("Request signature is not valid")
				http.Error(w, "Request signature is not valid", http.StatusBadRequest)
				return
			}
			log.Infof("Request signature is valid")
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
