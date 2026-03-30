const form = document.querySelector("form");
const body = document.body;

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
});

