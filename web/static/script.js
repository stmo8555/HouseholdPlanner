if ("serviceWorker" in navigator) {
    navigator.serviceWorker.register("/sw.js");
}

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

    const grid = button.closest(".extracted-grocery-grid");
    button.closest(".extracted-grocery-row").remove();

    const remaining = grid.querySelectorAll(".extracted-grocery-row:not(.grocery-header)");
    if (remaining.length === 0) {
        document.getElementById("extract-cancel")?.click();
    }
});

document.body.addEventListener("click", event => {
    const button = event.target.closest(".quick-btn");
    if (!button)
        return;

    document.getElementById("product-input").value = button.textContent;
});
