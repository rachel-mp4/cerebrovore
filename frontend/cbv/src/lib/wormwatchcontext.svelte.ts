import type * as cbv from "./types"
export class WormWatchContext extends EventTarget {
  wwqueue: Array<cbv.WormWatchEntry> = $state([])
  // if the server's timestamp is much higher than my timestamp,
  // i need to add to my timestamp in the future so that way i
  // can get my times to work out nicely. to get the amount to
  // add, i need to subtract my small number from the server's
  // bigger number. i will have a medium-big number "offset" that
  // i can now add to my Date.now() values in order to accurately
  // compare them with the server times
  offset: number = 0
  playingIndex: number | undefined = $state()
  start: number | undefined = $state()
  pause: number | undefined = $state()

  constructor(url: string) {
    super()
    document.addEventListener("cbv:wormwatch", (e) => {
      const ev = e as CustomEvent
      const wwe = ev.detail
      console.log(wwe)
      switch (wwe.type) {
        case "timeS": {
          this.offset = wwe.timestamp - Date.now()
          break
        }

        case "queue": {
          this.wwqueue = [...this.wwqueue, ...wwe.entries]
          break
        }

        case "start": {
          this.playingIndex = wwe.index
          this.start = wwe.timestamp + this.offset
          this.pause = undefined
          this.dispatchEvent(
            new CustomEvent("start", { detail: wwe })
          )
          break
        }

        case "pause": {
          this.playingIndex = wwe.index
          this.pause = wwe.timestamp + this.offset
          this.dispatchEvent(
            new CustomEvent("pause", { detail: wwe })
          )
          break
        }

        case "clear": {
          this.start = undefined
          this.pause = undefined
          this.wwqueue = []
          this.dispatchEvent(
            new CustomEvent("clear", { detail: wwe })
          )
          break
        }
      }
    })
  }

  getTimeToStart(): number | undefined {
    if (this.start === undefined) {
      return undefined
    }
    return this.start - Date.now() - this.offset
  }
  isPlaying(): boolean {
    return this.start !== undefined && this.pause === undefined
  }
}
