package metrics

import (
	"fmt"
	"net/http"
	"runtime"
)

type Server struct {
	addr string
}

func NewServer(addr string) *Server {
	return &Server{addr: addr}
}

type metricsHandler struct{}

func (h metricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.WriteHeader(http.StatusOK)

	_, _ = fmt.Fprintf(w, "# HELP go_goroutines Number of goroutines\n")
	_, _ = fmt.Fprintf(w, "# TYPE go_goroutines gauge\n")
	_, _ = fmt.Fprintf(w, "go_goroutines %d\n", runtime.NumGoroutine())

	_, _ = fmt.Fprintf(w, "# HELP go_memory_alloc_bytes Bytes allocated by the heap\n")
	_, _ = fmt.Fprintf(w, "# TYPE go_memory_alloc_bytes gauge\n")
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	_, _ = fmt.Fprintf(w, "go_memory_alloc_bytes %d\n", ms.Alloc)
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.Handle("/metrics", metricsHandler{})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "ok")
	})

	return http.ListenAndServe(s.addr, mux)
}
