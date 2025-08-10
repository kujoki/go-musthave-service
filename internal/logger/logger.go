package logger

import (
	"net/http"
	"go.uber.org/zap"
	"time"
)

type (
    responseData struct {
        status int
        size int
    }

    loggingResponseWriter struct {
        http.ResponseWriter
        responseData *responseData
    }
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
    size, err := r.ResponseWriter.Write(b) 
    r.responseData.size += size
    return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
    r.ResponseWriter.WriteHeader(statusCode) 
    r.responseData.status = statusCode
} 

func WithLogging(sugar zap.SugaredLogger, h http.Handler) http.HandlerFunc {
    logFn := func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

		responseData := &responseData {
            status: 0,
            size: 0,
        }
		lw := loggingResponseWriter {
            ResponseWriter: w,
            responseData: responseData,
        }

		h.ServeHTTP(&lw, r)

        duration := time.Since(start)

        sugar.Infoln(
            "uri", r.RequestURI,
            "method", r.Method,
            "status", responseData.status,
            "duration", duration,
            "size", responseData.size,
        )
	}
	return http.HandlerFunc(logFn)
}