import diff from 'fast-diff'

type line = {
  tokens: token[],
  start: number,
  quote?: quote
}

enum quote {
  up,
  down,
  left,
  right
}

type token = {
  state: state,
  start: number,
  text: string,
}

enum state {
  text,
  hashtag,
  mention,
  code,
  bold,
  italic
}

type nodeline = {
  start: number,
  tokens: nodetoken[],
  quote?: quote
}

type nodetoken = {
  start: number,
  state: nodestate,
  text: string, // avoid using me for determining token length BUT i probably won't avoid this
  styles: styles // so try and always make this correct!
  sentinelstyles?: styles
}

enum nodestate {
  text,
  hashtag,
  mention
}

type diffline = {
  tokens: difftoken[],
  quote?: quote
}

type difftoken = {
  state: nodestate,
  text: string
  diffs: [-1 | 0 | 1, string][],
  styles: styles,
}

type styles = {
  italic: boolean
  bold: boolean
  code: boolean
}

export function render(ll: line[]): string {
  var res = ""
  var i = false
  var b = false
  var c = false

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

export function renderprocessed(nll: nodeline[]): string {
  var res = ""
  const div = document.createElement("div")
  for (let i = 0; i < nll.length; i++) {
    if (i !== 0) {
      res += "\n"
    }
    const nl = nll[i]
    if (nl.quote !== undefined) {
      res += '<span class="'
      switch (nl.quote) {
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
    for (let j = 0; j < nl.tokens.length; j++) {
      const t = nl.tokens[j]
      switch (t.state) {
        case nodestate.text: {
          div.textContent = t.text
          res += `${sstart(t.styles)}${div.innerHTML}${send(t.styles)}`
          break
        }
        case nodestate.hashtag: {
          res += `<a href="/p/${t.text.slice(1)}">${sstart(t.styles)}${t.text}${send(t.styles)}</a>`
          break
        }
        case nodestate.mention: {
          res += `<a href="/profile/${t.text.slice(1)}">${sstart(t.styles)}${t.text}${send(t.styles)}</a>`
          break
        }
      }
    }
    if (nl.quote !== undefined) {
      res += "</span>"
    }
  }

  return res
}

function rendercombined(dll: diffline[]): string {
  var res = ""
  for (let i = 0; i < dll.length; i++) {
    if (i !== 0) {
      res += "\n"
    }
    const nl = dll[i]
    if (nl.quote !== undefined) {
      res += '<span class="'
      switch (nl.quote) {
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
    for (let j = 0; j < nl.tokens.length; j++) {
      const t = nl.tokens[j]
      switch (t.state) {
        case nodestate.text: {
          res += `${sstart(t.styles)}${renderdiffs(t.diffs)}${send(t.styles)}`
          break
        }
        case nodestate.hashtag: {
          res += `<a href="/p/${t.text.slice(1)}">${sstart(t.styles)}${renderdiffs(t.diffs)}${send(t.styles)}</a>`
          break
        }
        case nodestate.mention: {
          res += `<a href="/profile/${t.text.slice(1)}">${sstart(t.styles)}${renderdiffs(t.diffs)}${send(t.styles)}</a>`
          break
        }
      }
    }
    if (nl.quote !== undefined) {
      res += "</span>"
    }
  }
  return res
}

const div = document.createElement("div")

function renderdiffs(diffs: diff.Diff[]): string {
  let res = ""
  for (let i = 0; i < diffs.length; i++) {
    const diff = diffs[i]
    div.textContent = diff[1]
    switch (diff[0]) {
      case -1: {
        res += `<span class="removed">${div.innerHTML.replaceAll("\n", "")}</span>`
        break
      }
      case 0: {
        res += div.innerHTML.replaceAll("\n", "")
        break
      }
      case 1: {
        res += `<span class="appended">${div.innerHTML.replaceAll("\n", "")}</span>`
        break
      }
    }
  }
  return res
}

export function diffzizz(local: string, echo: string): string {
  if (local === "") {
    if (echo === "") {
      return ""
    }
    div.textContent = echo
    return `<span class="removed">${div.innerHTML}</span>`
  }
  if (echo === "") {
    return `<span class="appended">${renderprocessed(preprocess(parse(local)))}</span>`
  }
  if (local === echo) {
    return renderprocessed(preprocess(parse(local)))
  }
  const ppp = preprocess(parse(local))
  const diffs = diff(echo, local)
  console.log(diffs)
  const s = rendercombined(combine(ppp, diffs))
  console.log(s)
  return s
}

function combine(ppp: nodeline[], diffs: diff.Diff[]): diffline[] {
  const res: diffline[] = []
  // idea: 
  // want to go over every token in ppp and create a token for diffs
  // inside it, we essentially want to color each letter according to the
  // relevant diffs, by putting a subdiff into that difftoken. the way
  // we do this is we track which diff we're on right now, try to add as much
  // of it into the token's subdiff array as possible (if it's a subtract, 
  // all of it, if it's an add or equality, we add up to the remaining length
  // of the token's string) then if it was add or equality, we subtract from
  // remaining length the length of the subdiff, and if we added the whole
  // diff, advance to next one, otherwise we need to store for the next token
  // the amount we are partially into that diff so that we can continue coloring
  // where we left off

  // in addition, we need to check for a delete at every fencepost. here, we 
  // check the styles of both sides of the fencepost. if they share styles,
  // we insert into the end of the first one if it's a text node, otherwise the 
  // start of the second one if it's a text node, otherwise it should make a new 
  // text node between them which has the same styles as both of them, and just
  // contains the deleted text

  // if they don't share styles, then the styles of deleted text between them
  // should be the set of styles they do share. now we look at the two fenceposts
  // if the first one is a text node and it has the set of styles they share, then
  // we add the delete to the end of that one, otherwise if the second one is a
  // text node and it shares these styles, then we add it to the beginning of that
  // one, otherwise we create a text node between them which has those shared styles
  // and just contains the deleted text

  // if we create a text node between them, it should go on the line of the second
  // fencepost (which may be the same as the first fencepost, but it may not be)

  let dcursor = 0
  let diff: diff.Diff = diffs[dcursor]
  let add = false
  let adddiff: diff.Diff
  let addstyles: styles = { italic: false, bold: false, code: false }
  // first fencepost: a special case
  if (diff[0] === -1) { // first fencepost has a delete
    add = true
    adddiff = diff
    dcursor += 1
    diff = diffs[dcursor]
  }

  let dinto = 0
  let didx = 0

  for (let i = 0; i < ppp.length; i++) {
    const nl = ppp[i]
    console.log(nl)
    const dl: diffline = { quote: nl.quote, tokens: [] }

    for (let j = 0; j < nl.tokens.length; j++) {
      const nt = nl.tokens[j]
      const dt: difftoken = {
        state: nt.state,
        text: nt.text,
        diffs: [],
        styles: { ...nt.styles }
      }
      if (add) {
        add = false
        if (sequal(addstyles, nt.styles) && nt.state === nodestate.text) {
          dt.diffs.push(adddiff!)
        } else {
          dl.tokens.push({ state: nodestate.text, text: "", diffs: [adddiff!], styles: { ...addstyles } })
        }
      }
      let d1sdi = diff[1].slice(dinto)
      while (nl.start + nt.start + nt.text.length > didx + (diff[0] === -1 ? 0 : d1sdi.length)) {
        switch (diff[0]) {
          case -1: {
            dt.diffs.push(diff)
            break
          }
          case 0:
          case 1: {
            dt.diffs.push([diff[0], d1sdi])
            didx += d1sdi.length
            break
          }
        }
        dinto = 0
        dcursor += 1
        diff = diffs[dcursor]
        if (diff === undefined) {
          dl.tokens.push(dt)
          res.push(dl)
          return res
        }
        d1sdi = diff[1]
      }
      // now we know that EITHER the diff is going to cross over a word
      // OR the diff end and the word end perfectly match each other. 
      if (nl.start + nt.start + nt.text.length === didx + d1sdi.length) {
        dt.diffs.push([diff[0], d1sdi]) // case perfect match
        didx += d1sdi.length
        dinto = 0
        dcursor += 1
        diff = diffs[dcursor]
        if (diff === undefined) {
          dl.tokens.push(dt)
          res.push(dl)
          return res
        }
        d1sdi = diff[1]

        // subtract fence checks!
        if (diff[0] === -1) {
          if (nt.sentinelstyles !== undefined) { // we are the last node
            if (sequal(nt.sentinelstyles, nt.styles)) { // and we didn't just change the styles
              if (dt.state === nodestate.text) {
                dt.diffs.push(diff)
                dl.tokens.push(dt)
              } else {
                dl.tokens.push(dt)
                const tk: difftoken = { state: nodestate.text, text: "", diffs: [diff], styles: { ...nt.styles } }
                dl.tokens.push(tk)
              }
            } else { // and we DID just change the styles
              dl.tokens.push(dt)
              const tk: difftoken = { state: nodestate.text, text: "", diffs: [diff], styles: { ...nt.sentinelstyles } }
              dl.tokens.push(tk)
            }
            res.push(dl)
            // we know all remaining lines are empty, but they may still exist
            while (++i < ppp.length) {
              // prefix operator increments then returns, which is what we want to
              // quickly loop through all remaining lines
              const nl = ppp[i]
              const dl: diffline = { quote: nl.quote, tokens: [] }
              res.push(dl)
            }
            return res
          }

          let nextnt = nl.tokens[j + 1]
          let idx = 1
          while (nextnt === undefined) {
            if (i + idx >= ppp.length) {
              console.error("we failed to add a sentinel")
              // since our current token has undefined sentinelstyles 
              // (otherwise we would have early returned) and we couldn't find
              // a token in any remaining line
              break
            }
            nextnt = ppp[i + idx].tokens[0]
            idx++
          }
          if (sequal(nt.styles, nextnt.styles)) {
            if (nt.state === nodestate.text) {
              dt.diffs.push(diff)
            } else {
              add = true
              adddiff = diff
              addstyles = { ...nt.styles }
            }
          } else {
            // styles changed between this token and next token, so the styles
            // for the subtract diff that we wanna add should be the ones in common
            // which is sand (style and) function
            addstyles = sand(nt.styles, nextnt.styles)
            if (nt.state === nodestate.text && sequal(addstyles, nt.styles)) {
              dt.diffs.push(diff)
            } else {
              add = true
              adddiff = diff
            }
          }
          dcursor += 1
          diff = diffs[dcursor]
          if (diff === undefined) {
            // this case seems bad, since we know we have a next token, but there
            // is no next diff (if there's a token, then there should be a corresponding
            // add or equal diff)
            // additionally, suppose we set add = true, this early flush means
            // that we dropped this subtract diff
            console.error("bad state detected, here's what i've got so far")
            dl.tokens.push(dt)
            res.push(dl)
            return res
          }
          d1sdi = diff[1]
        }

      } else {
        //case diff crossover token boundary
        const word = d1sdi.slice(0, nl.start + nt.start + nt.text.length - didx)
        dt.diffs.push([diff[0], word])
        dinto += word.length
        didx += word.length
      }
      dl.tokens.push(dt)
    }
    res.push(dl)
  }

  return res
}

function sequal(s: styles, t: styles): boolean {
  if (s.italic !== t.italic) {
    return false
  }
  if (s.bold !== t.bold) {
    return false
  }
  if (s.code !== t.code) {
    return false
  }
  return true
}

function sand(s: styles, t: styles): styles {
  return { italic: s.italic && t.italic, bold: s.bold && t.bold, code: s.code && t.code }
}


export function preprocess(ll: line[]): nodeline[] {
  const res: nodeline[] = []
  let lasti: undefined | number
  var s = { italic: false, bold: false, code: false }
  for (let i = 0; i < ll.length; i++) {
    const l = ll[i]
    const nl: nodeline = { start: l.start, tokens: [], quote: l.quote }
    var skip = 0
    for (let j = 0; j < l.tokens.length; j++) {
      if (skip > 0) {
        skip = Math.max(skip - 1, 0)
        continue
      }
      const t = l.tokens[j]
      const tt = l.tokens[j + 1]
      const ttt = l.tokens[j + 2]
      switch (t.state) {
        case state.mention:
        case state.hashtag: {
          nl.tokens.push({ state: t.state === state.mention ? nodestate.mention : nodestate.hashtag, start: t.start, text: t.text, styles: { ...s } })
          break
        }

        case state.text: {
          switch (tt?.state) {
            case state.text: {
              console.error("there shouldn't be two text tokens in a row, so i messed up somewhere", ll)
              return res
            }

            case state.italic: {
              if (s.italic) {
                nl.tokens.push({ state: nodestate.text, start: t.start, text: t.text + tt.text, styles: { ...s } })
                s.italic = false
                skip = 1
                continue
              }
              break
            }

            case state.bold: {
              if (s.bold) {
                nl.tokens.push({ state: nodestate.text, start: t.start, text: t.text + tt.text, styles: { ...s } })
                s.bold = false
                skip = 1
                continue
              }
              break
            }

            case state.code: {
              if (s.code) {
                nl.tokens.push({ state: nodestate.text, start: t.start, text: t.text + tt.text, styles: { ...s } })
                s.code = false
                skip = 1
                continue
              }
              break
            }
          }
          // if we have a text token followed by a non closing thing, it's just the text token
          nl.tokens.push({ state: nodestate.text, start: t.start, text: t.text, styles: { ...s } })
          break
        }

        case state.italic: {
          if (s.italic) {
            nl.tokens.push({ state: nodestate.text, start: t.start, text: t.text, styles: { ...s } })
            s.italic = false
            continue
          }

          //now we are in a starting token
          s.italic = true
          switch (tt?.state) {
            case state.italic: { // tbh this can't happen for italic but this may change who cares
              console.error("i think parser messed up but who cares")
              nl.tokens.push({ state: nodestate.text, start: t.start, text: t.text + tt.text, styles: { ...s } })
              s.italic = false
              skip = 1
              continue
            }

            case state.text: {
              switch (ttt?.state) {
                case state.italic: {
                  nl.tokens.push({ state: nodestate.text, start: t.start, text: t.text + tt.text + ttt.text, styles: { ...s } })
                  s.italic = false
                  skip = 2
                  continue
                }

                case state.text: {
                  console.error("there shouldn't be two text tokens in a row, so i messed up somewhere", ll)
                  return res
                }

                default: {
                  nl.tokens.push({ state: nodestate.text, start: t.start, text: t.text + tt.text, styles: { ...s } })
                  skip = 1
                  continue
                }
              }
            }

            default: {
              nl.tokens.push({ state: nodestate.text, start: t.start, text: t.text, styles: { ...s } })
              continue
            }
          }
        }

        case state.bold: {
          if (s.bold) {
            nl.tokens.push({ state: nodestate.text, start: t.start, text: t.text, styles: { ...s } })
            s.bold = false
            continue
          }

          //now we are in a bold starting token
          s.bold = true
          switch (tt?.state) {
            case state.bold: {
              nl.tokens.push({ state: nodestate.text, start: t.start, text: t.text + tt.text, styles: { ...s } })
              s.bold = false
              skip = 1
              continue
            }

            case state.text: {
              switch (ttt?.state) {
                case state.bold: {
                  nl.tokens.push({ state: nodestate.text, start: t.start, text: t.text + tt.text + ttt.text, styles: { ...s } })
                  s.bold = false
                  skip = 2
                  continue
                }

                case state.text: {
                  console.error("there shouldn't be two text tokens in a row, so i messed up somewhere", ll)
                  return res
                }

                default: {
                  nl.tokens.push({ state: nodestate.text, start: t.start, text: t.text + tt.text, styles: { ...s } })
                  skip = 1
                  continue
                }
              }
            }

            default: {
              nl.tokens.push({ state: nodestate.text, start: t.start, text: t.text, styles: { ...s } })
              continue
            }
          }
        }

        case state.code: {
          if (s.code) {
            nl.tokens.push({ state: nodestate.text, start: t.start, text: t.text, styles: { ...s } })
            s.code = false
            continue
          }

          //now we are in a starting token
          s.code = true
          switch (tt?.state) {
            case state.code: {
              nl.tokens.push({ state: nodestate.text, start: t.start, text: t.text + tt.text, styles: { ...s } })
              s.code = false
              skip = 1
              continue
            }

            case state.text: {
              switch (ttt?.state) {
                case state.code: {
                  nl.tokens.push({ state: nodestate.text, start: t.start, text: t.text + tt.text + ttt.text, styles: { ...s } })
                  s.code = false
                  skip = 2
                  continue
                }

                case state.text: {
                  console.error("there shouldn't be two text tokens in a row, so i messed up somewhere", ll)
                  return res
                }

                default: {
                  nl.tokens.push({ state: nodestate.text, start: t.start, text: t.text + tt.text, styles: { ...s } })
                  skip = 1
                  continue
                }
              }
            }

            default: {
              nl.tokens.push({ state: nodestate.text, start: t.start, text: t.text, styles: { ...s } })
              continue
            }
          }
        }
      }
    }
    res.push(nl)
    // when we push the line, we can see if it had a token in it
    if (nl.tokens.length > 0) {
      lasti = i
    }
  }
  // if at least one line had a token in it, then lasti is defined
  // so we can find the last token in that line and add the styles that remain
  // at the end as our sentinelstyles so that way if the diff ends with like a 
  // subtract that happens after a style close tag, we can render it unstyled as
  // desired
  if (lasti != undefined) {
    const lastline = res[lasti]
    const lt = lastline.tokens[lastline.tokens.length - 1]
    if (lt !== undefined) {
      lt.sentinelstyles = s
    }
  }
  return res
}

function ibcstart(italic: boolean, bold: boolean, code: boolean): string {
  return `${code ? "<code>" : ""}${bold ? "<b>" : ""}${italic ? "<em>" : ""}`
}

function sstart(s: styles): string {
  return `${s.code ? "<code>" : ""}${s.bold ? "<b>" : ""}${s.italic ? "<em>" : ""}`
}

function send(s: styles): string {
  return `${s.italic ? "</em>" : ""}${s.bold ? "</b>" : ""}${s.code ? "</code>" : ""}`
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
      sm.word = (nchar === "*") ? "**" : "*"
      return flush(sm)
    case "`":
      sm = flush(sm)
      sm.state = state.code
      sm.word = "`"
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
    if (sm.state === state.bold) {
      sm.skip = true
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

// function find(within: { start: number }[], start: number): number {
//   if (within.length === 0 || within.length === 1) {
//     return 0
//   }
//   let left = 0
//   let right = within.length
//   let middle = Math.floor((left + right) / 2)
//   while (right - left > 1) {
//     let middlestart = within[middle].start
//     if (middlestart === start) {
//       return middle
//     } else if (middlestart < start) {
//       left = middle
//     } else {
//       right = middle
//     }
//     middle = Math.floor((left + right) / 2)
//   }
//   return middle
// }

// const test1 = [{ start: 0 }, { start: 4 }, { start: 9 }, { start: 15 }, { start: 21 }]
// const test2 = [{ start: 0 }, { start: 4 }, { start: 9 }, { start: 15 }, { start: 21 }, {start: 34}]
