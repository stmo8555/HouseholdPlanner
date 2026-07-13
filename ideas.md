# Ideas

A running list of feature ideas for Household Planner. These are not commitments —
just captured thinking, with rough notes on approach and open questions so we can
pick one up later. DB resets are acceptable, so schema changes are fair game.

---

## 1. Barcode scanning to add items

**What:** Point the phone camera at a product barcode to add it to the current
grocery list, instead of typing the name.

**Why:** Fastest possible "add item" for things you have in hand (restocking the
pantry). Natural fit for a mobile PWA.

**Approach (rough):**
- Frontend: use the native `BarcodeDetector` API where available (Chromium/Android),
  with a vendored fallback library (e.g. ZXing/Quagga) for browsers that lack it.
  Camera access is `getUserMedia` — HTTPS only. Assets get vendored under
  `web/static/vendor/` like htmx/fuse already are.
- Barcode → product name: needs a lookup source. Two options:
  - **Open Food Facts** API (free, has Swedish products) — but that's an outbound
    network call, so it needs the same SSRF-style care and won't work offline.
  - **Learn-as-you-go**: first time a barcode is unknown, ask the user to name it
    and remember `barcode → product` for the household. No external dependency;
    gets smarter over time. Likely the better first version.
- Schema: add a `barcode` column to `products` (or a `product_barcode` mapping
  table if one product can have several).

**Open questions:**
- Native `BarcodeDetector` coverage on the user's actual devices — worth checking
  before committing to it as the primary path.
- Offline behaviour: scanning offline could queue like the existing offline
  pick/unpick sync.

---

## 2. Receipt scanning → price history → recommended price

The bigger idea, in three stages that build on each other.

**Stage A — Scan a receipt to capture prices**
Photograph a grocery receipt; extract each line item's name + price and attach
those costs to the matching products.

- Uses AI vision (image input), which the current setup does **not** do yet: the AI
  client (`internal/ai/client.go`) sends text only, to GPT-4.1-nano via the OpenAI
  Responses API with structured output. Receipt OCR needs an image-capable model and
  a new image-input variant of `SendStructuredRequest`. The structured-output +
  JSON-schema pattern (`internal/ai/schemas.go`) carries over cleanly — just add a
  `ReceiptLine {name, price, quantity}` schema.
- **The hard part is matching**, not OCR. Swedish receipts (ICA, Coop, Willys,
  Hemköp…) use cryptic abbreviations ("BLAND FÄRS", "MJÖLK ARLA 3%"). Mapping those
  to existing `products` rows is the real work — probably a review-before-save
  screen like the current AI extract flow (`extract_review_page.html`), where the
  user confirms/corrects the product each line maps to.

**Stage B — Price history per item**
Once prices are captured, store them over time and show a per-product history
(list or small chart): what you paid, when, and where (store).

- New schema: a `product_prices` table (household_id, product_id, price, currency,
  store, purchased_at, source) — nothing price-related exists in the DB today.
- Surface it from the item edit modal or a dedicated product view.

**Stage C — Recommended / typical price**
From the history, show a suggested price per item (e.g. median, or lowest seen) so
you know if today's price is good.

- Decision needed: **household-only** (just your own purchases) vs **shared/
  crowdsourced** across households (better data, but a privacy + data-model change).
  Start household-only.
- Could later flag "cheaper at store X" if store is captured.

**Suggested sequencing:** A → B → C. Stage A is the prerequisite and the riskiest
(vision + matching); B and C are mostly data modelling + display once prices exist.

**Open questions:**
- Vision model choice and per-scan cost (nano may not be vision-capable; a small
  vision model would be used only on explicit receipt scans, so volume is low).
- How much manual correction is acceptable in the receipt review step.
- Store capture: parse the store from the receipt, or ask once per scan?

---

## Backlog (other ideas discussed, not prioritised)

- **Quantity + unit** — replace the free-text `amount` with numeric qty + unit;
  enables merging duplicates ("2 milk" + "1 milk" → "3 milk").
- **Search / filter within a list** — live type-to-filter (cheapest win, no schema).
- **Undo on delete** — toast-with-undo instead of hard deletes.
- **Meal planning → list** — weekly recipe planner that pushes ingredients onto a
  list; would revive the already-built-but-disabled recipes feature.
- **Item notes** — small free-text note per item.
- **Per-item assignee** — tag who's buying an item.
- **Manual reorder (drag-and-drop)** within a category.
- **Show frequency / last bought** — surface `groceries_history.times_added`, which
  is already tracked but never shown.
