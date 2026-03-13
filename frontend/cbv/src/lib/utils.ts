// this is some nonsense that ai told me to do because i ran into an issue
// where random control characters were getting adding to certain text of
// a user who also used a RTL language, likely we don't care but i'm
// leaving this around for convenience
export function sanitizeUnicodeControls(input: string) {
  return input
    .normalize('NFKC') // Unicode normalization
    .replace(/[\u0000-\u001F\u007F-\u009F]/g, '') // Control characters
    .replace(/[\u200B-\u200F\u202A-\u202E\u2060-\u206F]/g, '') // Invisible/directional
    .replace(/[\uFEFF]/g, '') // Byte order mark
    .trim();
}

export function smartAbsoluteTimestamp(then: number): string {
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

export function nSecondsToHMS(n: number): string {
  const nHour = Math.floor(n / 3600)
  const nMin = Math.floor((n % 3600) / 60)
  const nSec = Math.floor(n % 60)
  const fhour = (nHour !== 0) ? String(nHour) : ""
  const fmin = (fhour !== "") ? String(nMin).padStart(2, '0') : (nMin !== 0) ? String(nMin) : ""
  const fsec = (fmin !== "") ? String(nSec).padStart(2, '0') : String(nSec)
  let res = ""
  if (fhour !== "") res += fhour + ":"
  if (fmin !== "") res += fmin + ":"
  res += fsec
  return res
}

export function nSecondsOutOfMax(n: number, maxN: string): string {
  if (n < 0) {
    n = 0
  }
  const nHour = Math.floor(n / 3600)
  const nMin = Math.floor((n % 3600) / 60)
  const nSec = Math.floor(n % 60)
  if (maxN.length > 6) {
    return `${nHour}:${String(nMin).padStart(2, '0')}:${String(nSec).padStart(2, '0')}`
  } else if (maxN.length > 3) {
    return `${String(nMin)}:${String(nSec).padStart(2, '0')}`
  }
  return String(nSec)
}

export function dumbAbsoluteTimestamp(then: number): string {
  return (new Date(then)).toString()
}

export function b36encodenumber(n: number): string {
  return n.toString(36)
}
