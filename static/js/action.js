(() => {
  const trycloseactionmodal = (e) => {
    const rm = document.getElementById("action-modal")
    if (rm) {
      if (e.target.closest("button.close")) {
        rm.remove()
        document.querySelector(".actioning").classList.remove("actioning")
      }
      if (!e.target.closest("#action-modal-content")) {
        rm.remove()
        document.querySelector(".actioning").classList.remove("actioning")
      }
    }
  }
  const tryremoveimageblur = (e) => {
    const rd = e.target.closest(".reported")
    if (rd) {
      rd.classList.remove("reported")
      return true
    }
    return false
  }
  const actionify = (e, classname) => {
    const rb = e.target.closest("button.report")
    const db = e.target.closest("button.delete")
    if (rb || db) {
      const b = rb ?? db
      const tx = b.closest(classname)
      tx.classList.add("actioning")
      const id = tx.id
      var handle = tx.querySelector(".handle")?.textContent.trim() ?? ""
      if (handle.startsWith("@")) {
        handle = handle.slice(1)
      }
      if (rb) {
        const iws = tx.querySelectorAll(".image-wrapper")
        iws.forEach((iw) => {
          iw.classList.add("reported")
        })
        const prefix = window.location.origin + "/t/"
        const pl = prefix.length
        const href = window.location.href
        var tid
        if (href.startsWith(prefix)) {
          tid = href.slice(pl)
        }
        document.body.insertAdjacentHTML('beforeend', `<div id="action-modal"><div id="action-modal-content"><h2>reporting post #${id}</h2><form hx-post="/report" hx-swap="outerHTML"><input type="hidden" name="postid" value="${id}" />${tid !== "" ? '<input type="hidden" name="threadid" value="' + tid + '" />' : ''}<input type="hidden" name="username" value="${handle}" /><div class="form-field"><label for="reason">reason</label><input type="text" name="reason"/></div><div class="confirm-cancel"><div class="form-submit"><input type="submit" value="report"/></div><button class="close">close</button></div></form></div></div>`)
      } else {
        document.body.insertAdjacentHTML('beforeend', `<div id="action-modal"><div id="action-modal-content"><h2>deleting post #${id}</h2><div class="confirm-cancel"><button hx-delete="/p/${id}" hx-target="#action-modal-content">delete</button><button class="close">close</button></div></div></div>`)
      }
      htmx.process(document.getElementById("action-modal"))
    }
  }
  const pn = window.location.pathname
  if (pn.startsWith("/t/") || pn.startsWith("/inbox")) {
    document.addEventListener("click", (e) => {
      trycloseactionmodal(e)
      if (tryremoveimageblur(e)) {
        return
      }
      actionify(e, ".tx")
    })
  } else if (pn.startsWith("/ft/")) {
    document.addEventListener("click", (e) => {
      trycloseactionmodal(e)
      if (tryremoveimageblur(e)) {
        return
      }
      actionify(e, ".forum-post")
    })
  }
})()

