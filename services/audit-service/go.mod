module github.com/salemshafik/pote/services/audit-service

go 1.23.0

require (
	github.com/go-chi/chi/v5 v5.1.0
	github.com/salemshafik/pote/packages/auth-utils v0.0.0
	github.com/salemshafik/pote/packages/config v0.0.0
	github.com/salemshafik/pote/packages/logger v0.0.0
)

replace (
	github.com/salemshafik/pote/packages/auth-utils => ../../packages/auth-utils
	github.com/salemshafik/pote/packages/config => ../../packages/config
	github.com/salemshafik/pote/packages/logger => ../../packages/logger
)
