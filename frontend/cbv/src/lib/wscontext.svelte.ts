import type * as cbv from "./types"
import { isMessage, isImage, isMedia } from "./types"
import { b36encodenumber } from "./utils"
import * as lrc from '@rachel-mp4/lrcproto/gen/ts/lrc'
import { numToHex } from "./colors"

export class WSContext {
  existingindices: Map<number, boolean> = new Map()
  items: Array<cbv.Item> = $state(new Array())
  log: Array<cbv.LogItem> = $state(new Array())
  topic: string = $state("")
  connected: boolean = $state(false)
  conncount = $state(0)
  ws: WebSocket | null = null
  color: number = $state(Math.floor(Math.random() * 16777216))

  nick: string = "wanderer"
  curMsg: string = $state("")
  myMessage: cbv.Message | undefined
  messageactive: boolean = false
  myImageRef: string | undefined = $state()
  myImageNonce: string | undefined
  myMedia: cbv.Media | undefined
  mediaactive: boolean = false

  shouldTransmit: boolean = $state(true)
  lrceventqueue: Array<lrc.Edit> = []

  constructor(defaultNick: string, defaultColor: number) {
    this.nick = defaultNick
    this.color = defaultColor
  }

  connect(url: string) {
    this.ws?.close()
    connectTo(url, this)
  }

  reconnect = (url: string) => {
    this.ws?.close()
    connectTo(url, this)
    this.items = []
  }

  disconnect = () => {
    this.ws?.close()
    this.ws = null
    this.items = []
  }

  starttransmit = () => {
    if (this.lrceventqueue.length != 0) {
      const evt: lrc.Event = {
        msg: {
          oneofKind: "editbatch",
          editbatch: {
            edits: this.lrceventqueue,
          }
        }
      }
      const byteArray = lrc.Event.toBinary(evt)
      this.ws?.send(byteArray)
      this.lrceventqueue = []
    }
  }

  insertLineBreak = () => {
    if (this.myMessage) {
      this.starttransmit()
      pubMessage(this)
      let body = this.curMsg
      const fd = new FormData()
      fd.append("id", b36encodenumber(this.myMessage.id))
      fd.append("color", numToHex(this.color))
      fd.append("nick", this.nick)
      fd.append("body", body)
      if (this.myMessage.lrcdata?.init?.nonce) {
        fd.append("nonce", b64encodebytearray(this.myMessage.lrcdata.init.nonce))
      }
      const endpoint = window.location.href
      console.log("endpoint: ", endpoint)
      fetch(endpoint, {
        method: "POST",
        body: fd,
      }).then((response) => {
        if (response.ok) {
          console.log(response)
        } else {
          throw new Error(`HTTP ${response.status}`)
        }
      }).catch(console.error)
      this.myMessage = undefined
      this.messageactive = false
      this.curMsg = ""
    } else if (this.messageactive) {
      // i believe this is the case where we just started typing and we haven't recieved the response from the initial
      // message yet. this potentially is not ideal because we may have myMessage defined and not set back to undefined
      // i'm just putting a note here to remind me about this possible race. i don't think it should have any issues...
      this.starttransmit()
      pubMessage(this)
      this.messageactive = false
      this.curMsg = ""
    }
  }

  pubImage = (alt: string) => {
    if (this.myMedia) {
      if (this.myImageRef) {
        const contentAddress = `/blob?cid=${this.myImageRef}`
        pubImage(alt, contentAddress, this)
        let body = this.curMsg
        const fd = new FormData()
        fd.append("id", b36encodenumber(this.myMedia.id))
        fd.append("color", numToHex(this.color))
        fd.append("nick", this.nick)
        fd.append("cid", this.myImageRef)
        fd.append("alt", alt)
        if (this.myMedia.lrcdata?.init?.nonce) {
          fd.append("nonce", b64encodebytearray(this.myMedia.lrcdata.init.nonce))
        }
        const endpoint = window.location.href
        console.log("endpoint: ", endpoint)
        fetch(endpoint, {
          method: "POST",
          body: fd,
        }).then((response) => {
          if (response.ok) {
            console.log(response)
          } else {
            throw new Error(`HTTP ${response.status}`)
          }
        }).catch(console.error)
      } else {
        pubImage(alt, undefined, this)
      }
      this.myMedia = undefined
      this.myImageRef = undefined
      this.mediaactive = false
    } else if (this.mediaactive) {
      if (this.myImageRef) {
        console.error("myImageRef should be undefined in this case") // why?
        this.myImageRef = undefined                                  // perhaps i assumed this because it should take a
      }                                                              // while to upload the image, and the round trip for
      pubImage(alt, undefined, this)                                 // both lrc and image upload should be the same?
      this.mediaactive = false
    }
  }

