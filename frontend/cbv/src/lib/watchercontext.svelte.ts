import type * as cbv from "./types"
import { b36encodenumber } from "./utils"
import { getVolume, getWatcherVolume, onVolumeChange, onVolumeWatcherChange } from "./volume"
export class WatcherContext {
  watchthreads: Array<cbv.WatchThread> = $state([])
  ws: WebSocket

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
    const ws = new WebSocket("/ts")
    ws.onopen = () => console.log("i'm watching (4) you")
    ws.onmessage = (event) => {
      console.log(event)
      const twe = JSON.parse(event.data)
      // don't show that there's a reply if we're currently looking at this thread!
      if (document.getElementById(b36encodenumber(twe.id))) {
        return
      }
      this.watcherping.currentTime = 0
      this.watcherping.play()
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
    }
    ws.onerror = (error) => console.error(error)
    ws.onclose = () => console.log("see ya!")
    this.ws = ws
  }
}
