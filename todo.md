# Production Readiness TODO

Findings from a full security/bug/production review (2026-07-09). Ordered by severity.
Complements `PRODUCTION_CHECKLIST.md` — items already tracked there are marked (PC).

## Critical — fix before going live

- [ ] **Re-enable CSRF protection** — `main.go:69-77` is commented out, so no mutating route validates a token (only `SameSite=Strict` on the cookie protects you today; login CSRF is fully open). The bug that broke it is in `CSRFMiddleware` (`main.go:262-268`): when gorilla/csrf rejects a request it writes the 403 but the Gin chain **continues to the route handler anyway**, and on success the wrapped request (which carries the token context) is never propagated back, so `csrf.Token(c.Request)` returns `""`. Fix:
  ```go
  func CSRFMiddleware() gin.HandlerFunc {
      return func(c *gin.Context) {
          ok := false
          CSRFHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
              ok = true
              c.Request = r // propagate token context so csrf.Token works
              c.Next()
          })).ServeHTTP(c.Writer, c.Request)
          if !ok { c.Abort() }
      }
  }
  ```
  Also: require the CSRF secret from env (no `strings.Repeat("a", 32)` fallback), and add a hidden token input to the plain non-HTMX forms (`pages/login.html`, `pages/register.html`, `pages/welcome.html`) — the meta-tag/`htmx:configRequest` approach only covers HTMX requests.
- [ ] **SSRF in recipe extraction** — `internal/ingredient/extractor.go:52` does a raw `http.Get(url)` on a user-supplied link from the **live** route `POST /groceries/lists/:id/extract` (`internal/grocery/handler.go:562-590`, only check is non-empty). Any authenticated user can make the server fetch `http://localhost:5432`, cloud metadata, LAN hosts. Fix: allow only http/https, resolve the host and reject loopback/private/link-local ranges (enforce again in a custom `DialContext` to beat DNS rebinding), use a client with `Timeout: 10s`, cap the body with `io.LimitReader`, and restrict redirects.
- [ ] **`DeleteHousehold`/`DeleteAccount` are broken** — `deleteHouseholdData` (`internal/household/repo.go:227-245`) never deletes from `household_product_category`, which references `households(id)` with no `ON DELETE CASCADE` (`init.sql:72-77`). Any household that ever set a category override (live since the category-edit feature) hits an FK violation → panic/500. Add the DELETE to the `stmts` list (and one for `products`-scoped overrides on account delete if applicable).
- [ ] **Owner can orphan the household (no owner left)** — two live routes:
  - `PromoteMember` (`internal/household/repo.go:123-181`): if `targetUID == uid`, the caller is "promoted" then demoted by the second UPDATE → household has zero owners.
  - `RemoveMember` (`internal/household/repo.go:89-121`): no check that `targetID != callerID` or that the target isn't the owner → owner deletes own membership, bypassing the `ErrOwnerMustTransfer` guard.
  Reject self-target (and/or owner-role targets) in both.
