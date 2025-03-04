package main

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/klauspost/compress/zstd"
)

var gzPool = sync.Pool{
	New: func() interface{} {
		w := gzip.NewWriter(io.Discard)
		return w
	},
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) WriteHeader(status int) {
	w.Header().Del("Content-Length")
	w.ResponseWriter.WriteHeader(status)
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func GzipCompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")

		gz := gzPool.Get().(*gzip.Writer)
		defer gzPool.Put(gz)

		gz.Reset(w)
		defer gz.Close()

		next.ServeHTTP(&gzipResponseWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

func ZstdCompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acceptEncoding := r.Header.Get("Accept-Encoding")
		if !strings.Contains(acceptEncoding, "zstd") {
			next.ServeHTTP(w, r)
			return
		}

		zstdWriter, err := zstd.NewWriter(w)
		if err != nil {
			http.Error(w, "Failed to create zstd writer", http.StatusInternalServerError)
			return
		}
		defer zstdWriter.Close()

		w.Header().Set("Content-Encoding", "zstd")
		w.Header().Del("Content-Length") // 删除 Content-Length，因为压缩后的长度未知

		zw := &zstdResponseWriter{ResponseWriter: w, Writer: zstdWriter}
		next.ServeHTTP(zw, r)
	})
}

type zstdResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (zw *zstdResponseWriter) Write(b []byte) (int, error) {
	return zw.Writer.Write(b)
}

func (zw *zstdResponseWriter) WriteHeader(statusCode int) {
	zw.ResponseWriter.WriteHeader(statusCode)
}

func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.String())
		next.ServeHTTP(w, r)
	})
}
