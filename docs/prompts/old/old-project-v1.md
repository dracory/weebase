OLD OVERVIEW:
Check docs/overview.md

IGNORE EVERYTHING BELOW THIS LINE

# Project Prompt: Modern Web Adminer Clone in Go (GORM-powered)

## Vision
Build a modern, single-application, single-endpoint, web-based database management
admin tool—functionally equivalent to Adminer/AdminerEvo—implemented in Go.
 using GORM as the ORM/DB abstraction. 
 
The deliverable is a reusable Go module/package (library) that exposes
an `http.Handler` and can be embedded into any Go project, plus an optional
thin standalone binary for self-hosting.

It supports multiple database engines and provides a clean, responsive UI
for listing the databases on the server, browsing the tables, editing the data,
running queries, and administering the databases.

## Non-Goals
- Not a replacement for full DB admin consoles like pgAdmin or SQL Server Management Studio.
- No vendor-specific advanced administration outside CRUD/schema basics and safe utilities.
- No desktop/native app; strictly a web app.

## Core Principles
- Security-first by default.
- Parity with Adminer’s essential features.
- Database-agnostic via GORM with pluggable drivers.
- Clean, fast, responsive UX with sensible limits and guardrails.
- Observable, testable, and easy to deploy (Docker-ready).

## Tech Stack
- Language: Go >= 1.22
- HTTP server: Go standard library `net/http` (no external web framework)
- ORM/DB: GORM (v2)
- DB drivers (GORM-compatible):
  - MySQL/MariaDB: gorm.io/driver/mysql
  - PostgreSQL: gorm.io/driver/postgres
  - SQLite: gorm.io/driver/sqlite
  - SQL Server: gorm.io/driver/sqlserver
- Frontend: Server-rendered templates using HB (HTML Builder) utilities with embedded CSS/JS (no external packages/CDNs; no Tailwind build step)
- Auth: Cookie-session with secure headers + optional OIDC
- Config: dracor/env
- Logging: Go slog (log/slog) for structured JSON logs
- Migrations: None for user DBs; internal app migrations OK
- Packaging: Docker + docker-compose
- Testing: go test , using only standard Go library
- API responses: standardize JSON envelopes via github.com/dracory/api

## High-Level Architecture
- HTTP layer: routes, handlers, request validation, auth middleware
- Service layer: connection mgmt, metadata discovery, query execution, transactions
- Data layer: GORM for CRUD, raw SQL passthrough for DDL/metadata when needed
- View layer: server-rendered templates with HB components/partials; accessible and responsive

### Connection Model
- “Connection Profile”: target DB (driver, host/port/DSN, creds, options).
- Profiles can be ephemeral (session) or persisted (encrypted at rest).
- Per-session selected connection; preference store per user.

## Embeddable Module/SDK (Library Mode)
- Goal: ship as an importable Go module that can be embedded into any Go web app, in addition to running as a standalone server.
- Package layout:
  - Root package: core types, services, interfaces, HTML templates, and an `http.Handler` implementation.
  - `cmd/server`: thin binary that mounts the handler on a single endpoint path.
- Public interfaces (examples):
  - `DriverRegistry` for enabling DB drivers.
  - `ConnectionStore` for persisted profiles (pluggable: file, DB, custom).
  - `AuthProvider`/`Authorizer` for integrating app-specific auth/RBAC.
  - `AuditSink` for audit event handling (log, webhook, custom store).
  - `Renderer` for theming/skinning and template overrides.
- HTTP integration pattern (framework-agnostic):
  - Expose a single `http.Handler` (e.g., `adminer.Handler`) that routes all operations via a query parameter (default: `action`).
  - Provide a helper `Register(mux *http.ServeMux, path string, h http.Handler)` for convenience; host app can mount with any router.
