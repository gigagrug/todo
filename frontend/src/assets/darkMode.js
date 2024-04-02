const isDarkMode = localStorage.getItem("darkMode")
const htmlTag = document.documentElement
const toggleSwitch = document.querySelector("#darkModeToggle")
if (isDarkMode === "true") {
  htmlTag.classList.add("dark")
  if (toggleSwitch) {
    toggleSwitch.checked = true
  }
}
if (toggleSwitch) {
  toggleSwitch.addEventListener("change", function () {
    if (this.checked) {
      htmlTag.classList.add("dark")
      localStorage.setItem("darkMode", "true")
    } else {
      htmlTag.classList.remove("dark")
      localStorage.setItem("darkMode", "false")
    }
  })
}