  cancelImage = () => {
    if (this.mediaactive) {
      pubImage(undefined, undefined, this)
      this.myMedia = undefined
      this.myImageRef = undefined
      this.mediaactive = false
    }
  }

  initImage = (blob: File) => {
    if (!this.myMedia) {
      initImage(this)
      this.mediaactive = true
      const uuid = crypto.randomUUID()
      const formData = new FormData()
      formData.append("file", blob)
      formData.append("uuid", uuid)
      this.myImageNonce = uuid
      fetch(`/blob`, {
        method: "POST",
        body: formData
      }).then((response) => {
        if (response.ok) {
          response.json().then((data) => {
            if (this.myImageNonce === data.uuid) {
              this.myImageRef = data.cid
              this.myImageNonce = undefined
              console.log("here's myImageRef", this.myImageRef)
            } else {
              console.error("nonce mismatch!!!")
            }
          })
        } else {
          throw new Error(`HTTP ${response.status}`)
        }
      }).catch((err) => { console.log(err) })
    }
  }

  insert = (idx: number, s: string) => {
    if (!this.messageactive) {
      initMessage(this)
      this.messageactive = true
    }
    insertMessage(idx, s, this)
    this.curMsg = insertSIntoAStringAtIdx(s, this.curMsg, idx)
  }

  delete = (idx: number, idx2: number) => {
    if (!this.messageactive) {
      return
    }
    deleteMessage(idx, idx2, this)
    this.curMsg = deleteFromAStringBetweenIdxs(this.curMsg, idx, idx2)
  }
  mute = (id: number) => {
    muteMessage(id, this)
  }

  unmute = (id: number) => {
    unmuteMessage(id, this)
  }

  setNick = (nick: string) => {
    setNick(nick, this)
  }
  setColor = (color: number) => {
    setColor(color, this)
  }

  setTopic = (topic: string) => {
    console.log("new topic:", topic)
    this.topic = topic
  }

  setConncount = (cc: number) => {
    this.conncount = cc
  }

  pushItem = (item: cbv.Item) => {
    if (this.existingindices.get(item.id)) {
      console.log("you tried to push an item who exists!")
      return
    }
    if (item.lrcdata.mine) {
      if (isMessage(item)) {
        this.myMessage = item
      } else if (isMedia(item)) {
        this.myMedia = item
      }
    }
    this.items.push(item)
    this.existingindices.set(item.id, true)
  }

  initMessage = (id: number, init: cbv.LrcInit, mine: boolean) => {
    if (this.existingindices.get(id)) {
      this.items = this.items.map((item: cbv.Item) => {
        return item.id === id && isMessage(item)
          ? { ...item, type: "message", lrcdata: { ...item.lrcdata, init: init } }
          : item
      })
    } else {
      console.log("push message init")
      this.pushItem({
        type: 'message',
        id: id,
        lrcdata: {
          body: '',
          mine: mine,
          muted: false,
          init: init,
        },
        replies: []
      })
    }
  }

  initMedia = (id: number, init: cbv.LrcInit, mine: boolean) => {
    if (this.existingindices.get(id)) {
      this.items = this.items.map((item: cbv.Item) => {
        return item.id === id && isImage(item)
          ? { ...item, type: "image", lrcdata: { ...item.lrcdata, init: init } }
          : item
      })
    } else {
      console.log("push media init")
      this.pushItem({
        type: 'image',
        id: id,
        lrcdata: {
          mine: mine,
          muted: false,
          init: init,
        },
        replies: []
      })
    }
  }

