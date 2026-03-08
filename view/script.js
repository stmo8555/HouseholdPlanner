const modal = document.getElementById("choreModal");
const modalContent = document.getElementById("modalContent");
const closeBtn = modal.querySelector(".close-btn");

document.querySelectorAll(".chore-card").forEach(card => {
  card.addEventListener("click", () => {
    // get the card title and details
    const title = card.querySelector("h2").textContent;
    const details = card.querySelector(".chore-details").innerHTML;

    // insert into modal
    modalContent.innerHTML = `<h2>${title}</h2>${details}`;

    // show modal
    modal.classList.add("active");
    document.body.style.overflow = "hidden"; // lock background scroll
  });
});

closeBtn.addEventListener("click", () => {
  modal.classList.remove("active");
  document.body.style.overflow = "";
});

modal.addEventListener("click", e => {
  if (e.target === modal) {
    modal.classList.remove("active");
    document.body.style.overflow = "";
  }
});
