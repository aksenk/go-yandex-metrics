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

const SignHeader = "HashSHA256"

func GetSignature(data []byte, cryptKey string) string {
	h := hmac.New(sha256.New, []byte(cryptKey))
	h.Write(data)
	sign := h.Sum(nil)
	return hex.EncodeToString(sign[:])
}

func Middleware(cryptKey string, log *zap.SugaredLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

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

			// этого нет в задании, но, похоже, нужно игнорировать проверку подписи если заголовок Hash=none
			if r.Header.Get("Hash") != "none" {
				// проверяем хэш запроса для только для запросов с телом, для совместимости со старыми методами
				// когда добавляем метрики простым POST'ом без тела запроса (например /update/counter/kek/1)
				if len(body) > 0 {
					reqSignHeader := r.Header.Get(SignHeader)
					if reqSignHeader == "" {
						log.Errorf("Header HashSHA256 is empty")
						w.Header().Set("Content-Type", "application/json")
						respErr := `{"error": "Header HashSHA256 is empty"}`
						w.Write([]byte(respErr))
						w.WriteHeader(http.StatusBadRequest)
						return
					}

					sign := GetSignature(body, cryptKey)

					if sign != reqSignHeader {
						log.Errorf("Request signature is not valid")
						w.Header().Set("Content-Type", "application/json")
						respErr := `{"error": "Request signature is not valid"}`
						w.Write([]byte(respErr))
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					log.Infof("Request signature is valid")

					w.Header().Set(SignHeader, sign)
				}
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
