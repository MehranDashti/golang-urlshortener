package metrics

import (
    "fmt"
    "net/http"
    "runtime"
)

// Server is a minimal HTTP server that demonstrates net/http
// without any framework. Gin is NOT used here.
//
// This is what Gin wraps under the hood:
//   - http.Handler interface (ServeHTTP method)
//   - http.HandlerFunc (function → Handler adapter)
//   - http.ServeMux (basic router)
type Server struct {
    addr string
}

func NewServer(addr string) *Server {
    return &Server{addr: addr}
}

// metricsHandler is a plain struct that implements http.Handler directly.
// Compare with Gin: gin.HandlerFunc is the equivalent of http.HandlerFunc,
// and gin.Engine.ServeHTTP is the equivalent of this ServeHTTP method.
type metricsHandler struct{}

func (h metricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // http.ResponseWriter.Header() must be set BEFORE WriteHeader()
    w.Header().Set("Content-Type", "text/plain; version=0.0.4")
    w.WriteHeader(http.StatusOK)

    // Write response body — plain text Prometheus-style metrics
    fmt.Fprintf(w, "# HELP go_goroutines Number of goroutines\n")
    fmt.Fprintf(w, "# TYPE go_goroutines gauge\n")
    fmt.Fprintf(w, "go_goroutines %d\n", runtime.NumGoroutine())

    fmt.Fprintf(w, "# HELP go_memory_alloc_bytes Bytes allocated by the heap\n")
    fmt.Fprintf(w, "# TYPE go_memory_alloc_bytes gauge\n")
    var ms runtime.MemStats
    runtime.ReadMemStats(&ms)
    fmt.Fprintf(w, "go_memory_alloc_bytes %d\n", ms.Alloc)
}

// Start registers routes on a plain http.ServeMux and starts the server.
// http.ServeMux is the stdlib router — the equivalent of Gin's Engine for routing.
func (s *Server) Start() error {
    mux := http.NewServeMux()

    // http.HandlerFunc converts a plain func into an http.Handler
    // This is what gin.HandlerFunc does under the hood
    mux.Handle("/metrics", metricsHandler{})
    mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        fmt.Fprint(w, "ok")
    })

    return http.ListenAndServe(s.addr, mux)
}