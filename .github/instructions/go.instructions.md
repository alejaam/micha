---
applyTo: "**/*.go"
---

# Go Coding Instructions

## Error Handling

- Always wrap errors with context: `fmt.Errorf("operation: %w", err)`
- Use guard clauses — return early on error, avoid nested ifs
- Map domain errors to HTTP status codes only at the handler layer

## Logging

- Use `slog` with structured fields: `slog.InfoContext(ctx, "msg", "key", val)`
- Never log sensitive data (passwords, tokens, PII)
- Always propagate `context.Context` as the first argument

## HTTP Handlers

- Apply `http.MaxBytesReader(w, r.Body, maxBytes)` before decoding
- Use `json.NewDecoder(r.Body).DisallowUnknownFields()`
- Validate and decode into a dedicated request struct; never decode directly into domain types

## Database

- Always use parameterised queries — never interpolate user input into SQL
- Scan results into `*Attributes` structs and rehydrate via domain constructors
- Wrap DB errors before returning them to the caller

## Testing

- Table-driven tests with `testify/assert` and `testify/require`
- Test file lives alongside the package it tests (`foo_test.go`)
- Use subtests: `t.Run("description", func(t *testing.T) { ... })`
- Mock outbound ports with hand-written fakes or `testify/mock`

## Domain Layer

- Never import `net/http`, database drivers, or any infrastructure package from `internal/domain/`
- All validation lives in the domain constructor; return typed errors from `domain/shared/errors.go`
- Use value objects for concepts with invariants (e.g., `Amount`, `Email`)
