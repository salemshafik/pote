# Shared Types

This package contains shared type definitions, API contracts, and protocol specifications used across Pote's services and frontend.

## Contents

- `proto/` — Protocol Buffer definitions for inter-service communication
- `openapi/` — OpenAPI 3.0 specifications for REST APIs
- `ts/` — TypeScript interfaces shared with the frontend

## Usage

### Protocol Buffers
Proto files define the canonical data models. Generate code for Go and TypeScript using:
```bash
# Generate Go code
protoc --go_out=. --go_opt=paths=source_relative proto/*.proto

# Generate TypeScript types
npx protoc-gen-ts proto/*.proto
```

### OpenAPI Specs
Used for API documentation and client code generation. Can be served via Swagger UI.

### TypeScript Interfaces
Shared types imported by the Next.js frontend to ensure type safety across the stack.
