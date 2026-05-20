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

export function render(ll: line[]): string {
  console.log(ll)
  var res = ""
  var i = false
  var b = false
  var c = false
  const div = document.createElement("div")

  for (let j = 0; j < ll.length; j++) {
    if (j !== 0) {
      res += "\n"
    }
    const l = ll[j]
    if (l.quote !== undefined) {
      res += '<span class="'
      switch (l.quote) {
        case quote.left: {
          res += 'left'
          break
        }
        case quote.up: {
          res += 'up'
          break
        }
        case quote.right: {
          res += 'right'
          break
        }
        case quote.down: {
          res += 'down'
          break
        }
      }
      res += ' quote">'
    }
    for (let k = 0; k < l.tokens.length; k++) {
      const t = l.tokens[k]
      switch (t.state) {
        case state.hashtag: {
          res += `<a href="/p/${t.text.slice(1)}">${ibcstart(i, b, c)}${t.text}${ibcend(i, b, c)}</a>`
          break
        }
        case state.mention: {
          res += `<a href="/profile/${t.text.slice(1)}">${ibcstart(i, b, c)}${t.text}${ibcend(i, b, c)}</a>`
          break
        }
        case state.text: {
          div.textContent = t.text
          res += `${ibcstart(i, b, c)}${div.innerHTML}${ibcend(i, b, c)}`
          break
        }
        case state.italic: {
          res += `${ibcstart(true, b, c)}*${ibcend(true, b, c)}`
          i = !i
          break
        }
        case state.bold: {
          res += `${ibcstart(i, true, c)}**${ibcend(i, true, c)}`
          b = !b
          break
        }
        case state.code: {
          res += `${ibcstart(i, b, true)}\`${ibcend(i, b, true)}`
          c = !c
          break
        }
      }
    }
    if (l.quote !== undefined) {
      res += "</span>"
    }
  }
  return res
}

function ibcstart(italic: boolean, bold: boolean, code: boolean): string {
  return `${code ? "<code>" : ""}${bold ? "<b>" : ""}${italic ? "<em>" : ""}`
}

function ibcend(italic: boolean, bold: boolean, code: boolean): string {
  return `${italic ? "</em>" : ""}${bold ? "</b>" : ""}${code ? "</code>" : ""}`
}

type statemachine = {
  lines: line[]
  line: line
  tokenstart: number
  state: undefined | state
  word: string
  skip: boolean
  first: boolean
}

function eat(sm: statemachine, char: string, nchar: string, idx: number): statemachine {
  if (sm.first) {
    sm.first = false
    sm.line.quote = checkquote(char)
  }
  if (sm.skip) {
    sm.skip = false
    return sm
  }
  switch (char) {
    case "*":
      sm = flush(sm)
      sm.state = (nchar === "*") ? state.bold : state.italic
      return flush(sm)
    case "`":
      sm = flush(sm)
      sm.state = state.code
      return flush(sm)
    case "\n":
      sm = flush(sm)
      sm.lines.push(sm.line)
      sm.line = { tokens: [], start: idx + 1 }
      sm.first = true
      sm.tokenstart = 0
      return sm
  }
  switch (sm.state) {
    case state.hashtag:
    case state.mention:
      if (isAlphanumeric(char)) {
        sm.word += char
        return sm
      } else {
        sm = flush(sm)
      }
  }
  switch (char) {
    case "#": {
      if (isAlphanumericNonzero(nchar)) {
        sm = flush(sm)
        sm.state = state.hashtag
        break
      }
    }
    case "@": {
      if (isAlphanumeric(nchar)) {
        sm = flush(sm)
        sm.state = state.mention
        break
      }
    }
    default: {
      sm.state = state.text
    }
  }
  sm.word += char
  return sm
}

function flush(sm: statemachine): statemachine {
  if (sm.state !== undefined) {
    sm.line.tokens.push({ state: sm.state, start: sm.tokenstart, text: sm.word })
    sm.tokenstart += sm.word.length
    switch (sm.state) {
      case state.bold:
        sm.tokenstart += 2
        sm.skip = true
        break
      case state.hashtag:
      case state.mention:
      case state.italic:
      case state.code:
        sm.tokenstart += 1
    }
    sm.word = ""
    sm.state = undefined
  }
  return sm
}

function checkquote(char: string): quote | undefined {
  switch (char) {
    case "^":
      return quote.up
    case ">":
      return quote.right
    case "<":
      return quote.left
    case "v":
    case "V":
      return quote.down
    default:
      return undefined
  }
}


export function parse(s: string): line[] {
  var sm: statemachine = {
    lines: [],
    line: { tokens: [], start: 0 },
    tokenstart: 0,
    state: undefined,
    word: "",
    skip: false,
    first: true
  }
  for (let i = 0; i < s.length; i++) {
    const char = s.charAt(i)
    const nchar = s.charAt(i + 1)
    sm = eat(sm, char, nchar, i)
  }
  sm = flush(sm)
  sm.lines.push(sm.line)
  return sm.lines
}

// possible cases:
// s contains 0, 1, or 2+ newlines
// case 2+: first line affects tokens before, last line affects tokens after, and we just parse everything in between else
//
export function insert(s: string, into: line[], at: number): line[] {
  const lineNum = find(into, at)
  const tokenNum = find(into[lineNum].tokens, at - into[lineNum].start)
  const newlines = parse(s)
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
