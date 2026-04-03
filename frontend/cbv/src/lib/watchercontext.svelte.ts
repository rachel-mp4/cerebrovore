import type * as cbv from "./types"
import { b36encodenumber } from "./utils"
import { getVolume, getWatcherVolume, onVolumeChange, onVolumeWatcherChange } from "./volume"
export class WatcherContext {
  watchthreads: Array<cbv.WatchThread> = $state([])
  newthreads: Array<cbv.WatchThread> = $state([])

  volume: number
  watcherping: HTMLAudioElement = new Audio('/wav/shortnotif.wav')
  watchervolume: number

  constructor() {
    this.volume = getVolume()
    this.watchervolume = getWatcherVolume()
    this.watcherping.volume = this.volume * this.watchervolume
    onVolumeChange((e) => {
      this.volume = e.detail.volume
      this.watcherping.volume = this.volume * this.watchervolume
    })
    onVolumeWatcherChange((e) => {
      this.watchervolume = e.detail.volume
      this.watcherping.volume = this.volume * this.watchervolume
    })
    document.addEventListener("cbv:watcher", (e) => {
      const ev = e as CustomEvent
      const twe = ev.detail
      // don't show that there's a reply if we're currently looking at this thread!
      if (document.getElementById(b36encodenumber(twe.id))) {
        return
      }
      this.watcherping.currentTime = 0
      this.watcherping.play()
      const isNew = twe.new ?? false
      if (isNew) {
        const newt: cbv.WatchThread = {
          type: 'watchthread',
          id: twe.id,
          ...(twe.topic && { topic: twe.topic }),
          bumps: 0,
          bumpedAt: Date.now(),
        }
        this.newthreads = [newt, ...this.newthreads]
        return
      }
      if (this.watchthreads.find((wti) => wti.id === twe.id)) {
        this.watchthreads = this.watchthreads.map((wti) => {
          return wti.id === twe.id ? { ...wti, bumps: wti.bumps + 1, bumpedAt: Date.now(), ...(twe.bumpLimit && { bumpLimit: twe.bumpLimit }) } : wti
        })
      } else {
        const newwt: cbv.WatchThread = {
          type: 'watchthread',
          id: twe.id,
          ...(twe.topic && { topic: twe.topic }),
          ...(twe.bumpLimit && { bumpLimit: twe.bumpLimit }),
          bumps: 1,
          bumpedAt: Date.now(),
        }
        this.watchthreads = [...this.watchthreads, newwt]
      }
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
