const cbvWSBuffer = {};
(() => {
  const wantsNewThreads = localStorage.getItem("new-threads")
  const baseurl = window.location.hostname
  const port = window.location.port
  const proto = window.location.protocol
  const withproto = `ws${proto.slice(4)}${baseurl}${(port !== "") ? ":" + port : ""}/ws?watcher=${wantsNewThreads === "" ? "yes" : "and-new-threads"}`
  const path = window.location.pathname
  const match = path.match(/^\/t\/([0-9a-z]+)/)
  if (match !== null) {
    const ntid = match[1]
    wsurl = `${withproto}&wormwatch=${ntid}&thread=${ntid}`
  } else {
    const match2 = path.match(/^\/ft\/([0-9a-z]+)/)
    if (match2 !== null) {
      const ntid = match2[1]
      wsurl = `${withproto}&wormwatch=${ntid}&thread=${ntid}`
    } else {
      wsurl = withproto
    }
  }
  const ws = new WebSocket(wsurl)
  const startT = Date.now()
  ws.onopen = () => { console.log("hello ws") }
  ws.onmessage = (e) => {
    const jed = JSON.parse(e.data)
    const etype = jed.type
    const data = jed.data
    const ev = new CustomEvent(`cbv:${etype}`, { detail: data })
    document.dispatchEvent(ev)
    if (etype === "notification") {
      const ic = document.getElementById("inbox-counter")
      const newCount = data.clear ? 0 : Number(ic.getAttribute("data-count")) + (data.count ?? 1)
      ic.setAttribute("data-count", newCount)
      ic.textContent = newCount !== 0 ? `${newCount.toString(36)} wMAIL${newCount !== 1 ? "s" : ''
        } ` : ""
    }
    // only buffer messages for first 100 seconds
    if (Date.now() - 100000 > startT) {
      return
    }
    const buffer = cbvWSBuffer[etype]
    if (buffer === undefined) {
      cbvWSBuffer[etype] = [data]
    } else {
      console.log("pushing")
      buffer.push(data)
      console.log("pushed ", data)
      cbvWSBuffer[etype] = buffer
    }
  }
  ws.onerror = (e) => {
    console.log("wserror,", e)
  }
  ws.onclose = (e) => {
    console.log("wsclose,", e)
  }
  if (wantsNewThreads === null) {
    localStorage.setItem("new-threads", "yes")
  }
})()
