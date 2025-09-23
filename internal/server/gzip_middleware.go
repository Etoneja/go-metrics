package server

import (
	"compress/gzip"
	"fmt"
	"mime"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

var compressibleTypes = map[string]bool{
	"application/json": true,
	"text/html":        true,
}

func isCompressibleContentType(contentType string) bool {
	mimeType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	res, ok := compressibleTypes[mimeType]
	if !ok {
		res = false
	}
	return res
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gz       *gzip.Writer
	compress bool
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	contentType := w.Header().Get("Content-Type")
	if isCompressibleContentType(contentType) {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length")
		w.compress = true
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if w.compress {
		if w.gz == nil {
			gz, err := gzip.NewWriterLevel(w.ResponseWriter, gzip.BestSpeed)
			if err != nil {
				return 0, err
			}
			w.gz = gz
		}
		return w.gz.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

func (w *gzipResponseWriter) Close() error {
	if w.gz != nil {
		err := w.gz.Close()
		if err != nil {
			return fmt.Errorf("failed to close gzip writer: %w", err)
		}
	}
	return nil
}

func (bmw *BaseMiddleware) GzipMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
				gz, err := gzip.NewReader(r.Body)
				if err != nil {
					http.Error(w, "Invalid gzip body", http.StatusBadRequest)
					return
				}
				defer func() {
					if err := gz.Close(); err != nil {
						bmw.logger.Warn("failed to close gzip reader", zap.Error(err))
					}
				}()

				r.Body = gz
			}

			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			gzw := gzipResponseWriter{ResponseWriter: w}
			defer func() {
				if err := gzw.Close(); err != nil {
					bmw.logger.Warn("failed to close gzip response writer", zap.Error(err))
				}
			}()

			next.ServeHTTP(&gzw, r)
		}

		return http.HandlerFunc(fn)
	}
}
