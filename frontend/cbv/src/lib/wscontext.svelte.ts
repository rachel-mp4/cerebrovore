import type * as cbv from "./types"
import { isMessage, isImage, isMedia } from "./types"
import { b36encodenumber } from "./utils"
import * as lrc from '@rachel-mp4/lrcproto/gen/ts/lrc'
import { numToHex } from "./colors"
import { getFocusPingVolume, getPingVolume, getVolume, onVolumeChange, onVolumeFocusPingChange, onVolumePingChange } from "./volume"


// the basic idea is that we start out in the ready state, and when we insert we transition into the started state
// at which point either we will recieve our init from the server and enter the normal happy state, or we can send
// in which case we enter the sentWithoutReceiving state. from the sentWithoutReceiving state, we can initialize a
// second message, which will appear to us just as the first one did, however the critical thing here is that we
// can no longer send it, and in fact it won't even send its lrc init event until we recieve our response from the
// first one. the idea here is that in all likelihood, a user should be able to spam message a lil bit, & they should
// be able to send one message fast even with high ping, and they should even be able to seamlessly start their next
// message with high ping, but the response from the first one should arrive by the time they have both sent the first
// message, started the second, and finished the second. in order to have a truly robust system, we'd need to up the
// complexity here a ton and track every message i send & recieve individually, which sucks, so this hopefully should
// be a middleground that's only moderate complexity but which handles 99.99% of legitimate cases
type lrcState = ready | sent | received | pubbedWithoutReceiving
type mediaUploadState = ready | uploading | uploaded

type ready = {
  kind: "ready"
}
const ready: ready = {
  kind: "ready"
}
type uploading = {
  kind: "uploading"
  nonce: string
}
type uploaded = {
  kind: "uploaded"
  cid: string
}

type sent = {
  kind: "sent"
  fakeId: number  // i know i will hate me for this in the future, but i think Date.now() is an easy enough solution 
}                 // because no collisions, and it's guaranteed to be above uint32 max

type received = {
  kind: "recieved"
  init: lrc.Init
}

type pubbedWithoutReceiving = {
  kind: "pubbedWithoutReceiving"
  fakeId: number
}

export class WSContext {
  existingindices: Map<number, boolean> = new Map()
  items: Array<cbv.Item> = $state(new Array())
  log: Array<cbv.LogItem> = $state(new Array())
  logidx: number = 0
  topic: string = $state("")
  connected: boolean = $state(false)
  conncount = $state(0)
  ws: WebSocket | null = null
  ts: WebSocket | null = null
  color: number = $state(Math.floor(Math.random() * 16777216))
  systemMessage: string | undefined = $state()
  replyLimit: boolean = false
  rttping: number = $state(0)
  rttpingstart: number | undefined
  pinginterval: number | undefined
  shouldCalcPing: boolean

  nick: string = "wanderer"
  anon: boolean = false
  curMsg: string = $state("")
  curMsgId: string | undefined = $state()
  myMessageState: lrcState = ready
  curImageBlobURL: string | undefined = $state()
  myMediaState: lrcState = $state(ready)
  myMediaUploadState: mediaUploadState = $state(ready)

  cc: BroadcastChannel

  volume: number
  ping: HTMLAudioElement = new Audio('/wav/notif.wav')
  pingvolume: number
  focusping: HTMLAudioElement = new Audio('/wav/shortnotif.wav')
  focuspingvolume: number

  lrceventqueue: Array<lrc.Edit> = []

