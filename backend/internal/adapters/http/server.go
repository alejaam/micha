package httpadapter

import (
	"net/http"
)

// ServerDependencies groups all resource-level dependencies for the HTTP server.
type ServerDependencies struct {
	Expense ExpenseHandlerDeps
}

// Server is the primary HTTP adapter.
type Server struct {
	port string
	mux  *http.ServeMux
}

// NewServer constructs a Server and registers all routes.
func NewServer(port string, deps ServerDependencies) Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	eh := newExpenseHandler(deps.Expense)
	mux.HandleFunc("POST /v1/expenses", eh.handleCreate)
	mux.HandleFunc("GET /v1/expenses/{id}", eh.handleGet)
	mux.HandleFunc("GET /v1/expenses", eh.handleList)
	mux.HandleFunc("PATCH /v1/expenses/{id}", eh.handlePatch)
	mux.HandleFunc("DELETE /v1/expenses/{id}", eh.handleDelete)

	return Server{port: port, mux: mux}
}

// Handler returns the underlying http.Handler for use with http.Server.
func (s Server) Handler() http.Handler {
	return s.mux
}
