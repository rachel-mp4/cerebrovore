document.addEventListener("click", (e) => {
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
  const rd = e.target.closest(".reported")
  if (rd) {
    rd.classList.remove("reported")
    return
  }
  const rb = e.target.closest("button.report")
  const db = e.target.closest("button.delete")
  if (rb || db) {
    const b = rb ?? db
    const tx = b.closest(".tx")
    tx.classList.add("actioning")
    const id = tx.id
    const handle = tx.querySelector(".handle").textContent.trim()
    if (rb) {
      const iw = tx.querySelector(".image-wrapper")
      if (iw) {
        iw.classList.add("reported")
      }
      document.body.insertAdjacentHTML('beforeend', `<div id="action-modal"><div id="action-modal-content"><h2>reporting post #${id}</h2><form hx-post="/report" hx-swap="outerHTML"><input type="hidden" name="postid" value="${id}" /><input type="hidden" name="username" value="${handle}" /><div class="form-field"><label for="reason">reason</label><input type="text" name="reason"/></div><div class="confirm-cancel"><div class="form-submit"><input type="submit" value="report"/></div><button class="close">close</button></div></form></div></div>`)
    } else {
      document.body.insertAdjacentHTML('beforeend', `<div id="action-modal"><div id="action-modal-content"><h2>deleting post #${id}</h2><div class="confirm-cancel"><button hx-delete="/p/${id}" hx-target="#action-modal-content">delete</button><button class="close">close</button></div></div></div>`)
    }
    htmx.process(document.getElementById("action-modal"))

  }
})
