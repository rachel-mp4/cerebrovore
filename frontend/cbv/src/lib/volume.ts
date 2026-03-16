export function emitVolumeChange(volume: number) {
  localStorage.setItem("volume", String(volume))
  document.dispatchEvent(new CustomEvent("volumechange", { detail: { volume } }))
}

export function onVolumeChange(handler: (e: any) => void) {
  document.addEventListener("volumechange", handler)
  return () => document.removeEventListener("volumechange", handler)
}

export function getVolume(): number {
  const volume = Number(localStorage.getItem("volume"))
  if (!Number.isFinite(volume)) {
    localStorage.setItem("volume", String(1))
    return 1
  }
  const clamped = Math.min(Math.max(volume, 0), 1)
  if (volume !== clamped) {
    localStorage.setItem("volume", String(clamped))
    return clamped
  }
  return volume
}

export function emitVolumeWatcherChange(volume: number) {
  localStorage.setItem("volumewatcher", String(volume))
  document.dispatchEvent(new CustomEvent("volumewatcherchange", { detail: { volume } }))
}

export function onVolumeWatcherChange(handler: (e: any) => void) {
  document.addEventListener("volumewatcherchange", handler)
  return () => document.removeEventListener("volumewatcherchange", handler)
}

export function getWatcherVolume(): number {
  const volume = Number(localStorage.getItem("volumewatcher"))
  if (!Number.isFinite(volume)) {
    localStorage.setItem("volumewatcher", String(1))
    return 1
  }
  const clamped = Math.min(Math.max(volume, 0), 1)
  if (volume !== clamped) {
    localStorage.setItem("volumewatcher", String(clamped))
    return clamped
  }
  return volume
}

export function emitVolumeWormWatchChange(volume: number) {
  localStorage.setItem("volumewormwatch", String(volume))
  document.dispatchEvent(new CustomEvent("volumewormwatchchange", { detail: { volume } }))
}

export function onVolumeWormWatchChange(handler: (e: any) => void) {
  document.addEventListener("volumewormwatchchange", handler)
  return () => document.removeEventListener("volumewormwatchchange", handler)
}

export function getWormWatchVolume(): number {
  const volume = Number(localStorage.getItem("volumewormwatch"))
  if (!Number.isFinite(volume)) {
    localStorage.setItem("volumewormwatch", String(1))
    return 1
  }
  const clamped = Math.min(Math.max(volume, 0), 1)
  if (volume !== clamped) {
    localStorage.setItem("volumewormwatch", String(clamped))
    return clamped
  }
  return volume
}

export function emitVolumeFocusPingChange(volume: number) {
  localStorage.setItem("volumefocusping", String(volume))
  document.dispatchEvent(new CustomEvent("volumefocuspingchange", { detail: { volume } }))
}

export function onVolumeFocusPingChange(handler: (e: any) => void) {
  document.addEventListener("volumefocuspingchange", handler)
  return () => document.removeEventListener("volumefocuspingchange", handler)
}

export function getFocusPingVolume(): number {
  const volume = Number(localStorage.getItem("volumefocusping"))
  if (!Number.isFinite(volume)) {
    localStorage.setItem("volumefocusping", String(1))
    return 1
  }
  const clamped = Math.min(Math.max(volume, 0), 1)
  if (volume !== clamped) {
    localStorage.setItem("volumefocusping", String(clamped))
    return clamped
  }
  return volume
}

export function emitVolumePingChange(volume: number) {
  localStorage.setItem("volumeping", String(volume))
  document.dispatchEvent(new CustomEvent("volumepingchange", { detail: { volume } }))
}

export function onVolumePingChange(handler: (e: any) => void) {
  document.addEventListener("volumepingchange", handler)
  return () => document.removeEventListener("volumepingchange", handler)
}

export function getPingVolume(): number {
  const volume = Number(localStorage.getItem("volumeping"))
  if (!Number.isFinite(volume)) {
    localStorage.setItem("volumeping", String(1))
    return 1
  }
  const clamped = Math.min(Math.max(volume, 0), 1)
  if (volume !== clamped) {
    localStorage.setItem("volumeping", String(clamped))
    return clamped
  }
  return volume
}
