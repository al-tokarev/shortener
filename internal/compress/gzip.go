package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/al-tokarev/shortener/internal/logger"
)

type compressWriter struct {
	w          http.ResponseWriter
	gzipWriter *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:          w,
		gzipWriter: gzip.NewWriter(w),
	}
}

func (cw *compressWriter) Header() http.Header {
	return cw.w.Header()
}

func (cw *compressWriter) Write(p []byte) (int, error) {
	return cw.gzipWriter.Write(p)
}

func (cw *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		cw.w.Header().Set("Content-Encoding", "gzip")
	}
	cw.w.WriteHeader(statusCode)
}

func (cw *compressWriter) Close() error {
	return cw.gzipWriter.Close()
}

type compressReader struct {
	r          io.ReadCloser
	gzipReader *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:          r,
		gzipReader: gzipReader,
	}, nil
}

func (cr compressReader) Read(p []byte) (n int, err error) {
	return cr.gzipReader.Read(p)
}

func (cr *compressReader) Close() error {
	if err := cr.r.Close(); err != nil {
		return err
	}
	return cr.gzipReader.Close()
}

var availableTypes = map[string]bool{
	"application/json": true,
	"text/html":        true,
}

func GzipMiddleware(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		new_resp_wr := w

		contentType := r.Header.Get("Content-Type")
		mainType := strings.Split(contentType, ";")[0]

		if availableTypes[mainType] {
			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportGzip := strings.Contains(acceptEncoding, "gzip")
			if supportGzip {
				cw := newCompressWriter(w)
				new_resp_wr = cw
				defer cw.Close()
			}

			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				cr, err := newCompressReader(r.Body)
				if err != nil {
					logger.Sugar.Debug("Error create compress reader")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				r.Body = cr
				defer cr.Close()
			}
		}
		h.ServeHTTP(new_resp_wr, r)
	})
}
