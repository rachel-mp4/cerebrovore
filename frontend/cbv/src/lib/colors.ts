export type ColorSet = {
  theme: string
  themetransparent: string
  themecontrast: string
  themecontrasttransparent: string
}

export function numIsDark(num: number): boolean {
  const color = numToHex(num)
  return hexIsDark(color)
}

export function numToHex(num: number) {
  const int = Math.max(Math.min(16777215, Math.floor(num)), 0)
  return "#" + int.toString(16).padStart(6, '0')
}

export function hexToNum(hex: string) {
  return Number("0x" + hex.slice(1))
}
export function hexIsDark(hex: string): boolean {
  const r = Number("0x" + hex.slice(1, 3))
  const g = Number("0x" + hex.slice(3, 5))
  const b = Number("0x" + hex.slice(5))
  const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255
  return luminance > 0.5
}
export function hexToTransparent(hex: string) {
  return hex + "80"
}

// Function to convert RGB to Hex
function rgbToHex([r, g, b]: [number, number, number]) {
  return `#${((1 << 24) | (r << 16) | (g << 8) | b).toString(16).slice(1).toUpperCase()}`;
}
