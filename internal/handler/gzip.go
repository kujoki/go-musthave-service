package handler

import ("compress/gzip"
"net/http"
"strings"
"io"
)

func GzipMiddlewareRequest(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") && r.Body != http.NoBody {
            gzReader, err := gzip.NewReader(r.Body)
            if err == nil {
                defer gzReader.Close()
                r.Body = gzReader
            }
        }
        next.ServeHTTP(w, r)
    })
}

func GzipMiddlewareResponse(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        contentType := r.Header.Get("Content-Type")
        if !strings.HasPrefix(contentType, "application/json") || !strings.HasPrefix(contentType, "text/html") {
            next.ServeHTTP(w, r)
            return
        }
        if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
            next.ServeHTTP(w, r)
            return
        }

        gz := gzip.NewWriter(w)
        defer gz.Close()

        gzWriter := &gzipResponseWriter{
            ResponseWriter: w,
            Writer:         gz,
        }
        next.ServeHTTP(gzWriter, r)
    })
}

type gzipResponseWriter struct {
    http.ResponseWriter
    Writer io.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
    contentType := w.Header().Get("Content-Type")
    if strings.HasPrefix(contentType, "application/json") || strings.HasPrefix(contentType, "text/html") {
        w.Header().Set("Content-Encoding", "gzip")
        return w.Writer.Write(b)
    }
    return w.ResponseWriter.Write(b)
}