- Configuration (library mode):
  - `Options` struct supports: enabled drivers, safe mode default, limits, base URL, storage paths, OIDC/local auth toggles, asset/theme overrides.
  - Accept functional options for extendability (e.g., `WithAuditSink(...)`).
  - Connection provisioning options:
    - `DefaultConnection` (optional): DSN or structured fields (driver, host, port, db, user, ssl, extras). When present, UI starts connected and manages this DB immediately.
    - `PreconfiguredProfiles` (optional): list of named `ConnectionProfile`s the user can pick from without entering credentials (creds may still be prompted if omitted).
    - `AllowAdHocConnections` (bool, default true): when true, users can enter connection details at runtime (Adminer-style) if no default/preconfigured connection is chosen.
    - `ReadOnlyMode` (bool): force read-only operations in the UI regardless of DB grants.
- Assets and templates:
  - Embed default templates/assets via `embed` and allow override hooks to load custom templates.
  - Expose a minimal theming API (colors, logo, title) and a full override option.
  - No external network dependencies: all CSS/JS/fonts/images shipped with the binary via `embed`.
- Extensibility hooks:
  - Middleware registration points (pre/post handler hooks, query execution hooks).
  - Result renderers: register custom table/JSON renderers.
- Versioning & distribution:
  - Semantic versioning (SemVer) for the module.
  - Go Module path: `github.com/dracory/weebase`.
  - Keep stable public interfaces; document any breaking changes clearly in release notes.
- Minimal embedding example (net/http, single endpoint):
  ```go
  package main

  import (
    "log"
    "net/http"
    weebase "github.com/dracory/weebase"
  )

  func main() {
    h := weebase.NewHandler(weebase.Options{
      EnabledDrivers:    []string{"postgres", "mysql", "sqlite"},
      SafeModeDefault:   true,
      AllowAdHocConnections: true,
    })

    // Mount on single endpoint; all actions via query, e.g., /db?action=browse_rows
    http.Handle("/db", h)

    log.Println("listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
  }
  ```

### Integration Scenarios
- Drop-in admin UI inside an existing SaaS admin panel.
- Per-tenant connections in multi-tenant apps (apply RBAC via `Authorizer`).
- Self-hosted internal tool: use module in binary mode or embedded mode.
- Extend with custom audit sinks and storage backends.

### Connection Provisioning Modes
- Preconfigured connection(s): If `DefaultConnection` or `PreconfiguredProfiles` are set in `Options`/config, the UI lands connected to the specified DB (or a profile picker), allowing immediate management with no manual input.
- Ad-hoc connection: If no connection is pre-set (or `AllowAdHocConnections` is true), the UI presents a connection form (driver, host, port, user, password, DB, options). On successful connect, the session becomes active and the user can start managing the DB.

## Single Endpoint Contract (`?action=`)
- All interactions are served from a single HTTP path (e.g., `/db`), with behavior selected by the `action` query parameter. Responses are HTML (server-rendered) with optional JSON for XHR. JSON responses MUST use the standard envelope from github.com/dracory/api.
- Default query key is `action` (configurable via `Options.ActionParam`, default `"action"`).
- Representative actions and params (non-exhaustive; to be documented precisely during implementation):
  - `action=connect` (POST): driver, host, port, db, user, password, options
  - `action=list_schemas` (GET)
  - `action=list_tables` (GET, schema)
  - `action=table_info` (GET, table)
  - `action=browse_rows` (GET, table, page, limit, sort, filter)
  - `action=insert` (POST, table, row fields...)
  - `action=update` (POST, table, pk/where, fields...)
  - `action=delete` (POST, table, pk/where)
  - `action=sql_execute` (POST, sql, transactional=[true|false])
  - `action=sql_explain` (POST, sql)
  - `action=export` (POST, table|sql, format=[csv|json|sql])
  - `action=import` (POST, table, file, mapping...)
  - `action=ddl_create_table|ddl_alter_table|ddl_drop_table` (POST, payload...)
  - `action=view_definition` (GET, view)
  - `action=login|logout` (POST)
  - `action=profiles` (GET/POST for listing/adding preconfigured profiles if allowed)
- CSRF protection applies to all state-changing actions (POSTs). Safe-mode confirms/blocks destructive DDL unless explicitly allowed.

