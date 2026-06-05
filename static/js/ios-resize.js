if (window.visualViewport) {
  window.visualViewport.addEventListener("resize", () => {
    const tx = document.querySelector(".transmitter")
    tx.style.bottom = (window.innerHeight - visualViewport.height)
  })
}