  initMute = (id: number) => {
    if (this.existingindices.get(id)) {
      this.items = this.items.map((item: cbv.Item) => {
        return item.id === id
          ? { ...item, lrcdata: { ...item.lrcdata, muted: true } } as typeof item
          : item
      })
    } else {
      console.log("push mute init")
      this.pushItem({
        type: 'enby',
        id: id,
        lrcdata: {
          mine: false,
          muted: true,
        },
        replies: []
      })
    }
  }

  pubMessage = (id: number) => {
    if (this.existingindices.get(id)) {
      this.items = this.items.map((item: cbv.Item) => {
        return item.id === id && isMessage(item)
          ? { ...item, type: "message", lrcdata: { ...item.lrcdata, pub: true } }
          : item
      })
    } else {
      console.log("push message pub")
      this.pushItem({
        type: "message",
        id: id,
        lrcdata: {
          mine: false,
          muted: false,
          body: "",
        },
        replies: []
      })
    }
  }

  pubMedia = (id: number, pub: cbv.LrcMediaPub) => {
    if (this.existingindices.get(id)) {
      this.items = this.items.map((item: cbv.Item) => {
        return item.id === id && isMedia(item)
          ? {
            ...item, type: "image",
            lrcdata: {
              ...item.lrcdata,
              pub: pub
            }
          }
          : item
      })
    } else {
      console.log("push media pub")
      this.pushItem({
        type: "image",
        id: id,
        lrcdata: {
          mine: false,
          muted: false,
          pub: pub,
        },
        replies: []
      })
    }
  }

  insertMessage = (id: number, idx: number, s: string) => {
    if (this.existingindices.get(id)) {
      this.items = this.items.map((item: cbv.Item) => {
        return item.id === id && isMessage(item)
          ? { ...item, type: "message", lrcdata: { ...item.lrcdata, body: insertSIntoAStringAtIdx(s, item.lrcdata.body, idx) } }
          : item
      })
    } else {

      console.log("push message insert")
      this.pushItem({
        type: "message",
        id: id,
        lrcdata: {
          mine: false,
          muted: false,
          body: insertSIntoAStringAtIdx(s, "", idx),
          pub: false
        },
        replies: []
      })
    }
  }

  deleteMessage = (id: number, idx1: number, idx2: number) => {
    if (this.existingindices.get(id)) {
      this.items = this.items.map((item: cbv.Item) => {
        return item.id === id && isMessage(item)
          ? { ...item, type: "message", lrcdata: { ...item.lrcdata, body: deleteFromAStringBetweenIdxs(item.lrcdata.body, idx1, idx2) } }
          : item
      })
    } else {

      console.log("push message delete")
      this.pushItem({
        type: "message",
        id: id,
        lrcdata: {
          mine: false,
          muted: false,
          body: deleteFromAStringBetweenIdxs("", idx1, idx2),
          pub: false
        },
        replies: []
      })
    }
  }

  pushToLog = (id: number, ba: Uint8Array, type: string) => {
    const bstring = Array.from(ba).map(byte => byte.toString(16).padStart(2, "0")).join('')
    const time = Date.now()
    this.log = [...this.log.filter(li => li.time > Date.now() - 3000), { id: id, binary: bstring, time: time, type: type, key: Math.random() }]
    console.log(this.log.length)
  }
}


const b64encodebytearray = (u8: Uint8Array): string => {
  return btoa(String.fromCharCode(...u8))
}

const insertSIntoAStringAtIdx = (s: string, a: string, idx: number) => {
  if (a === undefined) {
    a = ""
  }
  if (idx > a.length) {
    a = a.padEnd(idx)
  }
  return a.slice(0, idx) + s + a.slice(idx)
}

