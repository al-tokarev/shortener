package logger

import (
	"log"
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

func NewLogger() (*zap.SugaredLogger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	log.Println("Добавлен логгер")

	return logger.Sugar(), nil
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

func WithLogging(logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			timeNow := time.Now()

			uri := r.RequestURI
			method := r.Method

			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData{http.StatusOK, 0},
			}

			h.ServeHTTP(&lw, r)

			duration := time.Since(timeNow)
			logger.Infow("Request is finish",
				"uri", uri,
				"method", method,
				"duration", duration,
				"response size", lw.responseData.size,
				"status code", lw.responseData.status,
			)
		})
	}
}
