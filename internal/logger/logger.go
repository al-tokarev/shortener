package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData responseData
}

var Sugar zap.SugaredLogger

func Initialize() error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}

	Sugar = *logger.Sugar()
	return nil
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func WithLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeNow := time.Now()

		uri := r.RequestURI
		method := r.Method

		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData{0, 0},
		}

		Sugar.Infow("=== WithLogging BEFORE handler ===",
			"method", method,
			"uri", uri,
			"headers", r.Header)

		h.ServeHTTP(&lw, r)

		Sugar.Infow("=== WithLogging AFTER handler ===",
			"method", method,
			"uri", uri,
			"response_headers", lw.Header(),
			"status_code", lw.responseData.status)

		duration := time.Since(timeNow)
		Sugar.Infow("Request is finish",
			"uri", uri,
			"method", method,
			"duration", duration,
			"response size", lw.responseData.size,
			"status code", lw.responseData.status,
		)
	})
}
