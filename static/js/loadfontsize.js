(() => {
  let fsmin = localStorage.getItem("fs-min")
  let fspmax = localStorage.getItem("fs-pmax")
  if (fsmin === null) {
    fsmin = "1rem"
    localStorage.setItem("fs-min", "1rem")
  }
  if (fspmax === null) {
    fspmax = "1rem"
    localStorage.setItem("fs-pmax", "1rem")
  }
  const eub = document.getElementById("eats-ur-brain")
  if (eub !== null) {
    eub.style.setProperty("--fs-min", fsmin)
    eub.style.setProperty("--fs-pmax", fspmax)
  } else {
    const bue = document.getElementById("brains-ur-eat")
    if (bue !== null) {
      bue.style.setProperty("--fs-min", fsmin)
      bue.style.setProperty("--fs-pmax", fspmax)
    }
  }
})()
