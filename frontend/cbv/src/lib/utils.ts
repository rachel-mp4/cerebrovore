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
      return formatter.format(then).toLocaleLowerCase()
    } else if (now - then < 1000 * 60 * 60 * 24 * 6) {
      const formatter = new Intl.DateTimeFormat("en-us", { weekday: "long", dayPeriod: "long" })
      return formatter.format(then).toLocaleLowerCase()
    } else if (now - then < 1000 * 60 * 60 * 24 * 333) {
      const formatter1 = new Intl.DateTimeFormat("en-us", { weekday: "long" })
      const formatter2 = new Intl.DateTimeFormat("en-us", { month: "long" })
      const formatter3 = new Intl.DateTimeFormat("en-us", { dayPeriod: "long" })
      return `a ${formatter1.format(then)} in ${formatter2.format(then)} ${formatter3.format(then)}`.toLocaleLowerCase()
    } else {
      const formatter1 = new Intl.DateTimeFormat("en-us", { weekday: "long" })
      const formatter2 = new Intl.DateTimeFormat("en-us", { month: "long" })
      const formatter3 = new Intl.DateTimeFormat("en-us", { year: "numeric", dayPeriod: "long" })
      return `a ${formatter1.format(then)} in ${formatter2.format(then)} ${formatter3.format(then)}`.toLocaleLowerCase()
    }
  } catch {
    return `sometime who cares`
  }
}

export function dumbAbsoluteTimestamp(then: number): string {
  return (new Date(then)).toString()
}
