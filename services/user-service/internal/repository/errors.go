// Package repository provides data access for the user-service.
package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// Common repository errors shared across the user-service repositories.
var (
	ErrProfileNotFound     = errors.New("user profile not found")
	ErrProfileExists       = errors.New("user profile already exists")
	ErrContactNotFound     = errors.New("contact not found")
	ErrContactExists       = errors.New("contact already exists")
	ErrSelfContact         = errors.New("cannot add yourself as a contact")
	ErrInviteNotFound      = errors.New("invite not found")
	ErrInviteAlreadyExists = errors.New("a pending invite already exists for this email")
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
