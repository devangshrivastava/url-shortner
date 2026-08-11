# Backend TODO and edge cases

This checklist is based on the current Go backend in `backend/`.

## Request validation

- [ ] Validate that `url` is not blank after trimming whitespace.
- [ ] Parse the URL and accept only `http` and `https` schemes.
- [ ] Decide whether localhost, private IPs, userinfo, and non-standard ports are allowed.
- [ ] Add maximum lengths for the long URL, custom alias, expiry value, user-agent, and referer.
- [ ] Reject malformed JSON, empty JSON objects, unknown fields if strict request validation is desired, and excessively large request bodies.
- [ ] Validate custom aliases with a documented format, such as `[A-Za-z0-9_-]+`.
- [ ] Reject reserved aliases such as `analytics`, `shorten`, `health`, and future route names.
- [ ] Decide whether aliases are case-sensitive; normalize them consistently if they are not.
- [ ] Reject expiry values in the past and define the exact behavior when expiry equals the current time.
- [ ] Validate and normalize Unicode, whitespace, and percent-encoded values consistently.

## Idempotency and duplicate requests

- [ ] Add an `Idempotency-Key` request header to `POST /shorten`.
- [ ] Store the idempotency key, request fingerprint, result code, and response status in a database table with a unique constraint.
- [ ] Return the original result when the same key is retried with the same request body.
- [ ] Return `409 Conflict` when the same key is reused with a different request body.
- [ ] Handle retries after client timeouts or network failures without creating multiple short URLs.
- [ ] Define whether identical requests without an idempotency key create a new code or are deduplicated.
- [ ] Add expiration/cleanup rules for stored idempotency keys.
- [ ] Make custom-code creation atomic. The current check-then-insert flow has a race condition.
- [ ] Convert PostgreSQL duplicate-key errors into a stable `409 Conflict` response instead of panicking.
- [ ] Retry generated-code collisions rather than relying on probability or crashing.

## Error handling and HTTP behavior

- [ ] Replace every repository `panic(err)` with returned errors.
- [ ] Prevent a database error while saving a click from crashing the entire server.
- [ ] Return consistent JSON errors, for example `{ "error": "..." }`, for all failures.
- [ ] Distinguish validation errors (`400`), duplicate aliases (`409`), missing codes (`404`), expired codes (`410`), database failures (`500`), and unavailable dependencies (`503`).
- [ ] Return `201 Created` for successful URL creation if that is the chosen API contract.
- [ ] Add a global recovery/error middleware so unexpected handler errors do not terminate the process.
- [ ] Add request timeouts and use the request context instead of `context.Background()` for database operations.
- [ ] Define behavior for database timeouts, canceled client requests, and partial failures.
- [ ] Add a health/readiness endpoint that checks application and database availability.

## Redirect behavior

- [ ] Decide whether redirects should use `302`, `303`, or `307`, and document it.
- [ ] Ensure expired and missing URLs never create click records.
- [ ] Decide whether a click-recording failure should still allow the redirect; it should not block or crash the redirect path.
- [ ] Validate that redirects cannot produce malformed `Location` headers.
- [ ] Decide whether redirecting to `javascript:`, `data:`, `file:`, or other unsafe schemes is forbidden at creation time.
- [ ] Decide how aliases containing `/`, `.`, spaces, Unicode, or route-reserved values behave.
- [ ] Add protection against abusive redirect destinations and malicious URL shortening.

## Analytics and click tracking

- [ ] Return `404` when analytics are requested for a code that does not exist instead of returning zero-valued analytics.
- [ ] Decide whether analytics for expired URLs remain available.
- [ ] Ensure analytics queries return stable timestamp formatting and timezone behavior.
- [ ] Handle nullable database fields safely when scanning IP, user-agent, or referer values.
- [ ] Add pagination or a date range for analytics as click volume grows.
- [ ] Add indexes for `clicks.code` and timestamp-based queries.
- [ ] Decide whether repeated requests, bots, health checks, and prefetchers count as clicks.
- [ ] Define privacy and retention rules for IP addresses, user-agent strings, and referers.
- [ ] Consider IP anonymization and safeguards against forged proxy headers.
- [ ] Limit analytics access if the endpoint should not be public.

## Database, migrations, and startup

- [ ] Use a real migration system instead of relying only on `CREATE TABLE IF NOT EXISTS`.
- [ ] Add the initial SQL migration files under `backend/migrations/`.
- [ ] Change `expires_at` from free-form `TEXT` to a PostgreSQL timestamp type after planning a migration.
- [ ] Add foreign-key/index decisions for click records and URL records.
- [ ] Handle an existing database with an incompatible or partially applied schema.
- [ ] Add startup timeouts and clearer errors for PostgreSQL connection, authentication, and permissions failures.
- [ ] Document that automatic database creation requires the PostgreSQL `CREATEDB` privilege.
- [ ] Prefer a deployment/setup command for database creation when application startup should not have admin privileges.
- [ ] Close the database cleanly on SIGTERM/SIGINT and allow in-flight requests to finish.
- [ ] Use the `PORT` value from `.env` instead of always listening on hardcoded port `8080`.
- [ ] Make the public short-URL base configurable instead of hardcoding `http://localhost:8080`.
- [ ] Make CORS origins configurable and restrict them outside local development.

## Security and abuse prevention

- [ ] Add rate limiting for URL creation, redirects, and analytics.
- [ ] Add request-body and URL-size limits.
- [ ] Protect against alias enumeration if short codes are intended to be private.
- [ ] Add abuse reporting/blocklists or destination safety checks if this is publicly hosted.
- [ ] Avoid logging database credentials, full sensitive URLs, or personal analytics data.
- [ ] Review trusted proxy configuration before using `ClientIP()` in production.
- [ ] Add security headers and consider HTTPS-only deployment.
- [ ] Decide whether authentication or signed/private links will be required later.

## Testing and quality

- [ ] Add service tests for valid URLs, invalid URLs, aliases, duplicate aliases, generated codes, expiry, and idempotency.
- [ ] Add handler tests for JSON validation, status codes, CORS preflight, redirects, errors, and analytics.
- [ ] Add repository/integration tests against PostgreSQL for schema creation, duplicate keys, nulls, and failures.
- [ ] Add concurrency tests for duplicate custom aliases and repeated idempotency keys.
- [ ] Add migration tests for a new database and an existing database.
- [ ] Run `gofmt`, `go vet`, and a race detector in CI.
- [ ] Add API documentation with request/response examples and status-code behavior.
- [ ] Remove or clearly separate unused SQLite and in-memory repositories if PostgreSQL is the only supported backend.
