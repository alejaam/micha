package httpadapter

import (
	"net/http"
)

// ServerDependencies groups all resource-level dependencies for the HTTP server.
type ServerDependencies struct {
	Auth         AuthHandlerDeps
	Expense      ExpenseHandlerDeps
	Household    HouseholdHandlerDeps
	Member       MemberHandlerDeps
	Settlement   SettlementHandlerDeps
	JWTValidator TokenValidator
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

	// Public auth routes — no middleware.
	ah := newAuthHandler(deps.Auth)
	mux.HandleFunc("POST /v1/auth/register", ah.handleRegister)
	mux.HandleFunc("POST /v1/auth/login", ah.handleLogin)

	// Protected routes — all wrapped with JWT auth middleware.
	protect := AuthMiddleware(deps.JWTValidator)

	// Protected auth routes.
	mux.Handle("GET /v1/auth/me", protect(http.HandlerFunc(ah.handleMe)))

	eh := newExpenseHandler(deps.Expense)
	mux.Handle("POST /v1/expenses", protect(http.HandlerFunc(eh.handleCreate)))
	mux.Handle("GET /v1/expenses/{id}", protect(http.HandlerFunc(eh.handleGet)))
	mux.Handle("GET /v1/expenses", protect(http.HandlerFunc(eh.handleList)))
	mux.Handle("PATCH /v1/expenses/{id}", protect(http.HandlerFunc(eh.handlePatch)))
	mux.Handle("DELETE /v1/expenses/{id}", protect(http.HandlerFunc(eh.handleDelete)))

	hh := newHouseholdHandler(deps.Household)
	mux.Handle("POST /v1/households", protect(http.HandlerFunc(hh.handleCreate)))
	mux.Handle("GET /v1/households", protect(http.HandlerFunc(hh.handleList)))

	mh := newMemberHandler(deps.Member)
	mux.Handle("POST /v1/households/{household_id}/members", protect(http.HandlerFunc(mh.handleCreate)))
	mux.Handle("GET /v1/households/{household_id}/members", protect(http.HandlerFunc(mh.handleList)))

	sh := newSettlementHandler(deps.Settlement)
	mux.Handle("GET /v1/households/{household_id}/settlement", protect(http.HandlerFunc(sh.handleGetMonthly)))

	return Server{port: port, mux: mux}
}

// Handler returns the underlying http.Handler for use with http.Server.
func (s Server) Handler() http.Handler {
	return s.mux
}