## Feature Parity (Adminer/AdminerEvo)
- Connections:
  - Add/connect via host/port or DSN; select database/schema.
  - Manage multiple profiles; quick switching.
- Schema navigation:
  - List databases/schemas.
  - List tables/views; pagination and search.
  - Table details: columns, indexes, constraints, row count estimate.
  - View definition display.
- Data operations:
  - Browse rows: pagination, sorting, filtering (builder + raw WHERE).
  - Edit row; bulk update when feasible.
  - Insert rows with validation.
  - Delete single/multiple rows with confirmation and transaction safety.
- SQL console:
  - Multi-statement execution with transaction toggle.
  - Result tabs; copy/download CSV/JSON.
  - EXPLAIN/EXPLAIN ANALYZE where supported.
  - Save/load named queries (per user, per connection).
- DDL helpers:
  - Create/alter/drop table (editor with preview SQL).
  - Manage indexes and constraints.
  - Create/drop views; show definition; validate SQL.
  - Routines/functions: list, view definition; create/edit if supported.
- Import/Export:
  - Export table/query as CSV, JSON, SQL (INSERTs), optional gzip.
  - Import CSV/SQL with preview, column mapping, transactional apply.
- Utilities:
  - Row count estimate, table size (where supported).
  - Non-invasive “vacuum/analyze” hints where relevant.
  - Simple dependency graph (FK refs) when discoverable.
- Plugins/Extensions:
  - Hook points: middleware, menu injection, custom renderers, auth providers.
  - Registry-based extension loading (build tags or simple init registration).

## UX Requirements
- Responsive UI with embedded CSS (no external packages); dark mode.
- Global quick search: tables, views, columns.
- Breadcrumbs; context-aware actions.
- Keyboard shortcuts: run query, save, paginate.
- Safe-guards: confirmations and preview SQL.
- Clear error surfaces with copy-to-clipboard diagnostics.

## Security Requirements
- HTTPS (behind reverse proxy OK), HSTS, secure cookies, strict headers.
- CSRF protection for state-changing routes.
- Roles:
  - Admin: manage profiles/settings
  - User: use assigned connections
  - Guest (optional): read-only
- Secrets:
  - Never log creds/secrets.
  - Encrypt persisted connection creds at rest.
- Query safety:
  - “Safe mode” default: block/confirm DDL/DROP/ALTER/TRUNCATE unless enabled per session.
  - Default SELECT limit (e.g., 200) unless overridden.
  - Destructive ops wrapped in transactions by default.
- Rate limiting on auth/import endpoints.
- Audit log: who, action, object, rows affected (avoid PII).

## Observability
- Structured JSON logs with request IDs.
- Metrics: latency, error rates, queries, rows, driver usage.
- Health endpoints: /healthz, /readyz (DB ping when a connection is active).

## Configuration
- Env/config:
  - HTTP_PORT, BASE_URL, SESSION_SECRET, OIDC_*, TLS_*, STORAGE_PATH
  - ALLOW_PERSISTED_PROFILES (bool)
  - SAFE_MODE_DEFAULT (bool)
  - MAX_SELECT_ROWS, MAX_UPLOAD_SIZE_MB
  - ENABLED_DRIVERS=[mysql,postgres,sqlite,sqlserver,...]
  - DEFAULT_DSN (optional) or structured `DEFAULT_CONN_*` vars (DRIVER, HOST, PORT, DB, USER, PASSWORD, SSLMODE, OPTIONSJSON) for immediate auto-connect.
  - PRECONFIGURED_PROFILES (optional JSON or file path) to load named profiles at startup.
  - ALLOW_ADHOC_CONNECTIONS=true/false to permit Adminer-style runtime connection entry.
  - READ_ONLY_MODE=true/false to force read-only UI regardless of DB grants.
- Per-profile advanced options: SSL mode, timeouts, read-only flag, SSH tunnel (optional).

## Internal Data Model
- User: id, email, role, password hash (if local), OIDC subject
- ConnectionProfile: id, name, driver, DSN/host/port, db/schema, options, encrypted creds, owner/shared
- SavedQuery: id, user_id, connection_id, name, sql, created_at
- AuditEvent: id, user_id, connection_id, action, object_type, object_name, metadata JSON, created_at

