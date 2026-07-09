// Renders the offline shopping view (offline-list.html) from the snapshot
// taken when the user pressed "Shop offline" on a list page. Check-offs
// update the snapshot and the sync queue; nothing here talks to the server
// until "Go online & sync".

import {
    getQueue, queueSet, queueDelete, queueSize, getSnapshot, setSnapshot,
    clearSnapshot, syncQueue, toast,
} from "./offline.js";

const CATEGORY_ORDER = ["Dairy", "Fruit & veg", "Meat & fish", "Frozen", "Pantry", "Other"];
const DIVIDER_CLASS = {
    "Dairy": "dairy",
    "Fruit & veg": "fruit",
    "Meat & fish": "meat",
    "Frozen": "frozen",
    "Pantry": "pantry",
    "Other": "other",
};

function el(tag, className, text) {
    const node = document.createElement(tag);
    if (className) node.className = className;
    if (text !== undefined) node.textContent = text;
    return node;
}

function renderRow(item, pendingKeys) {
    const row = el("div", "grocery-row");
    row.dataset.id = item.id;
    if (pendingKeys.has(String(item.id))) row.classList.add("offline-pending");

    const btn = el("button", "offline-check");
    btn.type = "button";
    btn.setAttribute("aria-label", (item.picked ? "Unpick " : "Pick ") + item.name);
    btn.appendChild(el("div", "checkbox" + (item.picked ? " checked" : "")));
    row.appendChild(btn);

    const lineThrough = item.picked ? " line-through" : "";
    for (const [field, value] of [["name", item.name], ["amount", item.amount], ["brand", item.brand]]) {
        const span = el("span", lineThrough.trim() || null, value);
        span.dataset.field = field;
        row.appendChild(span);
    }
    row.appendChild(el("div"));
    return row;
}

function render() {
    const snapshot = getSnapshot();
    const empty = document.getElementById("offline-empty");
    const old = document.getElementById("offline-grocery-list");
    if (old) old.remove();

    if (!snapshot || !snapshot.items) {
        empty.hidden = false;
        updateHeaderCount();
        return;
    }
    empty.hidden = true;

    const queue = getQueue();
    const pendingKeys = new Set(
        Object.keys(queue)
            .filter(k => k.startsWith(`${snapshot.listId}:`))
            .map(k => k.split(":")[1]));

    const unpicked = snapshot.items.filter(i => !i.picked);
    const picked = snapshot.items.filter(i => i.picked);

    const section = el("section", "glass container grocery-list");
    section.id = "offline-grocery-list";
    section.appendChild(el("h2", null, `${snapshot.listName} (${unpicked.length})`));

    const grid = el("div", "grocery-grid");
    section.appendChild(grid);

    for (const category of CATEGORY_ORDER) {
        const items = unpicked.filter(i => (CATEGORY_ORDER.includes(i.category) ? i.category : "Other") === category);
        if (!items.length) continue;
        const divider = el("div", `grocery-divider ${DIVIDER_CLASS[category]}`);
        divider.appendChild(el("strong", null, `${category} (${items.length})`));
        grid.appendChild(divider);
        items.forEach(item => grid.appendChild(renderRow(item, pendingKeys)));
    }

    if (picked.length) {
        const divider = el("div", "grocery-divider picked");
        divider.appendChild(el("strong", null, `Picked (${picked.length})`));
        grid.appendChild(divider);
        picked.forEach(item => grid.appendChild(renderRow(item, pendingKeys)));
    }

    document.getElementById("offline-main").appendChild(section);
    updateHeaderCount();
}

function updateHeaderCount() {
    const count = queueSize();
    document.getElementById("offline-queue-count").textContent = count;
    document.getElementById("offline-queue-label").hidden = count === 0;
}

function toggleItem(id) {
    const snapshot = getSnapshot();
    if (!snapshot) return;
    const item = snapshot.items.find(i => i.id === id);
    if (!item) return;

    item.picked = !item.picked;
    setSnapshot(snapshot);
    if (item.picked === item.serverPicked) {
        // Back to the state the server already has: nothing to sync.
        queueDelete(snapshot.listId, id);
    } else {
        queueSet(snapshot.listId, id, item.picked);
    }
    render();
}

async function exitAndSync() {
    const btn = document.getElementById("offline-exit-btn");
    btn.disabled = true;
    const result = await syncQueue();
    btn.disabled = false;

    if (result === "network") {
        toast("Still offline — keep shopping");
        return;
    }
    if (result === "auth") {
        window.location = "/login";
        return;
    }
    if (result === "busy") return;

    const snapshot = getSnapshot();
    clearSnapshot();
    window.location = snapshot ? `/groceries/lists/${snapshot.listId}` : "/groceries";
}

if ("serviceWorker" in navigator) {
    navigator.serviceWorker.register("/sw.js");
}

document.body.addEventListener("click", event => {
    const check = event.target.closest(".offline-check");
    if (check) {
        const row = check.closest(".grocery-row");
        if (row) toggleItem(Number(row.dataset.id));
        return;
    }
    if (event.target.closest("#offline-exit-btn")) exitAndSync();
});

render();
