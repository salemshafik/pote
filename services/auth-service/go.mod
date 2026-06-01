module github.com/salemshafik/pote/services/auth-service

go 1.23.0

require (
	github.com/go-chi/chi/v5 v5.1.0
	github.com/go-chi/cors v1.2.1
	github.com/jackc/pgx/v5 v5.7.1
	github.com/salemshafik/pote/packages/auth-utils v0.0.0
	github.com/salemshafik/pote/packages/config v0.0.0
	github.com/salemshafik/pote/packages/logger v0.0.0
	golang.org/x/crypto v0.28.0
	golang.org/x/oauth2 v0.23.0
)

require (
	cloud.google.com/go/compute/metadata v0.3.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/text v0.19.0 // indirect
)

replace (
	github.com/salemshafik/pote/packages/auth-utils => ../../packages/auth-utils
	github.com/salemshafik/pote/packages/config => ../../packages/config
	github.com/salemshafik/pote/packages/logger => ../../packages/logger
)