## Key Actions (Single Endpoint)
All operations are handled through a single URL endpoint, controlled by the `action` query parameter. This list is representative and not exhaustive. All state-changing actions (POST) must be protected by CSRF tokens.

- **Authentication:**
  - `POST ?action=login`: `user`, `password`
  - `POST ?action=logout`
  - `GET/POST ?action=oidc_callback`: `code`, `state`
- **Connection Management:**
  - `POST ?action=connect`: `driver`, `host`, `port`, `db`, `user`, `password`, `options`
  - `GET ?action=disconnect`
  - `GET ?action=profiles`: List persisted profiles.
  - `POST ?action=profiles_save`: Save a new profile.
- **Schema Browsing:**
  - `GET ?action=list_schemas`
  - `GET ?action=list_tables`: `schema`
  - `GET ?action=table_info`: `table`
  - `GET ?action=view_definition`: `view`
- **Data Operations (CRUD):**
  - `GET ?action=browse_rows`: `table`, `page`, `limit`, `sort`, `filter`
  - `POST ?action=insert_row`: `table`, form fields for row data.
  - `POST ?action=update_row`: `table`, `pk` or `where`, form fields for row data.
  - `POST ?action=delete_row`: `table`, `pk` or `where`.
- **SQL Execution:**
  - `POST ?action=sql_execute`: `sql`, `transactional=[true|false]`
  - `POST ?action=sql_explain`: `sql`
  - `GET ?action=list_saved_queries`
  - `POST ?action=save_query`: `name`, `sql`
- **DDL (Schema Modification):**
  - `POST ?action=ddl_create_table`: `payload` (JSON or form data defining table structure).
  - `POST ?action=ddl_alter_table`: `table`, `payload`.
  - `POST ?action=ddl_drop_table`: `table`.
- **Import/Export:**
  - `POST ?action=export`: `table` or `sql`, `format=[csv|json|sql]`
  - `POST ?action=import`: `table`, `file`, `mapping`

## Database Support and Specifics
- Use GORM for CRUD as primary path.
- Metadata/introspection via dialect-aware raw SQL where needed.
- Dialect differences to handle:
  - Identifier quoting, LIMIT/OFFSET, defaults, types, constraints, generated columns.
- Enabled by default: PostgreSQL, MySQL/MariaDB, SQLite.
- Optional (feature flag): SQL Server.

## Import/Export Details
- Export:
  - CSV (streamed), configurable delimiter/quote, header row.
  - JSON (NDJSON and array).
  - SQL INSERTs; option to include CREATE TABLE.
- Import:
  - CSV preview, column mapping, type coercion, per-row error reporting.
  - SQL applied in a transaction; show affected rows.

## Performance and Limits
- Server-side pagination; streaming for large results.
- Configurable row/payload limits.
- Connection pool tuning via config.
- Cancellable queries (contexts with timeout; cancel UI).

## Testing Strategy
- Unit: metadata discovery, query exec, safety rules.
- Integration: containers for Postgres/MySQL; SQLite in-memory.
- UI smoke: connect, browse, CRUD, SQL console, import/export.
- Security tests: CSRF, auth flows, RBAC, safe mode enforcement.

## Deployment
- Single Go binary + embedded/static assets.
- Docker image (small, non-root) with healthchecks.
- docker-compose with sample Postgres/MySQL for demo.
- Example reverse proxy config (Caddy/Traefik/Nginx) with TLS.

## Documentation
- README: setup, config, quick start
- SECURITY.md: threat model, safe mode
- OPERATIONS.md: metrics, logs, backups (profiles, saved queries)
- DRIVER_NOTES.md: dialect quirks, feature coverage

## Roadmap
- Phase 1: MVP
  - Connect; browse schemas/tables; row viewer; read-only SQL console; CSV export; safe mode on
- Phase 2: CRUD + DDL Basics
  - Insert/update/delete; create/drop table/index/view; saved queries; CSV import
