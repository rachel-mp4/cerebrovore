(() => {
  const nh = document.getElementById("notification-header")
  const handleNotificationEvent = (data) => {
    if (data.clear !== undefined) {
      return
    }
    if (nh.classList.contains("mail")) {
      return
    }
    nh.classList.add("mail")
    const a = document.createElement("a")
    a.href = "/inbox"
    a.innerText = "you've got mail! click to refresh"
    nh.append(a)
  }

  const nes = cbvWSBuffer?.notification
  if (nes !== undefined) {
    nes.forEach(handleNotificationEvent)
  }

  document.addEventListener("cbv:notification", (e) => {
    handleNotificationEvent(e.detail)
  })
  const ra = document.getElementById("read-all")
  ra.addEventListener("click", () => {
    const lrh = document.getElementById("last-read-hr")
    const lre = document.getElementById("last-read-em")
    if (lrh) {
      const parent = lrh.parentElement
      parent.prepend(lre)
      parent.prepend(lrh)
    } else {
      const parent = document.querySelector(".notifications")
      const lrhr = document.createElement("hr")
      const lrem = document.createElement("em")
      lrem.innerText = "old news onwards!"
      lrhr.id = "last-read-hr"
      lrem.id = "last-read-em"
      parent.prepend(lrem)
      parent.prepend(lrhr)
    }
  })
})()
