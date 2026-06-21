# Production Launch Checklist

## Security (do before going live)

- [ ] **Set Gin to release mode** — add `GIN_MODE=release` to `.env`. Debug mode prints every route on startup and leaks info.
- [ ] **Change DB password** — `.env` still has `POSTGRES_USER: Admin`. Change the password to something strong. Note: editing `POSTGRES_PASSWORD` in `.env` does **nothing** to an existing volume — Postgres only reads it on first init. Rebuilding the volume (`docker compose down -v && docker compose up`) to re-seed is the only thing that applies it.
- [ ] **Verify OPENAI_API_KEY is in `.env`** — AI features (smart add, extract from recipe) will silently fail or error without it.
- [ ] **Add CSRF protection for forms/HTMX mutations** — partly mitigated already: the session cookie is `SameSite=Strict`, which blocks cross-site form posts. Add explicit tokens only if you relax SameSite later.
- [ ] **Change logout from `GET` to `POST`** — a `GET /logout` can be triggered by any embedded resource or prefetch. Make it a `POST` (and update the nav link to a small form/HTMX post).
- [ ] **Fix grocery household/list integrity checks** — confirm every grocery list/item endpoint scopes by `household_id` so one household can't read or mutate another's data via a guessed `:id`.
- [ ] **Harden recipe URL extraction against SSRF, timeouts, and huge responses** — the recipe routes are currently disabled in `main.go`; before re-enabling, validate/allowlist the URL, block internal addresses, and cap request timeout and response size.
- [x] **Add rate limiting to POST /login** — done: per-IP in-memory limiter (3 attempts / 10s, idle entries cleaned up every 5m) in `internal/login/handler.go`, with `TrustedPlatform = gin.PlatformCloudflare` so the real client IP is used behind the tunnel.
- [x] **Fix wrong-password login bug** — done.
- [x] **Don't publish Postgres to the host** — done: `db` now binds `127.0.0.1:5432:5432`, so Postgres is not exposed on the LAN.
- [x] **Bind the app to localhost only** — done: `app` binds `127.0.0.1:8080:8080`; reachable only through the Cloudflare tunnel.

## Deployment / Reliability

- [ ] **Cloudflare tunnel configured and running** — confirm `cloudflared` is installed on the Pi, tunnel is set up pointing to `localhost:8080`, and `cloudflared` is a systemd service so it starts on boot. Keep the tunnel as the **only** public entrypoint, and add a `restart` policy for `cloudflared` too.
- [ ] **Remove any router port-forwards for `8080` and `5432`** — with the tunnel as the sole entrypoint, no inbound ports should be forwarded to the Pi.
- [ ] **Confirm Cloudflare passes HTTPS through** — the session cookie is `Secure`-only, so it is never sent over plain HTTP. This is fine behind the tunnel (browser→Cloudflare is HTTPS), but if you ever test via `http://<pi-ip>:8080` directly, login will appear broken because the cookie won't be stored. Log in over the *public* URL and confirm the session sticks.
- [ ] **Add `/healthz` and an app Docker healthcheck** — the `db` service has a healthcheck; add a lightweight `/healthz` route and a healthcheck on the `app` service so Compose/monitoring can tell when it's actually serving.
- [ ] **Pin Docker images instead of using `latest`** — pin the app's base image and keep `postgres:18` (already pinned to a major) so rebuilds are reproducible.
- [ ] **Keep the ARM64 Docker build** — the Raspberry Pi is 64-bit; make sure the image builds/runs for `linux/arm64`.
- [ ] **Confirm the app is running on the Pi** — run `docker compose up -d` and test the public URL end-to-end before declaring it live.
- [x] **Add Docker restart policy** — done: `restart: unless-stopped` on both `app` and `db`, so they survive a Pi reboot.

## Data / Users

- [ ] **Verify your user passwords** — `steffe` and `anna` are seeded with bcrypt hashes in `init.sql`. Make sure you know the plaintext passwords. To reset one: generate a new bcrypt hash and `UPDATE users SET pwd = '...' WHERE username = '...'`.
- [ ] **Plan how to add future users** — there's no admin UI. For now: `INSERT INTO users (username, pwd) VALUES ('name', '<bcrypt-hash>')` + `INSERT INTO household_members (user_id, household_id) VALUES (...)`.

## Operational

- [ ] **Off-device nightly DB backups** — Postgres data lives in a Docker volume. Cron on the Pi: `docker exec household-planner-db pg_dump -U Admin db > backup.sql`, rotate weekly, and copy the dump **off the Pi** (an SD card failure shouldn't take the backups with it).
- [ ] **Document and verify DB restore** — a backup you've never restored isn't a backup. Write down the restore steps for a fresh Raspberry Pi, and test them once by restoring a dump into a throwaway container (`docker run --rm postgres:18` + `psql < backup.sql`).
- [ ] **Docker log rotation** — add to **both** the `app` and `db` services in `docker-compose.yml` to cap log disk usage:
  ```yaml
  logging:
    driver: "json-file"
    options:
      max-size: "10m"
      max-file: "3"
  ```

## Code quality / correctness

- [ ] **Replace handler panics/raw errors with safe responses** — repos now return domain `ErrNotFound` (good start), but most handlers still `panic` on error. Map errors to proper HTTP responses instead of relying on the recovery middleware.
- [ ] **Fix frontend JS runtime errors** — clear any console errors in the served pages.
- [ ] **Disable or finish Todo Smart Add** — the `todos` routes are commented out in `main.go`; either finish the feature or keep it disabled rather than shipping a half-wired endpoint.

## Project hygiene

- [ ] **Add `.env.example`** — document required variables (DB creds, `OPENAI_API_KEY`, `GIN_MODE`) without committing secrets.
- [ ] **Add `.dockerignore`** — keep the build context small (no `.git`, local artifacts, etc.).
- [ ] **Fill out the README** — Pi deploy, tunnel setup, backup, and restore instructions.
- [ ] **Add CI** — run `go build ./...` and `go test ./...` on push.
- [ ] **Add reliable tests** — that do not depend on live recipe websites (the current recipe test hits the network / panics on a nil extractor).

## Nice-to-have (not blocking)

- [ ] **Move Postgres data from SD card to USB SSD** — better durability and performance for the DB volume; can be done after launch.
- [x] **Decide on `/todos` visibility** — route and nav link removed, groceries only.

## Can ignore

- [ ] **Manual TLS certificate setup on the Pi** — not needed; the Cloudflare tunnel terminates public HTTPS.
