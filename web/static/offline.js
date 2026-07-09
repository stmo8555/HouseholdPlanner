// Shared state + sync engine for offline shopping mode. The offline view
// itself lives in offline-list.html / offline-view.js; this module holds the
// pieces both worlds need, plus the wiring for normal (online, SSR) pages.

export const QUEUE_KEY = "hp.offline.queue";
export const SNAPSHOT_KEY = "hp.offline.snapshot";

const MAX_TRIES = 3;

export function getQueue() {
    try {
        return JSON.parse(localStorage.getItem(QUEUE_KEY)) || {};
    } catch {
        return {};
    }
}

export function setQueue(queue) {
    localStorage.setItem(QUEUE_KEY, JSON.stringify(queue));
}

export function queueSize() {
    return Object.keys(getQueue()).length;
}

export function queueSet(listId, itemId, picked) {
    const queue = getQueue();
    queue[`${listId}:${itemId}`] = { picked, ts: Date.now(), tries: 0 };
    setQueue(queue);
}

export function queueDelete(listId, itemId) {
    const queue = getQueue();
    delete queue[`${listId}:${itemId}`];
    setQueue(queue);
}

export function getSnapshot() {
    try {
        return JSON.parse(localStorage.getItem(SNAPSHOT_KEY));
    } catch {
        return null;
    }
}

export function setSnapshot(snapshot) {
    localStorage.setItem(SNAPSHOT_KEY, JSON.stringify(snapshot));
}

export function clearSnapshot() {
    localStorage.removeItem(SNAPSHOT_KEY);
}

function csrfToken() {
    const meta = document.querySelector('meta[name="csrf-token"]');
    return meta ? meta.content : "";
}

export function toast(message) {
    let container = document.getElementById("offline-toast-container");
    if (!container) {
        container = document.createElement("div");
        container.id = "offline-toast-container";
        container.className = "offline-toast-container";
        container.setAttribute("aria-live", "polite");
        document.body.appendChild(container);
    }
    const el = document.createElement("div");
    el.className = "toast grocery-add-toast-msg";
    el.setAttribute("role", "status");
    el.textContent = message;
    el.addEventListener("animationend", () => el.remove());
    container.replaceChildren(el);
}

// --- Sync engine ---

async function refreshCsrfToken() {
    try {
        const resp = await fetch("/groceries", { cache: "no-store" });
        if (!resp.ok) return false;
        const html = await resp.text();
        const fresh = html.match(/name="csrf-token" content="([^"]*)"/);
        const meta = document.querySelector('meta[name="csrf-token"]');
        if (fresh && meta) {
            meta.content = fresh[1];
            return true;
        }
    } catch {
        // network gone again
    }
    return false;
}

function sendPicked(listId, itemId, picked) {
    return fetch(`/groceries/lists/${listId}/items/${itemId}/picked`, {
        method: "PATCH",
        headers: {
            "HX-Request": "true",
            "X-Offline-Sync": "1",
            "X-CSRF-Token": csrfToken(),
        },
        body: new URLSearchParams({ picked: String(picked) }),
    });
}

let syncing = false;

// Replays the queue oldest-first. Returns "done" when the queue drained,
// "auth" when the session expired, "network" when connectivity dropped.
export async function syncQueue() {
    if (syncing) return "busy";
    syncing = true;
    try {
        const queue = getQueue();
        const keys = Object.keys(queue).sort((a, b) => queue[a].ts - queue[b].ts);
        let csrfRetried = false;

        for (const key of keys) {
            const [listId, itemId] = key.split(":");
            const entry = queue[key];

            let resp;
            try {
                resp = await sendPicked(listId, itemId, entry.picked);
                if (resp.status === 403 && !csrfRetried) {
                    csrfRetried = true;
                    if (await refreshCsrfToken()) {
                        resp = await sendPicked(listId, itemId, entry.picked);
                    }
                }
            } catch {
                setQueue(queue);
                return "network";
            }

            if (resp.ok) {
                delete queue[key];
                setQueue(queue);
            } else if (resp.status === 401) {
                setQueue(queue);
                return "auth";
            } else {
                entry.tries = (entry.tries || 0) + 1;
                if (entry.tries >= MAX_TRIES) delete queue[key];
                setQueue(queue);
            }
        }
        return "done";
    } finally {
        syncing = false;
    }
}

// --- Entering offline mode (from a list page) ---

function snapshotFromDom(listId) {
    const grid = document.querySelector(".grocery-grid");
    if (!grid) return null;

    const title = document.querySelector(".grocery-list h2");
    const listName = title
        ? title.textContent.replace(/\s*\(\d+\)\s*$/, "").trim()
        : "Grocery list";

    const items = [];
    let category = "Other";
    let picked = false;
    for (const el of grid.children) {
        if (el.classList.contains("grocery-divider")) {
            picked = el.classList.contains("picked");
            if (!picked) {
                category = el.textContent.replace(/\s*\(\d+\)\s*$/, "").trim() || "Other";
            }
            continue;
        }
        if (!el.classList.contains("grocery-row") || el.classList.contains("grocery-header")) continue;
        const id = Number(el.dataset.id);
        if (!id) continue;
        const field = name => el.querySelector(`[data-field="${name}"]`)?.textContent.trim() || "";
        items.push({ id, name: field("name"), brand: field("brand"), amount: field("amount"), category, picked });
    }

    return { listId: Number(listId), listName, items, savedAt: Date.now() };
}

async function enterOfflineMode(listId) {
    let snapshot = null;
    try {
        const resp = await fetch(`/groceries/lists/${listId}/snapshot`, { cache: "no-store" });
        if (resp.ok) {
            snapshot = await resp.json();
            snapshot.savedAt = Date.now();
        }
    } catch {
        // no network: fall back to what's already on the page
    }
    if (!snapshot) snapshot = snapshotFromDom(listId);
    if (!snapshot) {
        toast("Couldn't start offline mode");
        return;
    }

    // Remember the server state so a pick that gets unpicked again cancels
    // out instead of staying queued as a pending change.
    snapshot.items.forEach(item => { item.serverPicked = item.picked; });

    setSnapshot(snapshot);
    window.location = "/offline";
}

// --- Wiring for normal pages ---

export function initOffline() {
    if ("serviceWorker" in navigator) {
        navigator.serviceWorker.register("/sw.js");
    }

    document.body.addEventListener("click", event => {
        const btn = event.target.closest("#shop-offline-btn");
        if (btn) enterOfflineMode(btn.dataset.listId);
    });

    document.body.addEventListener("submit", event => {
        if (!event.target.closest(".logout-form")) return;
        localStorage.removeItem(QUEUE_KEY);
        clearSnapshot();
    });

    // Leftover check-offs from an interrupted offline session (e.g. a login
    // redirect mid-sync): drain silently now that we're back on a normal page.
    if (queueSize() > 0 && navigator.onLine) {
        syncQueue().then(result => {
            if (result === "done") {
                clearSnapshot();
                if (location.pathname.startsWith("/groceries")) location.reload();
            }
        });
    }
}
