package server

import (
	"bytes"
	"io"
	"net/http"

	"github.com/etoneja/go-metrics/internal/common"
	"go.uber.org/zap"
)

type responseRecorder struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func (bmw *BaseMiddleware) HashMiddleware(hashKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			if r.Body == nil || r.ContentLength == 0 {
				next.ServeHTTP(w, r)
				return
			}

			requestHash := r.Header.Get(common.HashHeaderKey)

			if hashKey == "" || requestHash == "" {
				next.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				bmw.logger.Error("failed to read body",
					zap.Error(err),
				)
    			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(body))

			expectedHash := common.СomputeHash(hashKey, body)
			if !common.CompareHashes(requestHash, expectedHash) {
				http.Error(w, "Invalid request hash", http.StatusBadRequest)
				return
			}

			recorder := &responseRecorder{
				ResponseWriter: w,
				body:           &bytes.Buffer{},
			}

			next.ServeHTTP(recorder, r)

			if recorder.body.Len() > 0 {
				hash := common.СomputeHash(hashKey, recorder.body.Bytes())
				w.Header().Set(common.HashHeaderKey, hash)
			}
		}

		return http.HandlerFunc(fn)
	}
}