  constructor(defaultNick: string, defaultColor: number) {
    this.cc = new BroadcastChannel("chirpchan")
    this.nick = defaultNick
    this.color = defaultColor
    this.volume = getVolume()
    this.pingvolume = getPingVolume()
    this.focuspingvolume = getFocusPingVolume()
    this.ping.volume = this.volume * this.pingvolume
    this.focusping.volume = this.volume * this.focuspingvolume
    onVolumeChange((e) => {
      this.volume = e.detail.volume
      this.ping.volume = this.volume * this.pingvolume
      this.focusping.volume = this.volume * this.focuspingvolume
    })
    onVolumePingChange((e) => {
      this.pingvolume = e.detail.volume
      this.ping.volume = this.volume * this.pingvolume
    })
    onVolumeFocusPingChange((e) => {
      this.focuspingvolume = e.detail.volume
      this.focusping.volume = this.volume * this.focuspingvolume
    })
    this.anon = localStorage.getItem("anon") !== null
    this.shouldCalcPing = localStorage.getItem("displayPing") !== null
    const log = new Array<cbv.LogItem>(200)
    for (let i = 0; i < 200; i++) {
      log[i] = { type: "init", id: 0, binary: "", time: 0, key: Math.random(), ignore: true }
    }
    this.log = log
  }

  connect(url: string) {
    this.ws?.close()
    this.ts?.close()
    connectTo(url, this)
  }

  reconnect = (url: string) => {
    this.ws?.close()
    this.ts?.close()
    connectTo(url, this)
    this.items = []
  }

