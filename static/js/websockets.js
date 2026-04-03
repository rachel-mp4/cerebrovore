const baseurl = window.location.hostname
const port = window.location.port
const proto = window.location.protocol
const withproto = `ws${proto.slice(4)}${baseurl}${(port !== "") ? ":" + port : ""}/ws?watcher=andNewThreads`
const path = window.location.pathname
const match = path.match(/^\/t\/([0-9a-z]+)/)
if (match !== null) {
  const ntid = match[1]
  wsurl = `${withproto}&wormwatch=${ntid}&thread=${ntid}`
} else {
  wsurl = withproto
}
const ws = new WebSocket(wsurl)
ws.onopen = () => { console.log("hello ws") }
ws.onmessage = (e) => {
  const jed = JSON.parse(e.data)
  const ev = new CustomEvent(`cbv:${jed.type}`, { detail: jed.data })
  document.dispatchEvent(ev)
}
ws.onerror = (e) => {
  console.log("wserror,", e)
}
ws.onclose = (e) => {
  console.log("wsclose,", e)
}