const deleteFromAStringBetweenIdxs = (a: string, idx1: number, idx2: number) => {
  if (a === undefined) {
    a = ""
  }
  if (idx2 > a.length) {
    a = a.padEnd(idx2)
  }
  return a.slice(0, idx1) + a.slice(idx2)
}

export const connectTo = (url: string, ctx: WSContext) => {
  const ws = new WebSocket(url, "lrc.v1");
  ws.binaryType = "arraybuffer";
  ws.onopen = () => {
    console.log("connected")
    ctx.connected = true
    getTopic(ctx)
    setNick(ctx.nick, ctx)
    setColor(ctx.color, ctx)
  };
  const el: HTMLElement | null = document.getElementById("main-content")
  ws.onmessage = (event) => {
    console.log(event)
    if (el && el.scrollTop + el.clientHeight >= el.scrollHeight - 1) {
      const shouldScroll = parseEvent(event, ctx)
      if (shouldScroll !== 0) {
        console.log("scrolling there")
        setTimeout(() => {
          el.scrollTo(0, el.scrollHeight)
        }, 0)
      }
    } else {
      const shouldScroll = parseEvent(event, ctx)
      if (shouldScroll === 2) {
        setTimeout(() => {
          if (el) el.scrollTo(0, el.scrollHeight)
        }, 0)

      }
    }
    // i wonder why i commented this? it looks correct NGL, oh i think i was doing some stuff with ShouldScroll
    // parseEvent used to return a bool
    // if (shouldScroll) {
    //     setTimeout(() => {
    //         window.scrollTo(0, document.body.scrollHeight)
    //     }, 0)
    // }

  };
  ws.onclose = () => {
    console.log("closed")
    if (ws === ctx.ws) {
      ctx.connected = false
    }
  };
  ws.onerror = (event) => {
    console.log("errored:", event)
    console.log("readyState:", ws.readyState)
    if (ws === ctx.ws) {
      ctx.connected = false
      // probably i should retry with exp backoff or something
    }
  }
  ctx.ws = ws
}

export const initMessage = (ctx: WSContext) => {
  const evt: lrc.Event = {
    msg: {
      oneofKind: "init",
      init: {
        nick: ctx.nick,
        color: ctx.color,
      }
    }
  }
  const byteArray = lrc.Event.toBinary(evt)
  ctx.ws?.send(byteArray)
}

export const initImage = (ctx: WSContext) => {
  console.log("send media init!!!")
  const evt: lrc.Event = {
    msg: {
      oneofKind: "mediainit",
      mediainit: {
        nick: ctx.nick,
        color: ctx.color,
      }
    }
  }
  const byteArray = lrc.Event.toBinary(evt)
  ctx.ws?.send(byteArray)
}

export const pubImage = (alt: string | undefined, contentAddress: string | undefined, ctx: WSContext) => {
  const evt: lrc.Event = {
    msg: {
      oneofKind: "mediapub",
      mediapub: {
        alt: alt,
        contentAddress: contentAddress,
      }
    }
  }
  const byteArray = lrc.Event.toBinary(evt)
  ctx.ws?.send(byteArray)
}

export const insertMessage = (idx: number, s: string, ctx: WSContext) => {
  if (ctx.shouldTransmit) {
    const evt: lrc.Event = {
      msg: {
        oneofKind: "insert",
        insert: {
          utf16Index: idx,
          body: s
        }
      }
    }
    const byteArray = lrc.Event.toBinary(evt)
    ctx.ws?.send(byteArray)
  } else {
    const edit: lrc.Edit = {
      edit: {
        oneofKind: "insert",
        insert: {
          utf16Index: idx,
          body: s
        }
      }
    }
    ctx.lrceventqueue.push(edit)
  }
}

export const pubMessage = (ctx: WSContext) => {
  const evt: lrc.Event = {
    msg: {
      oneofKind: "pub",
      pub: {
      }
    }
  }
  const byteArray = lrc.Event.toBinary(evt)
  ctx.ws?.send(byteArray)
}

