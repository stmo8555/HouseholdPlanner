# Publish Todos

## Must Do

- [ ] Keep Cloudflare Tunnel as the only public entrypoint.
- [ ] Remove any router port forwards for `8080` and `5432`.
- [ ] Do not expose Postgres publicly in production Compose.
- [ ] Add off-device nightly Postgres backups.
- [ ] Document DB restore steps for a fresh Raspberry Pi.
- [ ] Consider moving Postgres data from SD card to USB SSD later.
- [ ] Add restart policies for app, DB, and `cloudflared`.
- [ ] Add `/healthz` and Docker healthchecks.
- [x] Fix wrong-password login bug.
- [ ] Add login rate limiting.
- [ ] Add CSRF protection for forms/HTMX mutations.
- [ ] Change logout from `GET` to `POST`.
- [ ] Fix grocery household/list integrity checks.
- [ ] Harden recipe URL extraction against SSRF, timeouts, and huge responses.
- [ ] Replace handler panics/raw errors with safe responses.
- [ ] Fix frontend JS runtime errors.
- [ ] Disable or finish Todo Smart Add.
- [ ] Add `.env.example`.
- [ ] Add `.dockerignore`.
- [ ] Pin Docker images instead of using `latest`.
- [ ] Keep ARM64 Docker build since the Raspberry Pi is 64-bit.
- [ ] Fill out README with Pi deploy, tunnel, backup, and restore instructions.
- [ ] Add CI with `go build ./...` and `go test ./...`.
- [ ] Add reliable tests that do not depend on live recipe websites.

## Can Ignore

- [ ] Manual TLS certificate setup on the Pi, since Cloudflare Tunnel handles public HTTPS.
