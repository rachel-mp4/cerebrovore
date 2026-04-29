// start of string, /p/, capture 1+ base36 digits, end of string
const linkRegExp = RegExp(/^\/p\/([0-9a-zA-Z]+)$/)
const scrollIfOnPage = (e) => {
  const link = e.target.closest("a[href]")
  if (!link) {
    return
  }
  const href = link.pathname
  const matches = href.match(linkRegExp)
  if (!matches) {
    return
  }
  // string.match uses index 1 for capture group
  const id = matches[1].toLowerCase()
  const target = document.getElementById(id)
  if (!target) {
    return
  }
  e.preventDefault()
  target.scrollIntoView()

}
if (window.location.pathname.startsWith("/t/")) {
  document.addEventListener("click", (e) => {
    scrollIfOnPage(e)
  })
} else if (window.location.pathname.startsWith("/ft/")) {
  document.addEventListener("click", (e) => {
    const time = e.target.closest(".time")
    if (time) {
      const el = time.closest(".forum-post")
      if (el) {
        el.classList.toggle("collapsed")
      }
      return
    }
    const reply = e.target.closest(".reply")
    if (reply) {
      e.preventDefault()
      const tc = reply.textContent.trim()
      const rf = document.getElementById("response-form")
      if (rf) {
        rf.value += `${tc}\n`
        const sel = window.getSelection()
        const el = reply.closest(".forum-post")
        console.log(el)
        if (!(el.contains(sel.anchorNode) && el.contains(sel.focusNode))) {
          rf.focus()
          return
        }
        const seltext = sel.toString()
        if (seltext !== "") {
          const sels = seltext.split("\n")
          sels.forEach((minisel) => {
            rf.value += `>${minisel}\n`
          })
        }
        rf.focus()
      }
      return
    }
    scrollIfOnPage(e)
  })
} else {
  document.addEventListener("click", (e) => {
    const reply = e.target.closest(".reply")
    if (!reply) {
      return
    }
    const newURL = `/p/${reply.innerHTML.trim().slice(1)}`
    if (newURL.match(linkRegExp)) {
      window.location.assign(newURL)
    }
  })
}
