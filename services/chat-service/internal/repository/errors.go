// Package repository provides data access for the chat-service.
package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// Common repository errors shared across the chat-service repositories.
var (
	ErrChatNotFound   = errors.New("chat not found")
	ErrMemberNotFound = errors.New("chat member not found")
	ErrMemberExists   = errors.New("user is already a member of this chat")
)

// PostgreSQL error codes. See https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
	pgCheckViolation      = "23514"
)

// pgErrorCode extracts the SQLSTATE code from a pgx error, returning the empty
// string if err is not a *pgconn.PgError. Using errors.As against the typed
// PgError is the robust approach (vs. substring-matching the error message).
func pgErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

// isUniqueViolation reports whether err is a PostgreSQL unique_violation.
func isUniqueViolation(err error) bool {
	return pgErrorCode(err) == pgUniqueViolation
}

// isForeignKeyViolation reports whether err is a PostgreSQL foreign_key_violation.
func isForeignKeyViolation(err error) bool {
	return pgErrorCode(err) == pgForeignKeyViolation
}

// isCheckViolation reports whether err is a PostgreSQL check_violation.
func isCheckViolation(err error) bool {
	return pgErrorCode(err) == pgCheckViolation
}
