import type * as cbv from "./types"
export class WatcherContext {
  watchthreads: Array<cbv.WatchThread> = $state([])
  ws: WebSocket

  constructor() {
    const ws = new WebSocket("/ts")
    ws.onopen = () => console.log("i'm watching (4) you")
    ws.onmessage = (event) => {
      const twe = JSON.parse(event.data)
      this.watchthreads = this.watchthreads.map((wti) => {
        return wti.id === twe.id ? { ...wti, bumps: wti.bumps + 1, bumpedAt: Date.now() } : wti
      })
    }
    ws.onerror = (error) => console.error(error)
    ws.onclose = () => console.log("see ya!")
    this.ws = ws
  }
}
