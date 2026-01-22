# Chirpy — Minimal Twitter-like Go API

Overview
--------
Chirpy is a small, production-oriented example API written in Go. It provides user registration and authentication, short "chirp" messages, and a refresh-token based authentication flow. The project demonstrates a clean separation between handlers and database access using sqlc-generated code, secure password hashing with Argon2id, JWT-based access tokens, and database migrations driven by the goose CLI.

Key technologies
----------------
- Language: Go (see `go.mod`)
- SQL code generation: `sqlc` (configuration: `sqlc.yaml`) — generates types and query methods in `internal/database`
- Migrations: `goose` (invoked in Makefile targets under `sql/schema`)
- Authentication: `github.com/golang-jwt/jwt/v5` for JWTs and `github.com/alexedwards/argon2id` for password hashing
- Database: PostgreSQL via `github.com/lib/pq`

Repository layout
-----------------
- `main.go` — application entrypoint and HTTP routing
- `auth_handler.go`, `users_handler.go`, `chirps_handler.go` — HTTP handlers for auth, user, and chirp endpoints
- `internal/auth` — authentication helpers (password hashing, JWT creation/validation, refresh token generation)
- `internal/database` — sqlc-generated database access layer (models and queries)
- `sql/schema` — SQL migration files (numbered SQL files)
- `sql/queries` — SQL query files used by sqlc
- `Makefile` — convenience targets for running migrations (`goose`)
- `sqlc.yaml` — sqlc configuration

Getting started (development)
-----------------------------
Prerequisites
- Go (version from `go.mod`)
- PostgreSQL
- `sqlc` CLI (`https://sqlc.dev`) installed to generate Go code from SQL
- `goose` CLI installed for running migrations (Makefile uses `goose`)

Environment
-----------
The server expects these environment variables (commonly provided via a `.env` file in development):

- `DB_URL` — PostgreSQL connection string (e.g. `postgres://user:pass@localhost:5432/chirpy?sslmode=disable`)
- `PLATFORM` — deployment platform identifier (used by platform-check middleware; set to `dev` for local admin access)
- `SECRETKEY` — HMAC secret used for signing JWT access tokens
- `POLKAKEY` — API key used by the Polka webhook handler

Database migrations (goose)
---------------------------
Migrations live in `sql/schema`. The repository includes Makefile shortcuts to run goose against a local Postgres instance.

Examples (local dev):

```sh
# run all pending migrations
make goose-up

# revert last migration
make goose-down

# down then up
make goose-re
```

Note: the Makefile calls `goose` and a default connection string; adjust the command or run `goose` directly if you use a different connection string.

Generating database code (sqlc)
-------------------------------
sqlc converts SQL queries in `sql/queries` and the schema in `sql/schema` into strongly typed Go code placed under `internal/database` as configured in `sqlc.yaml`.

To generate or regenerate the database code:

```sh
sqlc generate
```

This will produce the `Queries` methods used throughout the handlers (for example `CreateUser`, `GetUserByEmail`, `CreateRefreshToken`, `ListChirps`).

Authentication overview
-----------------------
The project implements a standard short-lived JWT access token with a long-lived refresh token stored in the database.

- Password hashing: `internal/auth.HashPassword` uses Argon2id (via `github.com/alexedwards/argon2id`), and `CheckPasswordHash` validates passwords.
- Access tokens: `internal/auth.MakeJWT` issues HS256-signed JWTs using the `SECRETKEY`. The token includes standard registered claims (issuer, subject, issued-at, expiry).
- Validation: `internal/auth.ValidateJWT` parses and validates incoming tokens and returns the `uuid` subject.
- Refresh tokens: `internal/auth.MakeRefreshToken` creates a secure random string. Refresh tokens are persisted in the `refresh_tokens` table via sqlc generated `CreateRefreshToken`, and can be revoked via `RevokeRefreshToken`.

High-level auth flows
---------------------
- Login (`POST /api/login`): validate credentials, return an access JWT and a refresh token (refresh saved in DB).
- Refresh (`POST /api/refresh`): client sends the refresh token as a Bearer token; server verifies the token exists and is not expired or revoked, and returns a new access JWT.
- Revoke (`POST /api/revoke`): revoke a refresh token (set `revoked_at` in DB).

HTTP endpoints (summary)
------------------------
Below are the main public endpoints provided by the server:

- `POST /api/users` — create a new user (body: `{ "email": ..., "password": ... }`)
- `POST /api/login` — exchange credentials for `{ token, refresh_token }`
- `POST /api/refresh` — exchange refresh token for a new access token (send refresh token as Bearer token)
- `POST /api/revoke` — revoke a refresh token
- `POST /api/chirps` — create a chirp (requires `Authorization: Bearer <access-token>`)
- `GET /api/chirps` — list chirps (optional `author_id` and `sort` query params)
- `GET /api/chirps/{chirpID}` — get a chirp by id
- `DELETE /api/chirps/{chirpID}` — delete a chirp (requires authorization; only the owner may delete)

Examples
--------
Create a user:

```sh
curl -X POST -H "Content-Type: application/json" \
	-d '{"email":"alice@example.com","password":"s3cret"}' \
	http://localhost:8080/api/users
```

Login and use a protected endpoint:

```sh
# login
curl -X POST -H "Content-Type: application/json" \
	-d '{"email":"alice@example.com","password":"s3cret"}' \
	http://localhost:8080/api/login

# assume response contains a JWT in `token`
curl -H "Authorization: Bearer <token>" \
	-X POST -H "Content-Type: application/json" \
	-d '{"body":"hello world"}' \
	http://localhost:8080/api/chirps
```

Development notes
-----------------
- The `internal/database` package is generated; do not edit sqlc-generated files directly. Edit SQL under `sql/queries` or the schema under `sql/schema` and re-run `sqlc generate`.
- Migrations are the source of truth for schema changes; add new numbered SQL migration files to `sql/schema` and apply them with `goose`.
- Secrets (like `SECRETKEY`) should be managed securely in production (e.g. environment config, secrets manager), not committed to source.

Contributing
------------
Submit pull requests for bug fixes and improvements. If you change the schema, add a new migration under `sql/schema` and update `sql/queries` as needed, then run `sqlc generate` and commit the generated code.

License
-------
This project includes a `LICENSE` file. See it for license terms.

Files to inspect
-----------------
- `sqlc.yaml` — sqlc configuration ([sqlc.yaml](sqlc.yaml))
- `Makefile` — includes `goose` targets for migrations ([Makefile](Makefile))
- `internal/auth` — JWT and password helpers ([internal/auth/auth.go](internal/auth/auth.go))
- `internal/database` — sqlc output (do not edit) ([internal/database](internal/database))

If you'd like, I can also:
- Run `sqlc generate` and commit generated files (requires `sqlc` installed locally).
- Run the migrations against a local DB and attempt to start the server.

---


