document.addEventListener("click", (e) => {
  const iw = e.target.closest(".image-wrapper")
  if (!iw) {
    return
  }
  const isThumb = iw.classList.contains("thumb")
  if (isThumb) {
    const imgs = iw.querySelectorAll("img")
    const src = iw.getAttribute("data-full")
    if (src) {
      imgs.forEach((img) => {
        img.src = src
      })
      iw.classList.remove("thumb")
    }
  } else {
    const imgs = iw.querySelectorAll("img")
    const src = iw.getAttribute("data-thumb")
    if (src) {
      imgs.forEach((img) => {
        img.src = src
      })
      iw.classList.add("thumb")
    }
  }
})

