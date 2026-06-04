document.addEventListener("click", (e) => {
  const iw = e.target.closest(".image-wrapper")
  if (!iw) {
    return
  }
  const isThumb = iw.classList.contains("thumb")
  if (isThumb) {
    const imgs = iw.querySelectorAll("img")
    const src = iw.dataset.full
    iw.classList.remove("thumb")
    if (src) {
      imgs.forEach((img) => {
        img.src = src
      })
    }
  } else {
    const imgs = iw.querySelectorAll("img")
    const src = iw.dataset.thumb
    iw.classList.add("thumb")
    if (src) {
      imgs.forEach((img) => {
        img.src = src
      })
    }
  }
  document.dispatchEvent(new CustomEvent("lrc:scrollIfAttached"))
})
