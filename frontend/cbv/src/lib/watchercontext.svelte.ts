import type * as cbv from "./types"
import { b36encodenumber } from "./utils"
import { getVolume, getWatcherVolume, getPingVolume, getFocusPingVolume, onVolumeChange, onVolumeWatcherChange, onVolumeFocusPingChange, onVolumePingChange } from "./volume"
export class WatcherContext {
  watchthreads: Array<cbv.WatchThread> = $state([])
  newthreads: Array<cbv.WatchThread> = $state([])

  volume: number
  watchervolume: number
  focuspingvolume: number
  pingvolume: number
  watcherping: HTMLAudioElement = new Audio('/wav/shortnotif.wav')
  forumfocusping: HTMLAudioElement = new Audio('/wav/shortnotif.wav')
  forumping: HTMLAudioElement = new Audio('/wav/shortnotif.wav')
  isforum: boolean
  // need the chirps array to avoid a memory leak, the issue is that we can get
  // sent lrc messages that don't ever get published to the site, so they never
  // flow through the watcher socket (also, we might not be watching the thread etc...)
  // so they'd otherwise never be removed from the set
  chirps: string[]
  alreadychirped: Set<string>
  lastfocus: boolean

  constructor(isforum: boolean) {
    this.isforum = isforum
    this.volume = getVolume()
    this.watchervolume = getWatcherVolume()
    this.focuspingvolume = getFocusPingVolume()
    this.pingvolume = getPingVolume()
    this.watcherping.volume = this.volume * this.watchervolume
    onVolumeChange((e) => {
      this.volume = e.detail.volume
      this.watcherping.volume = this.volume * this.watchervolume
      this.forumfocusping.volume = this.volume * this.focuspingvolume
      this.forumping.volume = this.volume * this.pingvolume
    })
    onVolumeWatcherChange((e) => {
      this.watchervolume = e.detail.volume
      this.watcherping.volume = this.volume * this.watchervolume
    })
    onVolumePingChange((e) => {
      this.pingvolume = e.detail.volume
      this.forumping.volume = this.volume * this.pingvolume
    })
    onVolumeFocusPingChange((e) => {
      this.focuspingvolume = e.detail.volume
      this.forumfocusping.volume = this.volume * this.focuspingvolume
    })
    this.chirps = []
    this.alreadychirped = new Set<string>()
    const cc = new BroadcastChannel("chirpchan")
    cc.postMessage("cc:focus")
    this.lastfocus = true
    cc.onmessage = (event) => {
      if (event.data == "cc:focus") {
        this.lastfocus = false
        return
      }
      if (typeof event.data === "string") {
        this.alreadychirped.add(event.data)
        this.chirps.push(event.data)
        while (this.chirps.length > 100) {
          const chirp = this.chirps.shift()
          if (chirp) this.alreadychirped.delete(chirp)
        }
      }
    }
    window.addEventListener("focus", () => {
      cc.postMessage("cc:focus")
      this.lastfocus = true
    })

    const handleWatcherEvent = (twe: any) => {
      const npid = b36encodenumber(twe.pid ?? twe.tid) //twe.tid is always defined, but pid is not always defined
      const ntid = b36encodenumber(twe.tid)
      // don't show that there's a reply if we're currently looking at this thread!
      if (document.getElementById(ntid)) {
        if (!isforum) {
          return
        }
        if (document.hidden || !document.hasFocus()) {
          this.forumping.currentTime = 0
          this.forumping.play()
        } else {
          this.forumfocusping.currentTime = 0
          this.forumfocusping.play()
        }
        this.alreadychirped.delete(npid)
        return
      }
      if (this.lastfocus) {
        if (!this.alreadychirped.has(npid)) {
          this.watcherping.currentTime = 0
          this.watcherping.play()
        }
      }
      this.alreadychirped.delete(npid)
      const isNew = twe.new ?? false
      if (isNew) {
        const newt: cbv.WatchThread = {
          type: 'watchthread',
          id: twe.tid,
          ...(twe.topic && { topic: twe.topic }),
          bumps: 0,
          bumpedAt: Date.now(),
        }
        this.newthreads = [newt, ...this.newthreads]
        return
      }
      if (this.watchthreads.find((wti) => wti.id === twe.tid)) {
        this.watchthreads = this.watchthreads.map((wti) => {
          return wti.id === twe.tid ? { ...wti, bumps: wti.bumps + 1, bumpedAt: Date.now(), ...(twe.bumpLimit && { bumpLimit: twe.bumpLimit }) } : wti
        })
      } else {
        const newwt: cbv.WatchThread = {
          type: 'watchthread',
          id: twe.tid,
          ...(twe.topic && { topic: twe.topic }),
          ...(twe.bumpLimit && { bumpLimit: twe.bumpLimit }),
          bumps: 1,
          bumpedAt: Date.now(),
        }
        this.watchthreads = [...this.watchthreads, newwt]
      }
    }

    // @ts-ignore
    const wes = cbvWSBuffer?.watcher
    if (wes !== undefined) {
      wes.forEach(handleWatcherEvent)
    }

    document.addEventListener("cbv:watcher", (e) => {
      const ev = e as CustomEvent
      const twe = ev.detail
      handleWatcherEvent(twe)
    })
  }

  rmIdx(idx: number) {
    const newt = this.newthreads[idx]
    if (newt === undefined) {
      return
    }
    this.newthreads = this.newthreads.filter((_, i) => i !== idx)
  }

  watchIdx(idx: number) {
    const newt = this.newthreads[idx]
    if (newt === undefined) {
      return
    }
    this.newthreads = this.newthreads.filter((_, i) => i !== idx)
    const endpoint = `/w/${b36encodenumber(newt.id)}`
    fetch(endpoint, {
      method: "POST"
    }).then((resp) => console.log(resp))
  }
}
