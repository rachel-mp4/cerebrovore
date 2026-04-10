const eub = document.getElementById("main-content")
const times = eub.querySelectorAll(".time")
const tft = (then) => {
  const now = Date.now()
  try {
    if (then > now) {
      return "in the future"
    } else if (now - then < 1000 * 60 * 60 * 18) {
      const formatter = new Intl.DateTimeFormat("en-us", { hour: "numeric", minute: "numeric" })
      return `at ${formatter.format(then).toLocaleLowerCase()}`
    } else if (now - then < 1000 * 60 * 60 * 24 * 6) {
      const formatter = new Intl.DateTimeFormat("en-us", { weekday: "long", dayPeriod: "long" })
      return `on ${formatter.format(then).toLocaleLowerCase()}`
    } else if (now - then < 1000 * 60 * 60 * 24 * 333) {
      const formatter1 = new Intl.DateTimeFormat("en-us", { weekday: "long" })
      const formatter2 = new Intl.DateTimeFormat("en-us", { month: "long" })
      const formatter3 = new Intl.DateTimeFormat("en-us", { dayPeriod: "long" })
      return `on a ${formatter1.format(then)} in ${formatter2.format(then)} ${formatter3.format(then)}`.toLocaleLowerCase()
    } else {
      const formatter1 = new Intl.DateTimeFormat("en-us", { weekday: "long" })
      const formatter2 = new Intl.DateTimeFormat("en-us", { month: "long" })
      const formatter3 = new Intl.DateTimeFormat("en-us", { year: "numeric", dayPeriod: "long" })
      return `on a ${formatter1.format(then)} in ${formatter2.format(then)} ${formatter3.format(then)}`.toLocaleLowerCase()
    }
  } catch {
    return `sometime who cares`
  }
}
times.forEach((time) => {
  const ts = time.getAttribute("datetime")
  time.innerText = `posted ${tft(Date.parse(ts))}`
})