- [ ] **Expired invite tokens work forever** — `ConsumeToken` (`internal/login/repo.go:20-37`) deletes by token with no `expires_at > NOW()` check; only the GET view validates expiry. Add the expiry predicate to the DELETE and add a cleanup job for expired invites (session cleanup exists, invite cleanup doesn't).

## High

- [ ] **`TrustedPlatform = gin.PlatformCloudflare` trusts a spoofable header** — `main.go:79`. Any client that can reach the app directly can set `CF-Connecting-IP` and bypass the per-IP login rate limiter (and bloat its `sync.Map`). Gate it behind an env flag set only when actually fronted by Cloudflare, and configure `SetTrustedProxies` for the real ingress.
- [ ] **No HTTP server timeouts and no graceful shutdown** — `main.go:103` uses `r.Run`. Replace with an explicit `http.Server{ReadHeaderTimeout: 5s, ReadTimeout: 30s, WriteTimeout: 30s, IdleTimeout: 120s}` plus `signal.NotifyContext` + `srv.Shutdown(ctx)` so deploys/`docker compose down` don't kill in-flight requests (and `pool.Close()` actually runs).
- [ ] **Index-out-of-range panic in `SaveExtracted`** — `internal/grocery/handler.go:601-615` indexes `brands[i]`/`amounts[i]` sized by `len(products)`; a form with mismatched field counts crashes the request. Validate the three slices are equal length → 400.
- [ ] **AI endpoints: unbounded input + no rate limit = cost abuse** — `smart-add`/`extract` take `c.PostForm("text")` with no length cap, no request body limit anywhere, and no rate limiting (login is the only limited route). Add a global body-size limit (`http.MaxBytesHandler` or middleware), an explicit input cap (~4–8 KB) before AI calls, and a per-user/household limiter on AI routes.
- [ ] **No deadlines on outbound calls** — OpenAI calls (`internal/ai/client.go:24`) and page fetches (`extractor.go:52`) run with no timeout; callers pass `gin.Context`, whose `Done()/Deadline()` are no-ops by default, so nothing cancels on client disconnect. Pass `c.Request.Context()` wrapped in `context.WithTimeout` (or set `engine.ContextWithFallback = true`).
- [ ] **`FromRecipeURL` ignores the AI error** — `internal/ingredient/extractor.go:60-61`: `err` assigned, never checked; API failure proceeds with a zero-value list. Return the error.
- [ ] **Registration flow bugs** — `internal/login/handler.go`:
  - `RegisterView:85-91`: missing `return` after invalid-token redirect → still renders the register form; redirect target `"login.html"` is a broken relative path (also at :53) — use `/login`.
  - `Register:45-77`: the invite is consumed and the user row created *before* validating household inputs; a malformed POST burns the invite and panics → 500. Validate first, wrap the whole flow in one transaction, return 400 with the form on bad input.
  - No password policy: empty passwords accepted, >72 bytes makes bcrypt error. Enforce a minimum length and reject >72 bytes.
- [ ] **No rate limiting on invite-code join** — `POST /welcome` and `POST /register/:token` accept the 6-char household code unthrottled (55^6 keyspace, but codes never expire). Apply the login limiter to these routes.
- [ ] **No DB migrations** — schema exists only as `init.sql` (runs on first volume init); any post-launch schema change is manual. Adopt `golang-migrate` or `goose`, convert `init.sql` into migration 0001. (PC)
- [ ] **No backups** — DB lives in a Docker volume on the Pi's SD card. Nightly `pg_dump` shipped off-device + one tested restore. (PC)
- [ ] **No indexes on FK/tenant columns** — `init.sql` has zero `CREATE INDEX`. Add indexes on `grocery_lists.household_id`, `groceries.grocery_list_id`, `groceries.product_id`, `sessions.user_id`, `sessions.expires_at`, `invites.household_id`, `todos.household_id`, `recipes.household_id`, `recipe_ingredient.recipe_id`, `restaurants.household_id`.
- [ ] **Dockerfile hardening** — pin base images (`golang:latest`/`alpine:latest` → versions), add a non-root `USER`, replace hardcoded `GOARCH=arm64` with the `TARGETARCH` build arg (currently builds a non-runnable image on amd64). (PC, partially)

## Medium

- [ ] **Replace `panic(err)` as the error path (~40 sites)** across `grocery`, `todo`, `household` handlers: bad `strconv.Atoi` params, bad dates, empty fields, and DB hiccups all become 500s with stack traces. Return 400 for parse/validation failures, 404 for not-found, generic 500 otherwise. Includes the double-write pattern `c.AbortWithStatus(500)` + `c.String(500, ...)` for *empty input* (should be 400) at `grocery/handler.go:158-162`, `todo/handler.go:70-73`.
- [ ] **Internal error text leaked to clients** — `c.String(500, err.Error())` at `todo/handler.go:97,131`, `grocery/handler.go:39`, `notification/handler.go:27,47`, `recipe/handler.go:33,48`. Log server-side, return generic messages.
- [ ] **Background jobs can crash or silently fail** — `login/job.go:8-16` panics inside a cron goroutine (robfig/cron doesn't recover by default → one transient DB error kills the process). Use `cron.WithChain(cron.Recover(...))` and log instead of panicking. Same for `todo/job.go` when re-enabled (errors from `AddFunc`/`ScheduleRepeats` are ignored there too).
- [ ] **`GetID` SELECT-then-INSERT race** — `internal/product/service.go:37-52`: two concurrent adds of a new product → unique-constraint 500. Use `INSERT ... ON CONFLICT (name, brand) DO UPDATE ... RETURNING id`.
- [ ] **`UpdateGrocery` ignores 0 rows affected** — `internal/grocery/repo.go:346-359` + `service.go:161-178`: an update against an ID outside the household returns 200 and still writes a category-override row; the two writes aren't in a transaction. Return `ErrNotFound` on 0 rows; wrap both in one tx.
- [ ] **`TransferGroceries` conflates "not yours" with "empty list"** — `internal/grocery/repo.go:162-186` errors on `RowsAffected()==0`, so transferring an empty list 500s. Verify ownership of both lists explicitly in a tx, then accept 0 rows.
- [ ] **ORDER BY built with `fmt.Sprintf`** — `internal/grocery/repo.go:254-272`. Currently shielded by a caller-side whitelist, but the repo itself is injectable by future callers. Move the column/direction whitelist into the repo.
- [ ] **Invite/auth policy inconsistencies** — any member can mint invite links (`GenerateInviteToken` has no owner check) while regenerating the code is owner-only; pick a policy. Invite links are also built from the raw `Host` header and `c.Request.TLS` (`household/handler.go:180-192`) → Host-header-injected phishing links and `http://` behind the TLS proxy. Use a configured canonical base URL.
- [ ] **Missing `ON DELETE` rules on FKs generally** — beyond the critical one above, `sessions.user_id`, `households.created_by`, `invites.created_by`, etc. have no CASCADE/SET NULL; deletion order is hand-maintained in `deleteHouseholdData`. Declare explicit rules in the schema so the DB enforces it.
- [ ] **Add `/healthz`** (DB `Ping`) + Compose healthcheck for the app service; configure `pgxpool` (`MaxConns`, lifetimes) via `ParseConfig` and `Ping` at startup so bad credentials fail fast instead of per-request. (PC)
- [ ] **Stop logging user/AI content to stdout** — `ai/client.go:42` prints full AI responses; debug prints in `extractor.go:110,202-215`, `todo/service.go:52`. Remove or use a leveled logger. Add Docker `logging` max-size to both compose services (unbounded json-file logs on an SD card). (PC)
- [ ] **Add security headers** — no CSP/HSTS/`X-Content-Type-Options`/`Referrer-Policy` anywhere. Add a small security-headers middleware (or set at the Cloudflare layer). ~~Vendor Fuse.js/htmx/fonts~~ (done 2026-07-09 as part of offline mode: all CDN deps now in `web/static/vendor/`).
- [ ] **Dev DB compose exposes Postgres on 0.0.0.0** — `household-db/docker-compose.db.yml:11`: use `127.0.0.1:5432:5432` and mount `db-init` as `:ro` (main compose already does both).
- [ ] **`OPENAI_API_KEY` not forwarded by docker-compose** — AI features fail in the container. Add it to `.env` + compose `environment`, and fail fast at startup if unset.
- [ ] **Rotate credentials for prod** — `.env` is correctly gitignored and never committed, but `Admin`/`Admin` DB creds and the seeded admin password are weak, and `main.go:212-213` hardcodes the same values as fallbacks. Generate strong secrets, remove the code fallbacks (fail if unset), re-init the volume. `ADMIN_USERNAME`/`ADMIN_PASSWORD` in `.env` appear unused by code — remove or wire up. (PC)
- [ ] **Session cookie `Secure` flag depends on `gin.Mode()`** — `login/handler.go:154-160`. Drive it from explicit config so a prod run without `GIN_MODE=release` doesn't ship cookies over plain HTTP.
- [ ] **Rate limiter is per-IP only and in-memory** — no per-username lockout (credential stuffing on one account is unthrottled) and state resets on restart. Add a per-account counter.
- [ ] **Todo `MarkDone` can duplicate future occurrences** — `todo/service.go:56-63,146-177`: unconditional repo UPDATE + scheduling without checking `NextID != nil`; done→undo→done or a double click creates duplicates. Also `MarkUnDone` ignores its error (`handler.go:104-116`). (Todos route currently disabled — fix before re-enabling.)

## Low / cleanup

- [ ] URL-escape DSN credentials (`main.go:219-222`) — breaks when a strong password contains `@`/`/`/`#`. Use `url.UserPassword` or key/value DSN format.
- [ ] Add `.dockerignore` (build context currently ships `.git`, `tmp/`, mockups). (PC)
- [ ] Add `.env.example` documenting `POSTGRES_*`, `OPENAI_API_KEY`, `GIN_MODE`, `PORT`, CSRF secret. (PC)
- [ ] Delete empty `package.json`/`package-lock.json` (no npm usage).
- [ ] Delete `web/static/assets/household-planner-final-pwa-icon-pack.zip` (244 KB publicly served); compress the 2.6 MB `background-mobile-dark.jpg`.
- [ ] `web/static/script.js:143`: `focusables` is undefined → `ReferenceError` on search toggle; remove `console.log(data)` at :90.
- [ ] Invite tokens travel in GET query strings → they land in `gin.Default()` request logs, history, and Referer. Add `Referrer-Policy: no-referrer` on the register page; consider single-use tokens.
- [ ] `getSession` (`login/repo.go:128-152`): drop the password hash from the SELECT (carried on every request for no reason); the `LEFT JOIN household_members` with `QueryRow` is nondeterministic if a user ever has two memberships — add a unique index on `household_members(user_id)` if single-household is the invariant.
- [ ] `login/middleware.go:31-34`: `ExtendSession` error ignored.
- [ ] `notification/handler.go:31-34`: negative `known_version` always reports changed → client refresh loop. Clamp to ≥ 0.
- [ ] Category vocabulary mismatch: `grocery/service.go:137-150` matches capitalized labels while CLAUDE.md/lookup JSON use lowercase — unify on one canonical constant set.
- [ ] `todo/repo.go:141,181`: `return todos, err` returns a stale nil instead of `rows.Err()`.
- [ ] `bump_household_version` row-level trigger issues N serialized updates on bulk ops (`DeletePicked`, extract save) — consider a statement-level trigger, or accept at current scale.
- [ ] Dead code: `internal/restaurant/` (fully commented out — and contains an unvalidated `http.Get` SSRF if ever revived), `extractor.go:210-222` `findJSON`, `recipe/service.go` `printNode`, unused `product.Service.AIService` field. Delete or branch.

## Before re-enabling disabled routes (todos / recipes / home)

- [ ] `todo` repeat value from the form is never validated (`ValidRepeats` exists but is unreferenced) and an unknown value hits `panic("WHY ARE WE HERE!")` in the scheduler — remotely triggerable process crash once todos + cron are live. Validate on `Add`, replace the panic with a logged skip.
- [ ] `todo.SmartAdd` extracts via AI, `println`s, and persists nothing — silent data loss plus a paid API call. Implement persistence or drop the endpoint.
- [ ] `recipe.Add` inserts recipe and ingredients in separate non-transactional calls with `panic` on parse errors mid-way → orphaned recipe rows; `recipe/repo.go:138` uses `context.Background()` instead of the request ctx.
- [ ] `home.Service.AI` panics on API error, ignores `json.Unmarshal` errors, sets no `MaxOutputTokens`, and builds a new client per request — route through `SendStructuredRequest` before re-enabling.
- [ ] Missing templates: `recipe/handler.go:101` renders `"groceries_extraction.html"`, `home/handler.go:118` renders `"ai_extraction.html"` — neither exists; `pages/index.html:8` references `{{template "header" .}}` but the define is `"layouts/header"`. Re-enabling these routes 500s immediately.
- [ ] Recipe/home pages never set `CSRFToken` in template data — HTMX mutations there will 403 once CSRF is enabled. Consider injecting the token via middleware instead of per-handler.

## Already solid (verified, no action needed)

- bcrypt cost 12; constant-time fake-hash compare + uniform error message on login (no user enumeration/timing leak).
- Server-side UUID session IDs, sliding 36h TTL capped at 30 days, hourly purge, server-side logout; fresh session on every login (no fixation).
- `household_id` re-derived from membership on every request — removed members lose access immediately; all live queries are household-scoped and parameterized (no SQL injection, no IDOR found on live routes).
- Invite/household codes generated with `crypto/rand`, unbiased, ambiguous chars excluded; registration is invite-gated.
- Templates: no XSS found — pure `html/template` auto-escaping, no `template.HTML`, htmx loaded with SRI.
- Main compose binds app and DB to `127.0.0.1`, sets `GIN_MODE=release`, has a DB healthcheck, mounts init `:ro`; `.env`/keys gitignored with clean git history.
