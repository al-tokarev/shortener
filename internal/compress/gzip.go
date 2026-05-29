package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

var availableTypes = map[string]bool{
	"application/json": true,
	"text/html":        true,
}

type compressWriter struct {
	w          http.ResponseWriter
	gzipWriter *gzip.Writer
	statusCode int
	compress   bool
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:          w,
		gzipWriter: nil,
		statusCode: http.StatusOK,
		compress:   false,
	}
}

func (cw *compressWriter) Header() http.Header {
	return cw.w.Header()
}

func (cw *compressWriter) Write(p []byte) (int, error) {
	if !cw.compress {
		return cw.w.Write(p)
	}
	return cw.gzipWriter.Write(p)
}

func (cw *compressWriter) WriteHeader(statusCode int) {
	cw.statusCode = statusCode

	contentType := cw.Header().Get("Content-Type")
	mainType := strings.Split(contentType, ";")[0]

	if statusCode >= 200 && statusCode < 300 && availableTypes[mainType] {
		cw.w.Header().Set("Content-Encoding", "gzip")
		cw.gzipWriter = gzip.NewWriter(cw.w)
		cw.compress = true
	}
	cw.w.WriteHeader(statusCode)
}

func (cw *compressWriter) Close() error {
	if cw.gzipWriter != nil {
		return cw.gzipWriter.Close()
	}
	return nil
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

func GzipMiddleware(logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger = logger.With(zap.String("component", "Gzip middleware"))
			newRespWr := w

			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportGzip := strings.Contains(acceptEncoding, "gzip")
			if supportGzip {
				cw := newCompressWriter(w)
				newRespWr = cw
				defer cw.Close()
			}

			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				cr, err := newCompressReader(r.Body)
				if err != nil {
					logger.Debug("Error create compress reader")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				r.Body = cr
				defer cr.Close()
			}

			h.ServeHTTP(newRespWr, r)
		})
	}
}
