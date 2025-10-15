# Errand Shop Backend

This is the backend service for Errand Shop, a Go (Golang) application that powers order management, deliveries, payments, products, and user management.

## Tech Stack
- Go 1.22+
- Fiber (HTTP server)
- GORM + PostgreSQL (DB)
- Paystack integration (payments)
- JWT-based auth

## Getting Started
- Ensure Go is installed: `go version`
- Set environment variables (copy `.env.example` to `.env` and fill values)
- Run database migrations if necessary
- Start the server: `go run cmd/server/main.go`

## Key Domains
- `orders`: cart, checkout, payment status, admin actions
- `payments`: initialize/process payments, webhooks, refunds
- `delivery`: zone matching, fee calculation, routes
- `products`: catalog, stock management
- `customers/users`: profiles, authentication

## Project Structure
- `cmd/server/main.go`: application bootstrap
- `internal/domain/*`: core business logic per domain
- `internal/http/handlers`: HTTP handlers
- `internal/repos`: repositories
- `data/delivery_zones.json`: delivery zone pricing data

## Common Commands
- Run server: `go run cmd/server/main.go`
- Lint/format: `go fmt ./...`
- Test (if present): `go test ./...`

## Notes
- Delivery fee uses zone-based pricing via the matcher.
- Payment success updates order payment status to `paid`.
- Ensure database and env vars are configured before starting.