export const deleteMessage = (idx: number, idx2: number, ctx: WSContext) => {
  if (ctx.shouldTransmit) {

    const evt: lrc.Event = {
      msg: {
        oneofKind: "delete",
        delete: {
          utf16Start: idx,
          utf16End: idx2
        }
      }
    }
    const byteArray = lrc.Event.toBinary(evt)
    ctx.ws?.send(byteArray)
  } else {
    const edit: lrc.Edit = {
      edit: {
        oneofKind: "delete",
        delete: {
          utf16Start: idx,
          utf16End: idx2
        }
      }
    }
    ctx.lrceventqueue.push(edit)
  }
}

export const muteMessage = (id: number, ctx: WSContext) => {
  const evt: lrc.Event = {
    msg: {
      oneofKind: "mute",
      mute: {
        id: id,
      }
    }
  }
  const byteArray = lrc.Event.toBinary(evt)
  ctx.ws?.send(byteArray)
}

export const unmuteMessage = (id: number, ctx: WSContext) => {
  const evt: lrc.Event = {
    msg: {
      oneofKind: "unmute",
      unmute: {
        id: id,
      }
    }
  }
  const byteArray = lrc.Event.toBinary(evt)
  ctx.ws?.send(byteArray)
}

export const getTopic = (ctx: WSContext) => {
  const evt: lrc.Event = {
    msg: {
      oneofKind: "get",
      get: {
        topic: "_"
      }
    }
  }
  const byteArray = lrc.Event.toBinary(evt)
  ctx.ws?.send(byteArray)
}

export const setNick = (nick: string, ctx: WSContext) => {
  ctx.nick = nick
  const evt: lrc.Event = {
    msg: {
      oneofKind: "set",
      set: {
        nick: nick
      }
    }
  }
  const byteArray = lrc.Event.toBinary(evt)
  ctx.ws?.send(byteArray)
}

export const setColor = (color: number, ctx: WSContext) => {
  ctx.color = color
  const evt: lrc.Event = {
    msg: {
      oneofKind: "set",
      set: {
        color: color
      }
    }
  }
  const byteArray = lrc.Event.toBinary(evt)
  ctx.ws?.send(byteArray)
}

