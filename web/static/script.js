import Fuse from "https://cdn.jsdelivr.net/npm/fuse.js@7.1.0/dist/fuse.basic.min.mjs";

const nav_toggle = document.getElementById("nav-toggle");
const nav = document.getElementById("primary-nav")

let toggled = false;

nav_toggle?.addEventListener("click", () => {
    toggled = !toggled;
    nav_toggle.textContent = toggled ? "×" : "≡";
    nav.style.display = toggled ? "block" : "none";
    // todo: add event when doc is clicked to toggle of nav on mobile
});

document.body.addEventListener("click", event => {
    const button = event.target.closest(".search-toggler");
    if (!button)
        return;

    toggleSearch()
});

document.body.addEventListener("click", event => {
    const textArea = event.target.closest("input");
    if (!textArea)
        return;

    this.style.height = "auto";
    this.style.height = this.scrollHeight + "px";
});

document.body.addEventListener("click", event => {
    const dt = event.target.closest("[data-target]");
    if (!dt) return;
    const btn = document.getElementById(dt.dataset.target);
    if (!btn) return;
    btn.focus();
    btn.classList.add("focused");
    btn.addEventListener("blur", () => btn.classList.remove("focused"), { once: true });
});

document.body.addEventListener("click", event => {
    const button = event.target.closest(".removeRowBtn");
    if (!button)
        return;

    button.closest(".extracted-grocery-row").remove();
});

document.body.addEventListener("click", event => {
    const button = event.target.closest(".quick-btn");
    if (!button)
        return;

    document.getElementById("product-input").value = button.textContent;
});


const recipeCards = [...document.querySelectorAll(".recipe-card")];
const searchInput = document.getElementById("fuzzy-search");
const matchInput = document.getElementById("fuzzy-match");

if (recipeCards.length > 0 && searchInput && matchInput) {
    const data = recipeCards.map(card => ({
        element: card,
        title: card.dataset.title || "",
        ingredients: [...card.querySelectorAll("li")].map(li => li.textContent.trim()),
    }));

    console.log(data)
    const fuseTitle = new Fuse(data, {
        keys: ["title"],
        threshold: 0.5
    });

    searchInput.addEventListener("input", e => {
        const q = e.target.value.trim();

        if (!q) {
            recipeCards.forEach(c => c.hidden = false);
            return;
        }

        recipeCards.forEach(c => c.hidden = true);

        fuseTitle.search(q).forEach(r => {
            r.item.element.hidden = false;
        });
    });

    matchInput.addEventListener("input", e => {
        const q = e.target.value.trim();

        if (!q) {
            recipeCards.forEach(card => card.hidden = false);
            return;
        }

        recipeCards.forEach(c => c.hidden = true);

        const tokens = q.toLowerCase().split(/\s+/).filter(Boolean);

        recipeCards.forEach(card => {
            const ingredients = [...card.querySelectorAll("li")]
                .map(li => li.textContent.trim().toLowerCase());

            const allTokensFound = tokens.every(token => {
                const fuse = new Fuse(ingredients, {
                    threshold: 0.2,
                    ignoreDiacritics: true,
                });

                return fuse.search(token).length > 0;
            });

            card.hidden = !allTokensFound;
        });
    });
}

function toggleSearch() {
    document.querySelector(".search-section").classList.toggle("hidden");
    focusables.forEach(e => {
        e.focus();
    });
}
