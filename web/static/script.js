import Fuse from "https://cdn.jsdelivr.net/npm/fuse.js@7.1.0/dist/fuse.basic.min.mjs";

const form = document.querySelector("form");
const body = document.body;
const focusables = document.querySelectorAll(".focus");

// if (form && body.classList.contains("flower-power")) {
//     form.addEventListener("submit", function(e) {
//         e.preventDefault(); // stop immediate submit
//
//         body.classList.add("slide-out");
//         setTimeout(() => {
//             form.submit(); // now actually submit
//         }, 800); // match CSS transition duration
//     });
// }
const nav_toggle = document.getElementById("nav-toggle");
const nav = document.getElementById("primary-nav")

let toggled = false;

nav_toggle?.addEventListener("click", () => {
    toggled = !toggled;
    nav_toggle.textContent = toggled ? "×" : "≡";
    nav.style.display = toggled ? "block" : "none";
    // todo: add event when doc is clicked to toggle of nav on mobile
});

document.querySelector(".add-toggler")?.addEventListener("click", toggleAddForm)
document.querySelector(".smart-add-toggler")?.addEventListener("click", toggleSmartAdd);
document.querySelector(".search-toggler")?.addEventListener("click", toggleSearch);

const textArea = document.getElementById('smart-add');
textArea?.addEventListener("input", function() {
    this.style.height = "auto";
    this.style.height = this.scrollHeight + "px";
});

document.body.addEventListener("click", event => {
    const button = event.target.closest(".removeRowBtn");
    if (!button)
        return;

    button.closest(".extracted-grocery-row").remove();
});

document.getElementById('extract-button')?.addEventListener("click", () => {
    document.querySelector(".extract-grocery-form").classList.toggle("hidden");
    focusables.forEach(e => {
        e.focus();
    });
});

document.querySelectorAll(".quick-btn").forEach(btn => {
    btn.addEventListener("click", () => {
        document.getElementById("product-input").value = btn.textContent;
    });
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
function toggleSmartAdd() {
    document.querySelector(".smart-add-form").classList.toggle("hidden");
    textArea.style.height = textArea.scrollHeight + "px";
    textArea.style.overflowY = "hidden";
    focusables.forEach(e => {
        e.focus();
    });
}

function toggleAddForm() {
    document.querySelector(".add-form").classList.toggle("hidden");
    focusables.forEach(e => {
        e.focus();
    });
}

function toggleSearch() {
    document.querySelector(".search-section").classList.toggle("hidden");
    focusables.forEach(e => {
        e.focus();
    });
}
