---
name: verify
description: Build, launch, and drive this Go+HTMX app to verify changes at the HTTP surface.
---

# Verifying householdPlanner changes

## Launch

The Postgres container must already be running (`docker compose -f household-db/docker-compose.db.yml up -d`). Then:

```bash
(set -a; . ./.env; set +a; OPENAI_API_KEY="${OPENAI_API_KEY:-dummy}" go run .)
```

Serves on :8080 (no proxy; `air` adds a proxy on :8090 but isn't needed for verification). A dummy `OPENAI_API_KEY` is fine unless driving AI features (smart-add, extract).

## Login and drive

Login form posts `uname`/`pwd`. The password is hashed into the `users` row by `db-init/zz-admin-user.sh` at first container start and never re-read from `.env` afterwards, so it depends on how old your volume is:

- Freshly initialized volume: `ADMIN_USERNAME` / `ADMIN_PASSWORD` from the root `.env` (currently `Admin` / `steffe`).
- The long-lived `household-planner_db_data` volume predates that script and was hand-reset to `Admin` / `testpass123`.

```bash
curl -s -c jar.txt -X POST http://localhost:8080/login -d "uname=Admin&pwd=steffe"
curl -s -b jar.txt http://localhost:8080/groceries
```

A brand-new database has no household, so `/groceries` redirects to `/welcome` until you create one:

```bash
curl -s -b jar.txt -X POST http://localhost:8080/welcome -d "household_name=La Casa"
```

HTMX endpoints return HTML partials. Mutating grocery endpoints expect the `#grocery-list-state` fieldset values — emulate with `-d "sort=product&order=asc"`. Grocery list pages live at `/groceries/lists/:id`; find item IDs via `data-id="N"` in the list HTML.

## Gotchas

- `TestCreateGroceries_RejectsForeignList` and `TestDeleteGrocery_ScopedToHousehold` in `internal/grocery` both fail: the `insertHousehold` helper doesn't supply `households.code`, which is `NOT NULL`. Pre-existing and unrelated to the dev DB — these tests start their own throwaway Postgres, so they fail everywhere until the helper is fixed.
- Static asset changes require bumping `STATIC_CACHE` in `web/static/sw.js` or the PWA serves stale files.
