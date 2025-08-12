package handler

import ("compress/gzip"
"net/http"
"strings")

func GzipMiddleware(next http.Handler) http.Handler {
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

