# Production Launch Checklist

## Security (do before going live)

- [ ] **Set Gin to release mode** — add `GIN_MODE=release` to `.env`. Debug mode prints every route on startup and leaks info.
- [ ] **Change DB password** — `.env` still has `POSTGRES_USER: Admin`. Change the password to something strong. Note: editing `POSTGRES_PASSWORD` in `.env` does **nothing** to an existing volume — Postgres only reads it on first init. Rebuilding the volume (`docker compose down -v && docker compose up`) to re-seed is the only thing that applies it.
- [ ] **Add rate limiting to POST /login** — nothing prevents brute-forcing credentials. A simple in-memory rate limiter (e.g. `golang.org/x/time/rate` keyed by IP) on the login handler is enough for 2 users.
- [ ] **Verify OPENAI_API_KEY is in `.env`** — AI features (smart add, extract from recipe) will silently fail or error without it.
- [ ] **Don't publish Postgres to the host** — `docker-compose.yml` maps `5432:5432` on all interfaces, exposing the DB to anything that can reach the Pi on the LAN. The app reaches Postgres over the internal Docker network, so remove the port mapping entirely, or bind `127.0.0.1:5432:5432` if you want local access.
- [ ] **Bind the app to localhost only** — `docker-compose.yml` maps `8080:8080` on all interfaces. With the Cloudflare tunnel you only need `127.0.0.1:8080:8080`, so the app is reachable *only* through the tunnel (not directly on the LAN). Bonus: the `Secure`-only session cookie is then never served over plain HTTP on the network.

## Deployment / Reliability

- [ ] **Add Docker restart policy** — `docker-compose.yml` has no `restart:` field, so containers stay dead after a Pi reboot. Add `restart: unless-stopped` to both the `app` and `db` services.
- [ ] **Cloudflare tunnel configured and running** — confirm `cloudflared` is installed on the Pi, tunnel is set up pointing to `localhost:8080`, and `cloudflared` is a systemd service so it starts on boot.
- [ ] **Confirm Cloudflare passes HTTPS through** — the session cookie is `Secure`-only, so it is never sent over plain HTTP. This is fine behind the tunnel (browser→Cloudflare is HTTPS), but if you ever test via `http://<pi-ip>:8080` directly, login will appear broken because the cookie won't be stored. Log in over the *public* URL and confirm the session sticks.
- [ ] **Confirm the app is running on the Pi** — run `docker compose up -d` and test the public URL end-to-end before declaring it live.

## Data / Users

- [ ] **Verify your user passwords** — `steffe` and `anna` are seeded with bcrypt hashes in `init.sql`. Make sure you know the plaintext passwords. To reset one: generate a new bcrypt hash and `UPDATE users SET pwd = '...' WHERE username = '...'`.
- [ ] **Plan how to add future users** — there's no admin UI. For now: `INSERT INTO users (username, pwd) VALUES ('name', '<bcrypt-hash>')` + `INSERT INTO household_members (user_id, household_id) VALUES (...)`.

## Operational

- [ ] **DB backup strategy** — Postgres data lives in a Docker volume. Simple cron on the Pi: `docker exec household-planner-db pg_dump -U Admin db > backup.sql`, rotate weekly.
- [ ] **Verify a restore actually works** — a backup you've never restored isn't a backup. Once, restore a dump into a throwaway container (`docker run --rm postgres:18` + `psql < backup.sql`) and confirm the data loads.
- [ ] **Docker log rotation** — add to **both** the `app` and `db` services in `docker-compose.yml` to cap log disk usage:
  ```yaml
  logging:
    driver: "json-file"
    options:
      max-size: "10m"
      max-file: "3"
  ```

## Nice-to-have (not blocking)

- [x] **Decide on `/todos` visibility** — route and nav link removed, groceries only.
