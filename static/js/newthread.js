document.addEventListener("paste", (e) => {
  const items = e.clipboardData?.items;
  if (items === undefined) {
    return;
  }
  for (const item of items) {
    if (item.type.startsWith("image/")) {
      const blob = item.getAsFile();
      if (blob === null) {
        return;
      }
      e.preventDefault();
      const dt = new DataTransfer()
      dt.items.add(blob)
      const nimg = document.getElementById("nt-img")
      nimg.files = dt.files
    }
  }
})
