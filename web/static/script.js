import Fuse from "https://cdn.jsdelivr.net/npm/fuse.js@7.1.0/dist/fuse.basic.min.mjs";


let navToggled = false;

document.body.addEventListener("click", event => {
    const buttonClicked = event.target.closest("#nav-toggle");
    const overlayClicked = event.target.closest(".overlay");
    const nav = document.getElementById("primary-nav");
    const button = document.getElementById("nav-toggle");
    const overlay = document.querySelector(".overlay");
    const toggle = () => {
        navToggled = !navToggled;
        button.classList.toggle("is-open", navToggled);
        nav.style.display = navToggled ? "block" : "none";
        overlay.style.display = navToggled ? "block" : "none";
    };
    if (buttonClicked) {
        toggle(buttonClicked);
    } else if (overlayClicked) {
        toggle();
    }
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

document.body.addEventListener("click", event => {
    const button = event.target.closest(".search-toggler");
    if (!button)
        return;

    toggleSearch()
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