function parseEvent(binary: MessageEvent<any>, ctx: WSContext): number {
  const byteArray = new Uint8Array(binary.data);
  const event = lrc.Event.fromBinary(byteArray)
  switch (event.msg.oneofKind) {
    case "ping": {
      return 0;
    }

    case "pong": {
      return 0
    }

    case "init": {
      const id = event.msg.init.id ?? 0
      if (id === 0) return 0
      const color = event.msg.init.color
      const nick = event.msg.init.nick
      const handle = event.msg.init.externalID
      const nonce = event.msg.init.nonce
      const mine = event.msg.init.echoed ?? false
      const init: cbv.LrcInit = {
        ...(color && { color: color }),
        ...(nick && { nick: nick }),
        ...(handle && { handle: handle }),
        ...(nonce && { nonce: nonce }),
      }
      ctx.initMessage(id, init, mine)
      ctx.pushToLog(id, byteArray, "init")
      return mine ? 2 : 1
    }

    case "mediainit": {
      const id = event.msg.mediainit.id ?? 0
      if (id === 0) return 0
      const color = event.msg.mediainit.color
      const nick = event.msg.mediainit.nick
      const handle = event.msg.mediainit.externalID
      const nonce = event.msg.mediainit.nonce
      const mine = event.msg.mediainit.echoed ?? false
      const init: cbv.LrcInit = {
        ...(color && { color: color }),
        ...(nick && { nick: nick }),
        ...(handle && { handle: handle }),
        ...(nonce && { nonce: nonce }),
      }
      ctx.initMedia(id, init, mine)
      ctx.pushToLog(id, byteArray, "init")
      return mine ? 2 : 1
    }

    case "pub": {
      const id = event.msg.pub.id ?? 0
      if (id === 0) return 0
      ctx.pubMessage(id)
      ctx.pushToLog(id, byteArray, "pub")
      return 0
    }

    case "mediapub": {
      const id = event.msg.mediapub.id ?? 0
      if (id === 0) return 0
      const pub: cbv.LrcMediaPub = {
        alt: event.msg.mediapub.alt ?? "",
        contentAddress: event.msg.mediapub.contentAddress
      }
      ctx.pubMedia(id, pub)
      ctx.pushToLog(id, byteArray, "pub")
      return 0
    }

    case "insert": {
      const id = event.msg.insert.id ?? 0
      if (id === 0) return 0
      ctx.pushToLog(id, byteArray, "insert")
      doinsert(id, event.msg.insert, ctx)
      return 1
    }

    case "delete": {
      const id = event.msg.delete.id ?? 0
      if (id === 0) return 0
      ctx.pushToLog(event.msg.delete.id ?? 0, byteArray, "delete")
      dodelete(id, event.msg.delete, ctx)
      return 1
    }


    case "mute": {
      const id = event.msg.mute.id ?? 0
      if (id === 0) return 0
      ctx.initMute(id)
      return 0
    }

    case "unmute": {
      return 0
    }

    case "set": {
      return 0
    }

    case "get": {
      if (event.msg.get.connected !== undefined) {
        ctx.setConncount(event.msg.get.connected)
      }
      if (event.msg.get.topic !== undefined) {
        ctx.setTopic(event.msg.get.topic)
      }
      return 0
    }
    //TODO: better logging system so that way even non hrt messages
    // can have the background effect!
    case "editbatch": {
      const id = event.id ?? 0
      if (id === 0) {
        return 0
      }
      event.msg.editbatch.edits.forEach((edit: lrc.Edit) => {
        switch (edit.edit.oneofKind) {
          case "insert": {
            doinsert(id, edit.edit.insert, ctx)
            return
          }
          case "delete": {
            dodelete(id, edit.edit.delete, ctx)
            return
          }
        }
      })
      return 1

    }
    case "replybatch": {
      event.msg.replybatch.replies.forEach((reply: lrc.Reply) => {
        switch (reply.reply.oneofKind) {
          case "detachreply": {
            dodetachreply(reply.reply.detachreply, ctx)
            return
          }
          case "attachreply": {
            doattachreply(reply.reply.attachreply, ctx)
            return
          }
        }
      })
      return 1
    }

  }
  return 0
}

function dodetachreply(detach: lrc.DetachReply, ctx: WSContext) {
  const from = detach.from
  if (from == null) {
    return
  }
  const to = detach.to
  if (!ctx.existingindices.get(to)) {
    console.log("manually detach me! unimplemented")
    return
  }
  ctx.items = ctx.items.map((item) => {
    return (item.id !== to) ? item : { ...item, replies: item.replies.filter((id) => id !== from) }
  })
}

function doattachreply(attach: lrc.AttachReply, ctx: WSContext) {
  const from = attach.from
  if (from == null) {
    return
  }
  const to = attach.to
  if (!ctx.existingindices.get(to)) {
    const toel = document.getElementById(b36encodenumber(to))
    if (!toel) {
      console.log("i can't find to")
      return
    }
    const toelf = toel.querySelector(".footer")
    if (!toelf) {
      console.log("i can't find toelf")
      return
    }
    if (toelf.childNodes) {
      const tn = document.createTextNode(" ")
      toelf.appendChild(tn)
    }
    const el = document.createElement("a")
    el.href = `/p/${b36encodenumber(from)}`
    el.innerText = `#${b36encodenumber(from)}`
    toelf.appendChild(el)
  }
  ctx.items = ctx.items.map((item) => {
    return (item.id !== to) ? item : { ...item, replies: [...item.replies, from] }
  })
}

function doinsert(id: number, insert: lrc.Insert, ctx: WSContext) {
  const idx = insert.utf16Index
  const s = insert.body
  ctx.insertMessage(id, idx, s)
}

function dodelete(id: number, del: lrc.Delete, ctx: WSContext) {
  const idx = del.utf16Start
  const idx2 = del.utf16End
  ctx.deleteMessage(id, idx, idx2)
}

