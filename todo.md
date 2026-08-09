# TODO

## High priority

- [ ] Add URL validation. Accept only valid `http` and `https` URLs instead of only checking that the URL is non-empty.
- [ ] Replace database-related `panic(err)` calls with returned errors and proper HTTP 500 responses.
- [ ] Handle concurrent custom-code creation safely. Keep the database unique constraint and handle duplicate-key errors without crashing the server.
- [ ] Handle generated short-code collisions by checking the database and generating another code when necessary.

## Medium priority

- [ ] Use the `PORT` value from `.env` instead of always listening on hardcoded port `8080`.
- [ ] Return HTTP 404 when analytics are requested for a code that does not exist.
- [ ] Add automated tests for shortening, redirects, expiration, duplicate codes, analytics, invalid input, and database failures.

## Configuration and maintenance

- [ ] Document that the PostgreSQL application user requires the `CREATEDB` privilege when automatic database creation is enabled.
- [ ] Run `gofmt` on all Go files to keep formatting consistent.
- [ ] Remove or clearly separate unused SQLite and in-memory repository implementations if PostgreSQL is the only supported backend.
