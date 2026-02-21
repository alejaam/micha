package httpadapter

import (
	"net/http"
)

type Server struct {
	port string
	mux  *http.ServeMux
}

func NewServer(port string) Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	return Server{port: port, mux: mux}
}

func (s Server) Start() error {
	return http.ListenAndServe(":"+s.port, s.mux)
}
