document.addEventListener("click", (e) => {
  const rm = document.getElementById("report-modal")
  if (rm) {
    if (e.target.closest("button.close")) {
      rm.remove()
      document.querySelector(".reporting").classList.remove("reporting")
    }
    if (!e.target.closest("#report-modal-content")) {
      rm.remove()
      document.querySelector(".reporting").classList.remove("reporting")
    }
  }
  const rb = e.target.closest("button.report")
  if (!rb) {
    const rd = e.target.closest(".reported")
    if (rd) {
      rd.classList.remove("reported")
    }
    return
  }
  const tx = rb.closest(".tx")
  tx.classList.add("reporting")
  const id = tx.id
  const handle = tx.querySelector(".handle").textContent.trim()
  const iw = tx.querySelector(".image-wrapper")
  if (iw) {
    iw.classList.add("reported")
  }
  document.body.insertAdjacentHTML('beforeend', `<div id="report-modal"><div id="report-modal-content"><h2>reporting post #${id}</h2><form hx-post="/report" hx-swap="outerHTML"><input type="hidden" name="postid" value="${id}" /><input type="hidden" name="username" value="${handle}" /><div class="form-field"><label for="reason">reason</label><input type="text" name="reason"/></div><div class="confirm-cancel"><div class="form-submit"><input type="submit" value="report"/></div><button class="close">close</button></div></form></div></div>`)
  htmx.process(document.getElementById("report-modal"))
})
