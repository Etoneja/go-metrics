package server

import (
	"bytes"
	"compress/gzip"
	"log"
	"net/http"
	"strings"
)

type bufferedResponseWriter struct {
	http.ResponseWriter
	buf         *bytes.Buffer
	statusCode  int
	wroteHeader bool
	headers     http.Header
}

func (w *bufferedResponseWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.buf.Write(data)
}

func (w *bufferedResponseWriter) WriteHeader(statusCode int) {
	if !w.wroteHeader {
		w.statusCode = statusCode
		w.wroteHeader = true
	}
}

func (w *bufferedResponseWriter) writeHeaders() {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	for key, values := range w.headers {
		for _, value := range values {
			w.ResponseWriter.Header().Set(key, value)
		}
	}

	w.ResponseWriter.WriteHeader(w.statusCode)
}

func (w *bufferedResponseWriter) Flush() {
	w.writeHeaders()
	if w.buf.Len() > 0 {
		w.ResponseWriter.Write(w.buf.Bytes())
	}
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Invalid gzip body", http.StatusBadRequest)
				return
			}
			defer gz.Close()
			r.Body = gz
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		if w.Header().Get("Content-Encoding") == "gzip" {
			next.ServeHTTP(w, r)
			return
		}

		var buf bytes.Buffer
		bw := &bufferedResponseWriter{
			ResponseWriter: w,
			buf:            &buf,
		}

		bw.Header().Set("Vary", "Accept-Encoding")

		next.ServeHTTP(bw, r)

		contentType := bw.Header().Get("Content-Type")
		if !isCompressibleContentType(contentType) {
			bw.Flush()
			return
		}

		// if buf.Len() < 1400 {
		//     bw.Flush()
		//     return
		// }

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length")

		bw.writeHeaders()

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			log.Printf("Failed to create gzip writer: %v", err)
			bw.Flush()
			return
		}
		defer gz.Close()

		_, err = gz.Write(buf.Bytes())
		if err != nil {
			log.Printf("Gzip write error: %v", err)
			http.Error(w, "Compression failed", http.StatusInternalServerError)
			return
		}
	})
}