- Phase 3: Advanced
  - EXPLAIN; routines; JSON/SQL export; RBAC; audit log; OIDC
- Phase 4: Plugins & Polish
  - Plugin hooks; dark mode; shortcuts; extra drivers; perf tuning

## Acceptance Criteria
- Parity with Adminer essentials:
  - Multi-DB connections; browse schemas/tables/views; CRUD rows; run SQL; import/export; basic DDL
- Online, self-hosted, secure auth, CSRF protection
- Works with PostgreSQL, MySQL/MariaDB, SQLite out of the box
- Safe mode on by default; confirmations for destructive ops
- Responsive UI with embedded CSS and dark mode
- Structured logs, basic metrics, health endpoints
 - JSON responses use the github.com/dracory/api standard envelope for consistency
- Docker image + compose; example works end-to-end
- Strong unit/integration coverage; UI smoke tests green
- Published as a reusable Go module with semantic versioning and stable public interfaces.
- Exposes a single `http.Handler` for mounting on any path; provides a small `Register(*http.ServeMux, path, handler)` helper.
- Clean embedding story: works under a configurable URL prefix, honors host app middleware, and allows template/theme overrides.
 - Supports both provisioning modes:
   - Preconfigured: honors DEFAULT_DSN or profiles to start connected immediately.
   - Ad-hoc: if enabled, renders connection form and connects session on success.
 - All UI assets (CSS/JS/fonts/images) are embedded via Go `embed`; no runtime CDN or external fetches.

## Nice-to-Haves
- Query plan visualization
- ER diagram (read-only)
- Data diff across tables/environments
- SSH tunnel per connection profile
- Read-only share links for saved queries

## Detailed Implementation Plan

### Phase 0: Project Scaffolding & Foundations
- [x] Initialize module path `github.com/dracory/weebase` and baseline `go.mod`
- [x] Establish project layout: root package, `cmd/server/`, `templates/`, `assets/`, `docs/`
- [x] Add logging (Go slog) wrapper with request ID middleware
- [x] Add configuration loader (env + flags), defaults, and validation
- [x] Add security headers middleware (HSTS, CSP, X-Frame-Options, etc.)
- [x] Add session/cookie setup with configurable `SESSION_SECRET`
- [x] Embed assets/templates with `embed` and plumb override hooks

### Phase 1: HTTP Handler and Single-Endpoint Router
- [x] Define `Options` and functional options in the root package
- [x] Implement `NewHandler(opts Options) http.Handler`
- [x] Implement action dispatcher based on `Options.ActionParam` (default `action`)
- [x] Define response helpers (HTML render, JSON envelope via `github.com/dracory/api`)
- [x] Add CSRF protection for POST actions

### Phase 2: Driver Registry and Connection Management
- [x] Create `DriverRegistry` with enable/disable flags from `Options`
- [x] Support Postgres, MySQL/MariaDB, SQLite initially; optional SQL Server via flag
- [x] Implement `ConnectionStore` (in-memory first; pluggable interface)
- [x] Implement `connect` action: form/DSN parsing, validation, open DB with GORM
- [x] Add connection-in-session handling, ping/health checks
- [x] Add `DefaultConnection` and `PreconfiguredProfiles` boot logic
- [x] Implement `profiles` list/create actions (guarded by RBAC)

### Phase 3: AuthN/Z and RBAC
- [ ] Implement local auth (optional) with password hashing (bcrypt/argon2)
- [ ] Wire OIDC (optional): config, callback, token validation
- [ ] Define roles: Admin/User/Guest and `Authorizer` interface
- [ ] Gate sensitive actions and profiles management by role
- [ ] Rate-limit auth endpoints

### Phase 4: Schema Discovery and Navigation
- [x] Implement dialect-aware metadata discovery (schemas, tables, views)
- [x] `list_schemas`, `list_tables`, `table_info` actions
- [x] `view_definition` action
- [x] Standardize identifier quoting per dialect
- [x] Pagination and search for tables/views

