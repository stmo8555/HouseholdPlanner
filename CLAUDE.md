# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Development with hot reload (proxy on :8090, app on :8080)
air

# Run directly
go run .

# Build
go build ./...

# Run all tests
go test ./...

# Run tests in a specific package
go test ./internal/grocery/...

# Start just the database
docker compose -f household-db/docker-compose.db.yml up -d

# Start full stack (app + db)
docker compose up
```

The `OPENAI_API_KEY` environment variable must be set for AI features to work. Database credentials are in `.env` (defaults: user=Admin, password=Admin, db=db, host=localhost:5432).

## Architecture

**Tech stack:** Go + Gin (HTTP), PostgreSQL (pgx), HTMX (frontend interactivity), OpenAI GPT-4.1 Nano (AI), Air (hot reload).

**Pattern:** Every domain under `internal/` follows Handler → Service → Repo. All handlers are registered in `main.go`. The auth middleware (`internal/login/middleware.go`) injects `user_id` and `household_id` into Gin's context for every authenticated route — handlers read `household_id` via `c.GetInt("household_id")` to enforce multi-tenancy.

**HTMX rendering:** Handlers return server-rendered HTML partials using `c.HTML(200, "template-name", data)`. Full page loads return a page template; subsequent HTMX interactions hit the same or different endpoints returning partial templates. Many mutating endpoints call a shared `RenderListPartial` helper to return an updated list after the mutation.

**Template naming:** Templates under `web/templates/` use `{{define "name"}}` blocks. Full pages are named like `"groceries.html"`; partials use a `package/name` convention like `"groceries/list"`, `"groceries/edit_modal"`. The `parseTemplates` function in `main.go` walks all `.html` files and registers them by their `{{define}}` name.

**AI structured output:** `internal/ai/client.go` provides a generic `SendStructuredRequest[T]` that generates a JSON Schema from the Go type `T` (via `invopop/jsonschema`) and sends it to the OpenAI Responses API for structured output. New AI extraction tasks follow the same pattern: define a schema struct in `internal/ai/schemas.go`, add a method to `internal/ai/service.go`.

**Product categorization:** When a product is first added, `internal/product/service.go` classifies it using `food_category_lookup.json` (exact match on the normalized name, then a `Frozen`-only single-token pass, then longest-phrase-first matching). Unrecognized products default to `"Other"`. The categories are defined once in `grocery.Categories` (`internal/grocery/model.go`) and are: `Dairy`, `Frozen`, `Pantry`, `Fruit & veg`, `Meat & fish`, `Other` — the values in `food_category_lookup.json` use these exact strings, and `IsValidCategory` gates the category endpoint against the same list. The `GroceriesView` struct groups items by category.

**Category overrides:** A household can regroup a product via `household_product_category`; all grocery reads resolve the effective category with `COALESCE(hpc.category, p.category)`. `Repo.SetCategoryOverride` deletes the row when the chosen category equals the product's default, so a round trip back to the original leaves no override behind. `products.category` is never updated after insert.

**Household versioning:** PL/pgSQL triggers increment `households.version` for grocery and household mutations. This enables optimistic concurrency checks without manual version bumps.

**Currently disabled:** The `home` route group is not wired up in `main.go`.

## Database

Schema is initialized from `household-db/db-init/init.sql` (auto-run on first container start). Default seed users: `steffe` and `anna` (both share household "la casa"). To reset, delete the Docker volume.
