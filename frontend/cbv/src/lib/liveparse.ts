export type line = {
  tokens: token[],
  start: number,
  quote?: quote
}

export enum quote {
  up,
  down,
  left,
  right
}

export type token = {
  state: state,
  start: number,
  text: string,
}

export enum state {
  text,
  hashtag,
  mention,
  code,
  bold,
  italic
}

export function parse(s: string): line[] {
  const res: line[] = []
  let l: line = { tokens: [], start: 0 }
  let tokenstart = 0
  let currentstate: undefined | state = undefined
  let word = ""
  let skip = false
  let first = true
  for (let i = 0; i < s.length; i++) {
    const char = s.charAt(i)
    if (first) {
      first = false
      switch (char) {
        case "^": {
          l.quote = quote.up
          break
        }
        case ">": {
          l.quote = quote.right
          break
        }
        case "v": {
          l.quote = quote.down
          break
        }
        case "V": {
          l.quote = quote.down
          break
        }
        case "<": {
          l.quote = quote.left
          break
        }
      }
    }
    if (skip) {
      skip = false
      continue
    }
    switch (char) {
      case "*": {
        if (currentstate !== undefined) {
          l.tokens.push({ state: currentstate, start: tokenstart, text: word })
          tokenstart += word.length
          currentstate = undefined
          word = ""
        }
        if (s.charAt(i + 1) === "*") {
          l.tokens.push({ state: state.bold, start: tokenstart, text: "**" })
          tokenstart += 2
          skip = true
        } else {
          l.tokens.push({ state: state.italic, start: tokenstart, text: "*" })
          tokenstart += 1
        }
        continue;
      }
      case "`": {
        if (currentstate !== undefined) {
          l.tokens.push({ state: currentstate, start: tokenstart, text: word })
          tokenstart += word.length
          currentstate = undefined
          word = ""
        }
        l.tokens.push({ state: state.code, start: tokenstart, text: "`" })
        tokenstart += 1
        continue;
      }
      case "\n": {
        if (currentstate !== undefined) {
          l.tokens.push({ state: currentstate, start: tokenstart, text: word })
          currentstate = undefined
          word = ""
        }
        res.push(l)
        l = { tokens: [], start: i + 1 }
        first = true
        tokenstart = 0
        continue
      }
    }
    switch (currentstate) {
      case undefined: {
        tokenstart = i
        switch (char) {
          case "#": {
            if (isAlphanumericNonzero(s.charAt(i + 1))) {
              currentstate = state.hashtag
            } else {
              currentstate = state.text
              word = "#"
            }
            continue
          }
          case "@": {
            if (isAlphanumeric(s.charAt(i + 1))) {
              currentstate = state.mention
            } else {
              currentstate = state.text
              word = "@"
            }
            continue
          }
          default: {
            word = char
            continue
          }
        }
      }
      case state.hashtag: {
        if (isAlphanumeric(char)) {
          word += char
        } else {
          l.tokens.push({ state: currentstate, start: tokenstart, text: word })
          tokenstart = i
          if (char === )
        }

      }

    }



  }
  return res
}

// possible cases:
// s contains 0, 1, or 2+ newlines
// case 2+: first line affects tokens before, last line affects tokens after, and we just parse everything in between else
//
export function insert(s: string, into: line[], at: number): line[] {
  const lineNum = find(into, at)
  const tokenNum = find(into[lineNum].tokens, at - into[lineNum].start)
  return into
}

function isAlphanumeric(char: string): boolean {
  const code = char.charCodeAt(0);
  return (code > 47 && code < 58) || // numeric (0-9)
    (code > 64 && code < 91) || // upper alpha (A-Z)
    (code > 96 && code < 123);  // lower alpha (a-z)
}

function isAlphanumericNonzero(char: string): boolean {
  const code = char.charCodeAt(0);
  return (code > 48 && code < 58) || // numeric (1-9)
    (code > 64 && code < 91) || // upper alpha (A-Z)
    (code > 96 && code < 123);  // lower alpha (a-z)
}

function find(within: { start: number }[], idx: number): number {
  if (within.length === 0 || within.length === 1) {
    return 0
  }
  let left = 0
  let right = within.length
  let middle = Math.floor((left + right) / 2)
  while (left <= right - 1) {
    let middleidx = within[middle].start
    if (middleidx === idx) {
      return middle
    } else if (middleidx < idx) {
      left = middle
    } else {
      right = middle
    }
    middle = Math.floor((left + right) / 2)
  }
  return middle
}