### Phase 5: Data Browsing and CRUD
- [x] `browse_rows` with server-side pagination, sort, basic filters
- [x] Safe default select limit (configurable)
- [ ] Row view: type-aware renderers, null display, copy/download
- [x] `insert_row`, `update_row`, `delete_row` with transaction wrapping
- [x] Enforce safe mode confirmations on destructive ops

### Phase 6: SQL Console
- [ ] Textarea editor with run/cancel, transactional toggle
- [ ] `sql_execute` and `sql_explain` actions with multi-statement support
- [ ] Result tabs; CSV/JSON download of results
- [ ] Saved queries per user/connection: list, save, delete

### Phase 7: DDL Helpers
- [ ] Table editor payload schema (create/alter/drop)
- [ ] Preview SQL before apply; confirm under safe mode
- [ ] Indexes and constraints management (MVP subset)

### Phase 8: Import/Export
- [ ] Export: CSV (streamed), JSON (NDJSON/array), SQL INSERTs (+ optional CREATE TABLE)
- [ ] Import: CSV preview, column mapping, type coercion, transactional apply
- [ ] Limits: `MAX_UPLOAD_SIZE_MB`, row limits, and per-row error report

### Phase 9: UI/UX and Theming
- [ ] Base layout with breadcrumbs, global quick search, responsive grid
- [ ] Dark mode toggle; persist preference in session/local storage
- [ ] Keyboard shortcuts: run query, save, paginate
- [ ] Error surfaces with copy-to-clipboard diagnostics
- [ ] Minimal theming API (title, logo, colors) + full override path

### Phase 10: Observability and Ops
- [x] Structured JSON logs with request IDs; redact secrets
- [x] Health endpoints: `/healthz`, `/readyz` (ping active connection)
- [ ] Audit events: who, action, object, rows affected (avoid PII); pluggable `AuditSink`

### Phase 11: Packaging & Deployment
- [x] Build the thin binary in `cmd/server` mounting `NewHandler`
- [ ] Dockerfile (distroless or slim), non-root, healthcheck
- [ ] docker-compose with sample Postgres/MySQL for demo
- [ ] Reverse proxy examples (Caddy/Traefik/Nginx) with TLS

### Phase 12: Testing Strategy
- [ ] Unit tests: metadata discovery, query exec, safety rules
- [ ] Integration tests: containers for Postgres/MySQL; SQLite in-memory
- [ ] UI smoke tests (Playwright/Cypress): connect, browse, CRUD, SQL, import/export
- [ ] Security tests: CSRF, auth flows, RBAC, safe mode enforcement

### Phase 13: Documentation
- [ ] README with quick start, config table, examples (embed + standalone)
- [ ] SECURITY.md: threat model, safe mode, secrets handling
- [ ] OPERATIONS.md: metrics, logs, backups for profiles/queries
- [ ] DRIVER_NOTES.md: dialect quirks, feature coverage matrix

### Phase 14: Versioning & Release
- [ ] Tag v0.x pre-releases; document breaking changes
- [ ] SemVer policy and changelog
- [ ] Publish module; CI pipeline for tests/lint/build/release

### Phase 15: Hardening and Nice-to-Haves
- [ ] Read-only mode enforcement irrespective of DB grants
- [ ] Query plan visualization (where supported)
- [ ] ER diagram (read-only FK graph)
- [ ] Data diff helpers across tables/environments
- [ ] SSH tunnel support per profile (optional)
- [ ] Read-only share links for saved queries

### Acceptance Checklist (Traceability to Criteria)
- [x] Multi-DB connections; browse schemas/tables/views
- [ ] CRUD rows; run SQL; import/export; basic DDL
- [ ] Secure auth, CSRF protection, safe mode on by default
- [x] Works with PostgreSQL, MySQL/MariaDB, SQLite out of the box
- [ ] Responsive UI with embedded CSS and dark mode
- [ ] Structured logs, basic metrics, health endpoints
- [x] JSON responses use `github.com/dracory/api` envelope
- [ ] Docker image + compose; example works end-to-end
- [ ] Strong unit/integration coverage; UI smoke tests green
- [x] Published as a reusable Go module with a single `http.Handler` and `Register` helper