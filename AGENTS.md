# AGENTS.md

## Commands
- `air` runs hot reload: Air builds `go build -o ./tmp/main .`, serves the app on `:8080`, and exposes the proxy on `:8090`.
- `go run .` runs the app directly; `go build ./...` is the fast full compile check.
- `go test ./...` runs all tests.
- Start only Postgres with `docker compose -f household-db/docker-compose.db.yml up -d`; start app plus DB with `docker compose up`.

## Runtime Setup
- DB env defaults are in `main.go`: `POSTGRES_USER=Admin`, `POSTGRES_PASSWORD=Admin`, `POSTGRES_HOST=localhost`, `POSTGRES_PORT=5432`, `POSTGRES_DB=db`, `POSTGRES_SSLMODE=disable`.
- AI paths use the OpenAI client default env loading; set `OPENAI_API_KEY` before using smart add or ingredient extraction.
- DB schema/seed data comes from `household-db/db-init/init.sql` and only runs on first volume creation; reset by removing the Docker volume.
- Seed users are `steffe` and `anna`, both in household `la casa`.

## Architecture
- `main.go` wires global services and routes. Groceries and household settings are the active authenticated areas.
- Domain packages under `internal/` follow Handler -> Service -> Repo. Repos take `pgxpool.Pool`; handlers should not query DB directly.
- Auth middleware sets `user_id` and `household_id` in Gin context; use `c.GetInt("household_id")` for tenant scoping on authenticated routes.
- Product categories are assigned in `internal/product/service.go` from `food_category_lookup.json` by exact match, then token match, else `other`.

## Templates And HTMX
- `parseTemplates("web/templates")` registers templates by `{{define "..."}}`, not by filepath. Keep `c.HTML` names aligned with define names.
- Full pages are named like `groceries.html`; partials use names like `groceries/list`, `groceries/edit_grocery`, and `edit-grocery-list`.
- Mutating HTMX handlers usually return the affected partial, not JSON.
- Grocery edit dialogs are loaded into `#modal-root`; `web/templates/pages/groceries.html` opens any swapped-in `<dialog>` on `htmx:afterSwap`.

## Frontend
- There is no frontend build pipeline; `package.json` is empty. Templates, `web/static/style.css`, and `web/static/script.js` are served directly.
- Frontend libraries and fonts are vendored under `web/static/vendor/`.

## Database Gotchas
- `bump_household_version()` triggers increment `households.version` on grocery and household mutations.
- Grocery history is updated when groceries are inserted and drives quick-select/top-product UI.
