import type * as cbv from "./types"
import { isMessage, isImage, isMedia } from "./types"
import * as lrc from '@rachel-mp4/lrcproto/gen/ts/lrc'

export class WSContext {
  existingindices: Map<number, boolean> = new Map()
  existinguris: Map<string, string> = new Map()
  items: Array<cbv.Item> = $state(new Array())
  log: Array<cbv.LogItem> = $state(new Array())
  topic: string = $state("")
  connected: boolean = $state(false)
  conncount = $state(0)
  ws: WebSocket | null = null
  color: number = $state(Math.floor(Math.random() * 16777216))

  threadID: string
  nick: string = "wanderer"
  handle: string = ""
  curMsg: string = $state("")
  myMessage: cbv.Message | undefined
  messageactive: boolean = false
  myImageRef: string | undefined
  myImageNonce: string | undefined
  myMedia: cbv.Media | undefined
  mediaactive: boolean = false

  shouldTransmit: boolean = $state(true)
  lrceventqueue: Array<lrc.Edit> = []

  constructor(threadID: string, defaultHandle: string, defaultNick: string, defaultColor: number) {
    console.log(threadID)
    this.threadID = threadID
    this.handle = defaultHandle
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
        const contentAddress = `TODO: come up with contentAddress scheme`
        pubImage(alt, contentAddress, this)
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
      formData.append("image", blob)
      formData.append("uuid", uuid)
      this.myImageNonce = uuid
      // fetch(`TODO: come up with imageUploadEndpoint`, {
      //   method: "POST",
      //   body: formData
      // }).then((response) => {
      //   if (response.ok) {
      //     response.json().then((data) => {
      //       if (this.myImageNonce === data.uuid) {
      //         this.myImageRef = data.imageRef
      //         this.myImageNonce = undefined
      //         console.log("here's myImageRef")
      //       } else {
      //         console.error("nonce mismatch!!!")
      //       }
      //     })
      //   } else {
      //     throw new Error(`HTTP ${response.status}`)
      //   }
      // }).catch((err) => { console.log(err) })
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
  setHandle = (handle: string) => {
    setHandle(handle, this)
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
        }
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

// this was for uploading nonce so that you could confirm anonymous posts
// const b64encodebytearray = (u8: Uint8Array): string => {
//   return btoa(String.fromCharCode(...u8))
// }

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
    setHandle(ctx.handle, ctx)
  };
  ws.onmessage = (event) => {
    console.log(event)
    parseEvent(event, ctx)
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
        externalID: ctx.handle
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
        externalID: ctx.handle
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

export const setHandle = (handle: string, ctx: WSContext) => {
  ctx.handle = handle
  const evt: lrc.Event = {
    msg: {
      oneofKind: "set",
      set: {
        externalID: handle
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

function parseEvent(binary: MessageEvent<any>, ctx: WSContext): boolean {
  const byteArray = new Uint8Array(binary.data);
  const event = lrc.Event.fromBinary(byteArray)
  switch (event.msg.oneofKind) {
    case "ping": {
      return false;
    }

    case "pong": {
      return false
    }

    case "init": {
      const id = event.msg.init.id ?? 0
      if (id === 0) return false
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
      return true
    }

    case "mediainit": {
      const id = event.msg.mediainit.id ?? 0
      if (id === 0) return false
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
      return true
    }

    case "pub": {
      const id = event.msg.pub.id ?? 0
      if (id === 0) return false
      ctx.pubMessage(id)
      ctx.pushToLog(id, byteArray, "pub")
      return false
    }

    case "mediapub": {
      const id = event.msg.mediapub.id ?? 0
      if (id === 0) return false
      const pub: cbv.LrcMediaPub = {
        alt: event.msg.mediapub.alt ?? "",
        contentAddress: event.msg.mediapub.contentAddress
      }
      ctx.pubMedia(id, pub)
      ctx.pushToLog(id, byteArray, "pub")
      return false
    }

    case "insert": {
      const id = event.msg.insert.id ?? 0
      if (id === 0) return false
      ctx.pushToLog(id, byteArray, "insert")
      doinsert(id, event.msg.insert, ctx)
      return false
    }

    case "delete": {
      const id = event.msg.delete.id ?? 0
      if (id === 0) return false
      ctx.pushToLog(event.msg.delete.id ?? 0, byteArray, "delete")
      dodelete(id, event.msg.delete, ctx)
      return false
    }


    case "mute": {
      const id = event.msg.mute.id ?? 0
      if (id === 0) return false
      ctx.initMute(id)
      return false
    }

    case "unmute": {
      return false
    }

    case "set": {
      return false
    }

    case "get": {
      if (event.msg.get.connected !== undefined) {
        ctx.setConncount(event.msg.get.connected)
      }
      if (event.msg.get.topic !== undefined) {
        ctx.setTopic(event.msg.get.topic)
      }
      return false
    }
    //TODO: better logging system so that way even non hrt messages
    // can have the background effect!
    case "editbatch": {
      const id = event.id ?? 0
      if (id === 0) {
        return false
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
      return false

    }

  }
  return false
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

