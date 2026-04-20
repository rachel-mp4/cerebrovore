// start of string, /p/, capture 1+ base36 digits, end of string
const linkRegExp = RegExp(/^\/p\/([0-9a-zA-Z]+)$/)
if (window.location.pathname.startsWith("/t/")) {
  document.addEventListener("click", (e) => {
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
