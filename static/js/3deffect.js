// this is hell to get to work with scrolling, but it looks so good when we don't care.
// so we're leaving it in the code, but we're not serving it. maybe at some point we can
// revisit it
(() => {
  var curx = window.innerWidth / 2
  var cury = window.innerHeight / 2
  var targetx = window.innerWidth / 2
  var targety = window.innerHeight / 2
  const update = () => {
    curx = curx * 0.7 + (targetx) * 0.3
    cury = cury * 0.7 + (targety) * 0.3
    const relx = (curx - window.innerWidth / 2) / window.innerWidth
    const rely = (cury - window.innerHeight / 2) / window.innerWidth
    const atan = Math.atan2(rely, relx) + Math.PI / 2 + .1
    const corx = Math.cos(atan)
    const cory = Math.sin(atan)
    requestAnimationFrame(() => {
      document.body.style.transform = `rotate3d(${corx},${cory},0,${Math.sqrt(relx * relx + rely * rely) / 16}rad)`
      update()
    })
  }
  update()
  document.addEventListener("mousemove", (e) => {
    targetx = e.pageX
    targety = e.pageY
  })
})()
