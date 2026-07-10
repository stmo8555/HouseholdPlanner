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

Login form posts `uname`/`pwd`; dev credentials are `Admin` / `testpass123`:

```bash
curl -s -c jar.txt -X POST http://localhost:8080/login -d "uname=Admin&pwd=testpass123"
curl -s -b jar.txt http://localhost:8080/groceries
```

HTMX endpoints return HTML partials. Mutating grocery endpoints expect the `#grocery-list-state` fieldset values — emulate with `-d "sort=product&order=asc"`. Grocery list pages live at `/groceries/lists/:id`; find item IDs via `data-id="N"` in the list HTML.

## Gotchas

- `TestDeleteGrocery_ScopedToHousehold` in `internal/grocery` fails against the current dev DB (missing `code` column default) — pre-existing, not a signal about your change.
- Static asset changes require bumping `STATIC_CACHE` in `web/static/sw.js` or the PWA serves stale files.
