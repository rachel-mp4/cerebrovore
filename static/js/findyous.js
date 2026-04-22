(() => {
  const prefix = window.location.origin + "/p/"
  const pl = prefix.length
  const aa = document.querySelectorAll(".body a");
  aa.forEach((a) => {
    const href = a.href
    if (href.startsWith(prefix)) {
      const id = href.slice(pl)
      const tx = document.getElementById(id)
      if (tx && tx.classList.contains("you")) {
        a.classList.add("you")
      }
    }
  })
})()
