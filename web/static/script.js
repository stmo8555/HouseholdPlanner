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

const btns = document.querySelectorAll(".removeRowBtn");
btns.forEach((btn) => {
    const parentRow = btn.closest("tr");
    btn.addEventListener("click", () => {
        console.log("remove")
        parentRow.remove();
    });
});

const recipeCards = [...document.querySelectorAll(".recipe-card")];
const searchInput = document.getElementById("fuzzy-search");

if (recipeCards.length > 0 && searchInput) {
    const data = recipeCards.map(card => ({
        element: card,
        title: card.dataset.title || "",
    }));

    const fuse = new Fuse(data, {
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

        fuse.search(q).forEach(r => {
            r.item.element.hidden = false;
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
