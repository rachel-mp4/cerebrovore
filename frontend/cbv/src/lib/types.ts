export type Item = Message | Media | Enby

export type Enby = {
  type: 'enby'
  id: number
  lrcdata: LrcBaseItem
}

export type Message = {
  type: 'message'
  id: number
  lrcdata: LrcMessage
}

export type Media = Image

export type Image = {
  type: 'image'
  id: number
  lrcdata: LrcMedia
}

export type LrcMessage = LrcBaseItem & {
  body: string
  pub?: LrcMessagePub
}

export type LrcMedia = LrcBaseItem & {
  pub?: LrcMediaPub
}

export type LrcBaseItem = {
  mine: boolean
  muted: boolean
  init?: LrcInit
}

export type LrcInit = {
  color?: number
  nick?: string
  handle?: string
  nonce?: Uint8Array
}

export type LrcMediaPub = {
  alt: string
  contentAddress?: string
}

export type LrcMessagePub = boolean

export type AspectRatio = {
  width: number
  height: number
}

export function isEnby(item: Item): item is Enby {
  return item.type === "enby"
}

export function isMessage(item: Item): item is Message {
  return item.type === 'message' || item.type === 'enby'
}

export function isImage(item: Item): item is Image {
  return item.type === 'image' || item.type === 'enby'
}

export function isMedia(item: Item): item is Media {
  return isImage(item)
}

export type LogItem = {
  id: number
  binary: string
  time: number
  type: string
  key: number
}

