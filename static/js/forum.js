(() => {
  window.history.replaceState({}, document.title, window.location.pathname)
})()

document.addEventListener("paste", (e) => {
  const items = e.clipboardData?.items;
  if (items === undefined) {
    return;
  }
  for (const item of items) {
    if (item.type.startsWith("image/")) {
      const blob = item.getAsFile();
      if (blob === null) {
        return;
      }
      e.preventDefault();
      const dt = new DataTransfer()
      dt.items.add(blob)
      const fimg = document.getElementById("ftx-image")
      fimg.files = dt.files
    }
  }

})

document.addEventListener("cbv:thread", (e) => {
  const tse = e.detail
  if (tse.remaining !== undefined) {
    const ls = document.getElementById("left-sidebar")
    if (ls !== null) {
      ls.style.setProperty("--remaining", tse.remaining)
    }
  }
  if (tse.id) {
    const newid = tse.id
    const npid = newid.toString(36)
    const bue = document.getElementById("brains-ur-eat")
    bue.insertAdjacentHTML("beforeend", `<button class="new-post" id="${npid}" hx-get="/fp/${npid}" hx-swap="outerHTML" hx-trigger="click from:.new-posts">new post! click to download</button>`)
    htmx.process(document.getElementById(npid))
    var c = "#452faa"
    if (tse.color) {
      const int = Math.max(Math.min(16777215, Math.floor(tse.color)), 0)
      c = "#" + int.toString(16).padStart(6, '0')
    }
    genpath(window.innerWidth * (Math.random() + Math.random()) / 2, 0, c)
  }
})

// the idea here is that once we obtain a new forum post, we need to client side render
// some aspects of it that are a bit harder to server side render, & we're using trigger 
// after settle to do this
document.addEventListener("cbv:htmxForumPost", (e) => {
  const npid = e.detail.value.toString(36)
  const newfp = document.getElementById(npid)
  const prefix = window.location.origin + "/p/"
  const pl = prefix.length
  const aa = newfp.querySelectorAll(".body a");
  aa.forEach((a) => {
    const href = a.href
    if (href.startsWith(prefix)) {
      const id = href.slice(pl)
      const tx = document.getElementById(id)
      if (tx) {
        const floor = tx.querySelector(".floor")
        if (floor) {
          floor.insertAdjacentHTML("beforeend", ` <a href="/p/${npid}">#${npid}</a>`)
        }
        if (tx.classList.contains("you")) {
          a.classList.add("you")
        }
      }
    }
  })
  const time = newfp.querySelector(".time")
  const ts = time.getAttribute("datetime")
  time.innerText = `${nft(Date.parse(ts))}`
  newfp.scrollIntoView()
})

const genpath = (x, t, c) => {
  const peakheight = Math.sin(t * 6.7 * 1.25) / 2
  if (peakheight < -.1) {
    return
  }
  const fnc = document.getElementById("forum-notification-canvas")
  fnc.width = window.innerWidth
  fnc.height = window.innerHeight

  const xx = [0, 2 * x / 3, 5 * x / 6, x, x + 5 * (window.innerWidth - x) / 6, x + (window.innerWidth - x) / 3, window.innerWidth]
  const yy = [window.innerHeight + 40, window.innerHeight, window.innerHeight - window.innerHeight / 12 * peakheight, window.innerHeight - window.innerHeight / 3 * peakheight, window.innerHeight - window.innerHeight / 12 * peakheight, window.innerHeight, window.innerHeight + 40]

  const ctx = fnc.getContext("2d")
  ctx.clearRect(0, 0, window.innerWidth, window.innerHeight)
  ctx.fillStyle = c
  ctx.beginPath()
  ctx.moveTo(Math.floor(xx[0]), Math.floor(yy[0]))
  ctx.bezierCurveTo(xx[1], yy[1], xx[2], yy[2], xx[3], yy[3])
  ctx.bezierCurveTo(xx[4], yy[4], xx[5], yy[5], xx[6], yy[6])
  ctx.fill()
  setTimeout(() => { genpath(x, t + .016, c) }, 16)
}
