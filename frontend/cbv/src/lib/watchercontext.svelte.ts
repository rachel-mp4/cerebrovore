import type * as cbv from "./types"
export class WatcherContext {
  watchthreads: Array<cbv.WatchThread> = $state([])
  ws: WebSocket

  constructor() {
    const ws = new WebSocket("/ts")
    ws.onopen = () => console.log("i'm watching (4) you")
    ws.onmessage = (event) => {
      console.log(event)
      const twe = JSON.parse(event.data)
      if (this.watchthreads.find((wti) => wti.id === twe.id)) {
        this.watchthreads = this.watchthreads.map((wti) => {
          return wti.id === twe.id ? { ...wti, bumps: wti.bumps + 1, bumpedAt: Date.now() } : wti
        })
      } else {
        const newwt: cbv.WatchThread = {
          type: 'watchthread',
          id: twe.id,
          ...(twe.topic && { topic: twe.topic }),
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
