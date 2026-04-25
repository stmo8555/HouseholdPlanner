const form = document.querySelector("form");
const body = document.body;
const focusables = document.querySelectorAll(".focus");

if (form && body.classList.contains("flower-power")) {
    form.addEventListener("submit", function(e) {
        e.preventDefault(); // stop immediate submit

        body.classList.add("slide-out");
        setTimeout(() => {
            form.submit(); // now actually submit
        }, 800); // match CSS transition duration
    });
}
const nav_toggle = document.getElementById("nav-toggle");
const nav = document.getElementById("primary-nav")

let toggled = false;

nav_toggle.addEventListener("click", () => {
    toggled = !toggled;
    nav_toggle.textContent = toggled ? "×" : "≡";
    nav.style.display = toggled ? "block" : "none";

    // todo: add event when doc is clicked to toggle of nav on mobile
});

document.querySelector(".add-toggler").addEventListener("click", toggleAddForm)
document.querySelector(".smart-add-toggler").addEventListener("click", toggleSmartAdd);

function toggleSmartAdd() {
    const textArea = document.getElementById('smart-add');
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
