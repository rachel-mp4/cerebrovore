// start of string, /p/, capture 1+ base36 digits, end of string
const linkRegExp = RegExp(/^\/p\/([0-9a-zA-Z]+)$/)
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

