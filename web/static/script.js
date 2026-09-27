if ("serviceWorker" in navigator) {
    navigator.serviceWorker.register("/sw.js");
}

const prefersMotion = () => matchMedia("(prefers-reduced-motion: no-preference)").matches;

document.body.addEventListener("click", event => {
    if (!event.target.closest("#nav-toggle, .overlay"))
        return;

    const header = document.querySelector("header");
    const open = !header.classList.contains("nav-open");
    header.classList.toggle("nav-open", open);
    const button = document.getElementById("nav-toggle");
    button.classList.toggle("is-open", open);
    button.setAttribute("aria-expanded", open);
});

document.body.addEventListener("click", event => {
    const dialog = event.target.closest("dialog");
    if (!dialog)
        return;

    var rect = dialog.getBoundingClientRect();
    var isInDialog = (rect.top <= event.clientY && event.clientY <= rect.top + rect.height &&
        rect.left <= event.clientX && event.clientX <= rect.left + rect.width);
    if (!isInDialog) {
        dialog.close();
    }
});

document.body.addEventListener("input", event => {
    const textArea = event.target.closest("textarea");
    if (!textArea)
        return;

    textArea.style.height = "auto";
    textArea.style.height = textArea.scrollHeight + "px";
});

document.body.addEventListener("click", event => {
    const dt = event.target.closest(".empty-state-help-item[data-target]");
    if (!dt) return;
    const btn = document.getElementById(dt.dataset.target);
    if (!btn) return;
    btn.click()
});

function collapse(element) {
    if (!prefersMotion())
        return Promise.resolve();

    const height = `${element.offsetHeight}px`;
    element.style.overflow = "hidden";
    return element.animate([
        { opacity: 1, transform: "none", height, minHeight: "0px", easing: "ease-in" },
        { opacity: 0, transform: "translateX(32px)", height, minHeight: "0px", offset: 0.4, easing: "cubic-bezier(0.22, 1, 0.36, 1)" },
        { opacity: 0, transform: "translateX(32px)", height: "0px", minHeight: "0px", paddingBlock: "0px", borderBottomWidth: "0px" },
    ], { duration: 420 }).finished;
}

document.body.addEventListener("click", async event => {
    const button = event.target.closest(".removeRowBtn");
    if (!button)
        return;

    const grid = button.closest(".extracted-grocery-grid");
    const row = button.closest(".extracted-grocery-row");
    button.disabled = true;
    await collapse(row);
    row.remove();

    const remaining = grid.querySelectorAll(".extracted-grocery-row:not(.grocery-header)");
    if (remaining.length === 0) {
        document.getElementById("extract-cancel")?.click();
    }
});

document.body.addEventListener("click", async event => {
    const button = event.target.closest("[data-paste-target]");
    if (!button)
        return;

    const input = document.querySelector(button.dataset.pasteTarget);
    try {
        input.value = (await navigator.clipboard.readText()).trim();
    } catch (_) { }
    input.focus();
});

document.addEventListener("close", event => {
    const dialog = event.target;
    if (dialog.parentElement?.id !== "modal-root")
        return;

    setTimeout(() => dialog.open || dialog.remove(), 300);
}, true);

const readRows = () => new Map([...document.querySelectorAll(".grocery-row[data-id]")]
    .map(row => [row.dataset.id, row.querySelector(".checkbox.checked") !== null]));
const findRow = id => document.querySelector(`.grocery-row[data-id="${id}"]`);

let knownRows = readRows();
const pendingFlash = new Set();

document.body.addEventListener("htmx:afterSwap", event => {
    const rows = readRows();
    if (event.detail.target.classList.contains("grocery-list")) {
        const dialogOpen = document.querySelector("dialog[open]") !== null;
        for (const [id, picked] of rows) {
            if (!knownRows.has(id)) {
                if (dialogOpen)
                    pendingFlash.add(id);
                else
                    findRow(id).classList.add("is-new");
            } else if (knownRows.get(id) !== picked) {
                findRow(id).classList.add("just-toggled");
            }
        }
    }
    knownRows = rows;
});

document.addEventListener("close", () => {
    pendingFlash.forEach(id => findRow(id)?.classList.add("is-new"));
    pendingFlash.clear();
}, true);

document.body.addEventListener("htmx:beforeTransition", event => {
    document.documentElement.classList.toggle("vt-page", event.detail.target.matches("body, main"));
});

document.body.addEventListener("htmx:beforeSwap", event => {
    document.documentElement.classList.add("intro-done");

    const target = event.detail.target;
    if (target.id === "household-sync-state"
        && target.querySelector(".household-notification")
        && event.detail.serverResponse.includes("household-notification")) {
        event.detail.shouldSwap = false;
    }
});

const SCRAMBLE_CHARS = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789";
const randomChars = length => Array.from({ length },
    () => SCRAMBLE_CHARS[Math.floor(Math.random() * SCRAMBLE_CHARS.length)]).join("");

function scramble(element) {
    const text = element.textContent;
    const steps = 16;
    let step = 0;
    const timer = setInterval(() => {
        step++;
        const settled = Math.floor(text.length * step / steps);
        element.textContent = text.slice(0, settled) + randomChars(text.length - settled);
        if (step === steps)
            clearInterval(timer);
    }, 45);
}

document.body.addEventListener("htmx:afterSettle", () => {
    document.querySelectorAll("[data-scramble]").forEach(element => {
        element.removeAttribute("data-scramble");
        if (prefersMotion())
            scramble(element);
    });
});
