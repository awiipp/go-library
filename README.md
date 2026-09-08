# go-library

A small library management system built with Go, split into two independent microservices — `user-service` and `book-service`. The services trust each other through an asymmetrically signed JWT (RS256) rather than calling one another's APIs to verify identity.

## Overview

`user-service` handles registration, login, refresh tokens, and profiles. It signs access tokens with a private key that never leaves the service. `book-service` handles the book catalog and borrowing/returning, and only holds the matching public key — enough to verify a token's signature locally, with no network call back to `user-service` on every request.

Each service owns its own PostgreSQL database and Redis cache. There is no cross-database join or shared table; the two services only agree on the shape of the JWT claims.

| Service | Responsibility |
|---|---|
| `user-service` | Registration, login, refresh tokens, profile, roles |
| `book-service` | Book catalog, borrowing, returning |

## Tech Stack

- **Go** with **Fiber** as the HTTP framework
- **PostgreSQL 17**, one database per service
- **Redis 8** for caching (user profiles, book details)
- **JWT (RS256)** for stateless, cross-service authentication
- **GORM** in user-service and **database/sql with raw SQL** in book-service — deliberately different, to practice both approaches
- **golang-migrate** for schema migrations
- **testify** for unit testing and mocking
- **Docker Compose** for Postgres, Redis, and migrations

## Features

**Auth & Authorization**
- Registration and login with bcrypt password hashing
- Short-lived JWT access tokens (RS256)
- Refresh tokens with rotation and reuse detection — reusing an already-rotated refresh token revokes every session for that user
- Role-based access control (`reader`, `admin`)

**Catalog & Borrowing**
- Book CRUD, restricted to `admin`
- Borrowing and returning, open to any authenticated user; stock is decremented/incremented inside a single database transaction to avoid race conditions when a book is borrowed concurrently
- Per-user borrowing history

**Reliability**
- Cache-aside pattern for user profiles and book details, with graceful degradation — a Redis outage never fails the underlying request
- Cache invalidation after writes (borrow/return)

## Project Layout

Each service follows the same clean architecture: `domain` holds only entities and repository interfaces, with no dependency on DTOs, databases, or external libraries. Dependencies flow one way — handler → usecase → domain ← repository. Both services share the same internal layout (`cmd`, `internal/domain`, `internal/usecase`, `internal/repository`, `internal/handler`, `internal/middleware`), with `user-service` additionally exposing a `cmd/seeder` command to create the initial admin account.

## Running Locally

Each service runs directly on the host (`go run`), while Postgres and Redis run in Docker.

### 1. Set up environment variables

Create a `.env` file in each service folder with your database, Redis, and RSA key path settings.

### 2. Generate the RSA keypair (user-service only)

```bash
cd user-service
mkdir -p certs
openssl genrsa -out certs/private.pem 2048
openssl rsa -in certs/private.pem -pubout -out certs/public.pem
cp certs/public.pem ../book-service/certs/public.pem
```

`book-service` only ever needs `public.pem` — it never holds the private key.

### 3. Start infrastructure and run migrations

```bash
# user-service
cd user-service
docker compose up -d postgres
docker compose run --rm migrate

# book-service
cd ../book-service
docker compose up -d postgres redis
docker compose run --rm migrate
```

### 4. Seed the admin account (once)

```bash
cd user-service
go run cmd/seeder/main.go
```

Credentials come from `SEED_ADMIN_EMAIL` / `SEED_ADMIN_PASSWORD` in `.env`. The command is idempotent and safe to re-run.

### 5. Run both services

```bash
# terminal 1
cd user-service && go run cmd/api/main.go

# terminal 2
cd book-service && go run cmd/api/main.go
```

## Cross-Service Authentication Flow

A client logs in against `user-service` and receives an access token (signed with the private key) and a refresh token (stored server-side as a SHA-256 hash). The access token is then sent to `book-service` as a bearer token; `book-service`'s `RequireAuth` middleware verifies its signature using the public key alone — a pure cryptographic check, with no database lookup or call back to `user-service`. When the access token expires, the client exchanges the refresh token for a new pair; the old refresh token is revoked as part of that exchange (rotation), and if a revoked refresh token is ever reused, every session belonging to that user is revoked as a signal that the token was compromised.

## Main Endpoints

**user-service** (`/v1/api`)
| Method | Path | Auth |
|---|---|---|
| POST | `/auth/register` | Public |
| POST | `/auth/login` | Public |
| POST | `/auth/refresh` | Public |
| GET | `/users/profile` | Bearer token |

**book-service** (`/v1/api`)
| Method | Path | Auth |
|---|---|---|
| GET | `/books` | Bearer token |
| GET | `/books/:id` | Bearer token |
| POST | `/books` | Bearer token + `admin` |
| PUT | `/books/:id` | Bearer token + `admin` |
| DELETE | `/books/:id` | Bearer token + `admin` |
| POST | `/books/:id/borrow` | Bearer token |
| POST | `/loans/:id/return` | Bearer token |
| GET | `/loans/me` | Bearer token |

## Testing

```bash
cd user-service && go test ./internal/usecase/... -v
cd book-service && go test ./internal/usecase/... -v
```

Unit tests use mocked repositories (`testify/mock`) and don't require Postgres or Redis to run. Coverage includes:
- Duplicate email/username handling on registration, including the race condition closed by a database unique constraint
- Identical error responses for "user not found" vs. "wrong password", to prevent user enumeration
- Transactional atomicity when borrowing a book — stock is never decremented if the loan record fails to write, or vice versa
- Ownership checks — a user cannot return a loan that belongs to someone else

## Notable Design Decisions

- **RS256 over HS256** — book-service never holds a secret capable of *issuing* tokens, only verifying them, which limits the blast radius if either service is compromised.
- **Loans live in book-service**, not a separate service — this avoids needing a saga/distributed-transaction pattern for something that a single local database transaction (decrement stock + record the loan) already handles correctly.
- **`user_id` in the `loans` table has no foreign key** — the referenced user lives in a different database entirely; its validity is guaranteed by a verified JWT claim rather than a database constraint.