  disconnect = () => {
    this.ws?.close()
    this.ts?.close()
    this.ws = null
    this.ts = null
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

  insertLineBreak = (): number => {
    switch (this.myMessageState.kind) {
      case "ready": {
        return 0 // do nothing
      }

      case "pubbedWithoutReceiving": {
        return 2 // wiggle (not allowed)
      }

      case "sent": {
        pubMessage(this)
        const fake = this.myMessageState.fakeId
        this.items = this.items.map((item) =>
          item.id === fake && isMessage(item)
            ? { ...item, ignore: true, lrcdata: { ...item.lrcdata, body: this.curMsg + "WARNING: YOU MAY BE DISCONNECTED // don't submit message so fast", pub: true } }
            : item)
        this.myMessageState = { kind: "pubbedWithoutReceiving", fakeId: this.myMessageState.fakeId }
        this.curMsg = ""
        this.curMsgId = undefined
        return 1 // send
      }

      case "recieved": {
        pubMessage(this)
        const id = postMessage(this.myMessageState.init, this.curMsg)
        this.items = this.items.map((item) => item.id === id
          && isMessage(item)
          ? { ...item, ignore: true, lrcdata: { ...item.lrcdata, body: this.curMsg, pub: true } }
          : item)
        this.myMessageState = ready
        this.curMsg = ""
        this.curMsgId = undefined
        return 1 // send
      }
    }
  }


  pubImage = (alt: string) => {
    this.curImageBlobURL = undefined
    switch (this.myMediaState.kind) {
      case "ready":
      case "pubbedWithoutReceiving": {
        return
      }

      case "sent": {
        switch (this.myMediaUploadState.kind) {
          case "ready":
          case "uploading": {
            pubImage(alt, undefined, this)
            break
          }

          case "uploaded": {
            const contentAddress = `/blob?cid=${this.myMediaUploadState.cid}`
            pubImage(alt, contentAddress, this)
            break
          }
        }
        this.myMediaState = { kind: "pubbedWithoutReceiving", fakeId: this.myMediaState.fakeId }
        this.myMediaUploadState = ready
        return
      }

      case "recieved": {
        switch (this.myMediaUploadState.kind) {
          case "ready":
          case "uploading": {
            pubImage(alt, undefined, this)
            break
          }

          case "uploaded": {
            const contentAddress = `/blob?cid=${this.myMediaUploadState.cid}`
            pubImage(alt, contentAddress, this)
            postImage(this.myMediaState.init, this.myMediaUploadState.cid, alt)
            break
          }
        }
        this.myMediaState = ready
        this.myMediaUploadState = ready
      }
    }
  }


  cancelImage = () => {
    this.curImageBlobURL = undefined
    this.myMediaUploadState = ready
    switch (this.myMediaState.kind) {
      case "ready":
      case "pubbedWithoutReceiving": {
        return
      }

      case "sent":
      case "recieved": {
        pubImage(undefined, undefined, this)
        this.myMediaState = ready
        return
      }
    }
  }

  initImage = (blob: File, blobUrl: string) => {
    if (this.myMediaState.kind === "ready" && this.myMediaUploadState.kind === "ready") {
      this.curImageBlobURL = blobUrl
      initImage(this)
      const fake = Date.now()
      this.pushMyItem({ type: 'image', id: fake, lrcdata: { mine: true, muted: false }, replies: [], })
      this.myMediaState = {
        kind: "sent",
        fakeId: fake
      }
      const uuid = crypto.randomUUID()
      const formData = new FormData()
      formData.append("file", blob)
      formData.append("uuid", uuid)
      this.myMediaUploadState = {
        kind: "uploading",
        nonce: uuid
      }
      fetch(`/blob`, {
        method: "POST",
        body: formData
      }).then((response) => {
        if (response.ok) {
          response.json().then((data) => {
            if (this.myMediaUploadState.kind === "uploading" && this.myMediaUploadState.nonce === data.uuid) {
              this.myMediaUploadState = {
                kind: "uploaded",
                cid: data.cid
              }
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
    if (this.myMessageState.kind === "ready") {
      const init = initMessage(this)
      const fake = Date.now()
      this.pushMyItem({ type: 'message', id: fake, lrcdata: { mine: true, muted: false, body: "", init: init }, replies: [], })
      this.myMessageState = {
        kind: "sent",
        fakeId: fake
      }
      document.dispatchEvent(new CustomEvent("lrc:scroll"))
    }
    insertMessage(idx, s, this)
    this.curMsg = insertSIntoAStringAtIdx(s, this.curMsg, idx)
  }

  delete = (idx: number, idx2: number) => {
    if (this.myMessageState.kind === "ready") {
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
    localStorage.setItem('nick', nick)
  }

  setAnon = (anon: boolean) => {
    this.anon = anon
    if (anon) {
      localStorage.setItem("anon", "yes")
    } else {
      localStorage.removeItem("anon")
    }
  }

  setColor = (color: number) => {
    setColor(color, this)
    localStorage.setItem('color', String(color))
  }

  setTopic = (topic: string) => {
    console.log("new topic:", topic)
    this.topic = topic
  }

  setConncount = (cc: number) => {
    this.conncount = cc
  }

  pingServer = () => {
    pingServer(this)
  }

  pushItem = (item: cbv.Item, init?: lrc.Init) => {
    if (this.existingindices.get(item.id)) {
      console.log("you tried to push an item who exists!")
      return
    }
    this.cc.postMessage(b36encodenumber(item.id))
    if (document.hidden || !document.hasFocus()) {
      this.ping.currentTime = 0
      this.ping.play()
    } else if (!item.lrcdata.mine) {
      this.focusping.currentTime = 0
      this.focusping.play()
    }
    if (item.lrcdata.mine) {
      if (isMessage(item)) {
        switch (this.myMessageState.kind) {
          case "ready": {
            console.error("your item was pushed while you were ready (didn't request init). i'm not sure how to proceed")
            return
          }

          case "recieved": {
            console.error("your item was pushed again after you already recieved it. i'm not sure how to proceed")
            return
          }

          case "sent": {
            const fake = this.myMessageState.fakeId
            this.items = this.items.filter((item) =>
              item.id !== fake
            )
            this.myMessageState = {
              kind: "recieved",
              init: init! // init should be set since we only know if it's mine if we're coming from an init
            }
            break
          }

          case "pubbedWithoutReceiving": {
            const fake = this.myMessageState.fakeId
            this.items = this.items.filter((item) =>
              item.id !== fake
            )
            this.myMessageState = ready
            if (this.lrceventqueue.length !== 0) {
              const init = initMessage(this)
              const fake = Date.now()
              this.pushMyItem({ type: 'message', id: fake, lrcdata: { mine: true, muted: false, body: "", init: init }, replies: [], })
              this.myMessageState = {
                kind: "sent",
                fakeId: fake
              }
              document.dispatchEvent(new CustomEvent("lrc:scroll"))
              this.starttransmit()
            }
            break
          }
        }
      } else if (isMedia(item)) {
        switch (this.myMediaState.kind) {
          case "ready": {
            console.error("your item was pushed while you were ready (didn't request init). i'm not sure how to proceed")
            return
          }

          case "recieved": {
            console.error("your item was pushed again after you already recieved it. i'm not sure how to proceed")
            return
          }

          case "sent": {
            const fake = this.myMediaState.fakeId
            this.items = this.items.filter((item) =>
              item.id !== fake
            )
            this.myMediaState = {
              kind: "recieved",
              init: init!
            }
            break
          }

          case "pubbedWithoutReceiving": {
            const fake = this.myMediaState.fakeId
            this.items = this.items.filter((item) =>
              item.id !== fake
            )
            this.myMediaState = ready
            break
          }
        }
      } else {
        console.error("what's happening? recieved non message non media from me", item)
      }
    }
    this.items.push(item)
    document.dispatchEvent(new CustomEvent("lrc:append"))
    this.existingindices.set(item.id, true)
  }

  pushMyItem = (item: cbv.Item) => {
    this.items.push(item)
    document.dispatchEvent(new CustomEvent("lrc:append"))
  }

  initMessage = (id: number, init: cbv.LrcInit, mine?: lrc.Init) => {
    if (mine) {
      this.curMsgId = b36encodenumber(id)
    }
    if (this.existingindices.get(id)) {
      this.items = this.items.map((item: cbv.Item) => {
        return item.id === id && isMessage(item)
          ? { ...item, type: "message", lrcdata: { ...item.lrcdata, init: init } }
          : item
      })
    } else {
      this.pushItem({
        type: 'message',
        id: id,
        lrcdata: {
          body: '',
          mine: mine !== undefined,
          muted: false,
          init: init,
        },
        replies: []
      }, mine)
    }
  }

  initMedia = (id: number, init: cbv.LrcInit, mine?: lrc.MediaInit) => {
    if (this.existingindices.get(id)) {
      this.items = this.items.map((item: cbv.Item) => {
        return item.id === id && isImage(item)
          ? { ...item, type: "image", lrcdata: { ...item.lrcdata, init: init } }
          : item
      })
    } else {
      this.pushItem({
        type: 'image',
        id: id,
        lrcdata: {
          mine: mine !== undefined,
          muted: false,
          init: init,
        },
        replies: []
      }, mine)
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
          ? { ...item, type: "message", lrcdata: { ...item.lrcdata, pub: true }, pubAt: Date.now() }
          : item
      })
    } else {
      this.pushItem({
        type: "message",
        id: id,
        lrcdata: {
          mine: false,
          muted: false,
          body: "",
        },
        replies: [],
        pubAt: Date.now()
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
            },
            pubAt: Date.now()
          }
          : item
      })
    } else {
      this.pushItem({
        type: "image",
        id: id,
        lrcdata: {
          mine: false,
          muted: false,
          pub: pub,
        },
        replies: [],
        pubAt: Date.now()
      })
    }
  }

  insertMessage = (id: number, idx: number, s: string) => {
    if (this.existingindices.get(id)) {
      this.items = this.items.map((item: cbv.Item) => {
        return item.id === id && isMessage(item) && !item.ignore
          ? { ...item, type: "message", lrcdata: { ...item.lrcdata, body: insertSIntoAStringAtIdx(s, item.lrcdata.body, idx) } }
          : item
      })
    } else {

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
        return item.id === id && isMessage(item) && !item.ignore
          ? { ...item, type: "message", lrcdata: { ...item.lrcdata, body: deleteFromAStringBetweenIdxs(item.lrcdata.body, idx1, idx2) } }
          : item
      })
    } else {

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
    var color: number | undefined
    if (type == "init" || type == "pub") {
      const item = this.items.find((item) => item.id === id)
      color = item?.lrcdata.init?.color
    }
    this.log = this.log.map((li, idx) => idx === this.logidx
      ? { id: id, color: color, binary: bstring, time: time, type: type, key: Math.random(), ignore: false }
      : li)
    this.logidx = (this.logidx + 1) % 200
  }
}

const postMessage = (init: lrc.Init, msg: string): number => {
  const fd = new FormData()
  const id = init.id ?? 0
  fd.append("id", b36encodenumber(id))
  fd.append("color", numToHex(init.color ?? 0))
  fd.append("nick", init.nick ?? "")
  fd.append("body", msg)
  if (init.externalID === undefined) {
    fd.append("anon", "yes")
  }
  if (init.nonce) {
    fd.append("nonce", b64encodebytearray(init.nonce))
  }
  const endpoint = window.location.href
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
  return id
}

const postImage = (mediainit: lrc.MediaInit, cid: string, alt: string): number => {
  const fd = new FormData()
  const id = mediainit.id ?? 0
  fd.append("id", b36encodenumber(id))
  fd.append("color", numToHex(mediainit.color ?? 0))
  fd.append("nick", mediainit.nick ?? "")
  fd.append("cid", cid)
  fd.append("alt", alt)
  if (mediainit.externalID === undefined) {
    fd.append("anon", "yes")
  }
  if (mediainit.nonce) {
    fd.append("nonce", b64encodebytearray(mediainit.nonce))
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
  return id
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
  const ws = new WebSocket(`${url}/ws`, "lrc.v1");
  ws.binaryType = "arraybuffer";
  ws.onopen = () => {
    console.log("connected")
    ctx.connected = true
    getTopic(ctx)
    setNick(ctx.nick, ctx)
    setColor(ctx.color, ctx)
  };
  ws.onmessage = (event) => {
    switch (parseEvent(event, ctx)) {
      case 1:
        document.dispatchEvent(new CustomEvent("lrc:scrollIfAttached"))
        break;
      case 2:
        document.dispatchEvent(new CustomEvent("lrc:scroll"))
    }
  }
  ws.onclose = (e) => {
    console.log(e)
    if (ws === ctx.ws) {
      if (ctx.connected === true) {
        const timeout = setTimeout(() => {
          window.location.reload()
        }, 10000)
        ctx.systemMessage = "DISCONNECTED! refreshing in 10 seconds, click to cancel"
        document.addEventListener("click", () => {
          ctx.systemMessage = "refresh aborted!"
          clearTimeout(timeout)
        })
      }
    }
  };
  ws.onerror = (event) => {
    console.log("errored:", event)
    console.log("readyState:", ws.readyState)
    if (ws === ctx.ws) {
      if (ctx.connected === true) {
        const timeout = setTimeout(() => {
          window.location.reload()
        }, 10000)
        ctx.systemMessage = "DISCONNECTED! refreshing in 10 seconds, click to cancel"
        document.addEventListener("click", () => {
          ctx.systemMessage = "refresh aborted!"
          clearTimeout(timeout)
        })

      }
    }
  }
  ctx.ws = ws
  if (ctx.shouldCalcPing) {
    ctx.pinginterval = setInterval(() => {
      ctx.rttpingstart = Date.now()
      ctx.pingServer()
    }, 3000)
  }

  const handleThreadEvent = (tse: any) => {
    if (tse.remaining !== undefined) {
      const ls = document.getElementById("left-sidebar")
      if (ls !== null) {
        ls.style.setProperty("--remaining", tse.remaining)
      }
    }
    if (tse.id !== undefined) {
      if (tse.deleted === true) {
        if (ctx.existingindices.get(tse.id)) {
          ctx.items = ctx.items.filter((item) => !(item.id === tse.id))
        } else {
          document.getElementById(b36encodenumber(tse.id))?.remove()
        }
        return
      }
      ctx.items = ctx.items.map((item) => {
        return (item.id === tse.id)
          ? {
            ...item,
            ...(tse.username && { username: tse.username }),
            ...(tse.renderedHTML && { renderedHTML: tse.renderedHTML })
          }
          : item
      })
    }
    if (tse.systemMessage !== undefined) {
      ctx.systemMessage = tse.systemMessage
    }
    if (tse.bumpLimit !== undefined) {
      ctx.systemMessage = "bump limit reached, look to find a new thread"
    }
    if (tse.replyLimit !== undefined) {
      ctx.replyLimit = true
      ctx.systemMessage = "reply limit reached; thread archived. you can continue messaging, but everything will be lost to history"
    }
  }

  // @ts-ignore
  const tes = cbvWSBuffer?.thread
  if (tes !== undefined) {
    tes.forEach(handleThreadEvent)
  }

  document.addEventListener('cbv:thread', (e) => {
    const ev = e as CustomEvent
    const tse = ev.detail
    handleThreadEvent(tse)
  })
}

export const initMessage = (ctx: WSContext): lrc.Init => {
  const init: lrc.Init = {
    nick: ctx.nick,
    ...(ctx.anon && { externalID: "" }),
    color: ctx.color,

  }
  const evt: lrc.Event = {
    msg: {
      oneofKind: "init",
      init: init
    }
  }
  const byteArray = lrc.Event.toBinary(evt)
  ctx.ws?.send(byteArray)
  return init
}

export const initImage = (ctx: WSContext) => {
  const evt: lrc.Event = {
    msg: {
      oneofKind: "mediainit",
      mediainit: {
        nick: ctx.nick,
        color: ctx.color,
        ...(ctx.anon && { externalID: "" }),
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
  switch (ctx.myMessageState.kind) {
    case "ready": {
      console.error("called insert while message is ready! (hasn't been initialized)")
      return
    }

    case "sent":
    case "recieved": {
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
      return
    }

    case "pubbedWithoutReceiving": {
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
      return
    }
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
  switch (ctx.myMessageState.kind) {
    case "ready": {
      console.error("called delete while message is ready! (hasn't been initialized)")
      return
    }

    case "sent":
    case "recieved": {
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
      return
    }

    case "pubbedWithoutReceiving": {
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
      return
    }
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

export const pingServer = (ctx: WSContext) => {
  const evt: lrc.Event = {
    msg: {
      oneofKind: "ping",
      ping: lrc.Ping,
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
      if (ctx.rttpingstart === undefined) {
        return 0
      }
      ctx.rttping = Date.now() - ctx.rttpingstart
      ctx.rttpingstart = undefined
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
        ...(id && { id: id }),
        ...(color && { color: color }),
        ...(nick && { nick: nick }),
        ...(handle && { handle: handle }),
        ...(nonce && { nonce: nonce }),
      }
      ctx.initMessage(id, init, mine ? event.msg.init : undefined)
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
        ...(id && { id: id }),
        ...(color && { color: color }),
        ...(nick && { nick: nick }),
        ...(handle && { handle: handle }),
        ...(nonce && { nonce: nonce }),
      }
      ctx.initMedia(id, init, mine ? event.msg.mediainit : undefined)
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

    case "kick": {
      window.location.href = window.location.origin
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
  const fn = b36encodenumber(from)
  const to = attach.to
  const tn = b36encodenumber(to)
  const fromel = document.getElementById(fn)
  const toel = document.getElementById(tn)
  if (fromel && toel?.classList.contains("you")) {
    const ss = fromel.querySelectorAll(`a[href="/p/${tn}"]`)
    ss.forEach((s) => {
      s.classList.add("you")
    })
  }
  if (!ctx.existingindices.get(to)) {
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

