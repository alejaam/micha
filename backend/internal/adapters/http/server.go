package httpadapter

import (
	"net/http"

	"micha/backend/internal/ports/outbound"
)

// ServerDependencies groups all resource-level dependencies for the HTTP server.
type ServerDependencies struct {
	Auth             AuthHandlerDeps
	Expense          ExpenseHandlerDeps
	RecurringExpense RecurringExpenseHandlerDeps
	Household        HouseholdHandlerDeps
	Member           MemberHandlerDeps
	Settlement       SettlementHandlerDeps
	Category         CategoryHandlerDeps
	SplitConfig      SplitConfigHandlerDeps
	JWTValidator     outbound.TokenValidator
	MemberRepo       outbound.MemberRepository
	AllowedOrigins   []string
}

// Server is the primary HTTP adapter.
type Server struct {
	port    string
	handler http.Handler
}

// NewServer constructs a Server and registers all routes.
func NewServer(port string, deps ServerDependencies) Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	// Public auth routes — no middleware.
	ah := newAuthHandler(deps.Auth)
	mux.HandleFunc("POST /v1/auth/register", ah.handleRegister)
	mux.HandleFunc("POST /v1/auth/login", ah.handleLogin)

	// Protected routes — JWT auth middleware.
	protect := AuthMiddleware(deps.JWTValidator)

	// Protected routes that also require household membership — JWT + HouseholdAuthz.
	householdAuthz := HouseholdAuthzMiddleware(deps.MemberRepo)
	protectHousehold := func(h http.Handler) http.Handler {
		return protect(householdAuthz(h))
	}

	// Member creation supports bootstrap flow: allow first member in empty households.
	householdMemberOrEmptyAuthz := HouseholdMemberOrEmptyAuthzMiddleware(deps.MemberRepo)
	protectMemberCreate := func(h http.Handler) http.Handler {
		return protect(householdMemberOrEmptyAuthz(h))
	}

	// Protected auth routes.
	mux.Handle("GET /v1/auth/me", protect(http.HandlerFunc(ah.handleMe)))

	eh := newExpenseHandler(deps.Expense)
	mux.Handle("POST /v1/expenses", protect(http.HandlerFunc(eh.handleCreate)))
	mux.Handle("GET /v1/expenses/{id}", protect(http.HandlerFunc(eh.handleGet)))
	mux.Handle("GET /v1/expenses", protect(http.HandlerFunc(eh.handleList)))
	mux.Handle("PATCH /v1/expenses/{id}", protect(http.HandlerFunc(eh.handlePatch)))
	mux.Handle("DELETE /v1/expenses/{id}", protect(http.HandlerFunc(eh.handleDelete)))

	reh := newRecurringExpenseHandler(deps.RecurringExpense)
	mux.Handle("POST /v1/recurring-expenses", protect(http.HandlerFunc(reh.handleCreate)))
	mux.Handle("GET /v1/recurring-expenses/{id}", protect(http.HandlerFunc(reh.handleGet)))
	mux.Handle("GET /v1/recurring-expenses", protect(http.HandlerFunc(reh.handleList)))
	mux.Handle("PATCH /v1/recurring-expenses/{id}", protect(http.HandlerFunc(reh.handleUpdate)))
	mux.Handle("DELETE /v1/recurring-expenses/{id}", protect(http.HandlerFunc(reh.handleDelete)))
	mux.Handle("POST /v1/recurring-expenses/generate", protect(http.HandlerFunc(reh.handleGenerate)))

	hh := newHouseholdHandler(deps.Household)
	mux.Handle("POST /v1/households", protect(http.HandlerFunc(hh.handleCreate)))
	mux.Handle("GET /v1/households", protect(http.HandlerFunc(hh.handleList)))
	mux.Handle("GET /v1/households/{household_id}", protectHousehold(http.HandlerFunc(hh.handleGet)))
	mux.Handle("PUT /v1/households/{household_id}", protectHousehold(http.HandlerFunc(hh.handleUpdate)))

	mh := newMemberHandler(deps.Member)
	mux.Handle("POST /v1/households/{household_id}/members", protectMemberCreate(http.HandlerFunc(mh.handleCreate)))
	mux.Handle("GET /v1/households/{household_id}/members", protectHousehold(http.HandlerFunc(mh.handleList)))
	mux.Handle("PUT /v1/households/{household_id}/members/{member_id}", protectHousehold(http.HandlerFunc(mh.handleUpdate)))
	mux.Handle("DELETE /v1/households/{household_id}/members/{member_id}", protectHousehold(http.HandlerFunc(mh.handleDelete)))

	sh := newSettlementHandler(deps.Settlement)
	mux.Handle("GET /v1/households/{household_id}/settlement", protectHousehold(http.HandlerFunc(sh.handleGetMonthly)))

	ch := newCategoryHandler(deps.Category)
	mux.Handle("POST /v1/households/{household_id}/categories", protectHousehold(http.HandlerFunc(ch.handleCreate)))
	mux.Handle("GET /v1/households/{household_id}/categories", protectHousehold(http.HandlerFunc(ch.handleList)))
	mux.Handle("DELETE /v1/households/{household_id}/categories/{category_id}", protectHousehold(http.HandlerFunc(ch.handleDelete)))

	sch := newSplitConfigHandler(deps.SplitConfig)
	mux.Handle("PUT /v1/households/{household_id}/split-config", protectHousehold(http.HandlerFunc(sch.handleUpdate)))

	// Apply middleware chain: RequestID -> CORS -> routes
	cors := CORSMiddleware(CORSConfig{AllowedOrigins: deps.AllowedOrigins})
	handler := RequestIDMiddleware(cors(mux))

	return Server{port: port, handler: handler}
}

// Handler returns the underlying http.Handler for use with http.Server.
func (s Server) Handler() http.Handler {
	return s.handler
}
