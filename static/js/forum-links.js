document.addEventListener("click", (e) => {
  const fl = e.target.closest(".forum-link")
  if (fl) {
    href = fl.getAttribute("data-link")
    window.location.assign(href)
  }
})
