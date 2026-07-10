import { ref } from 'vue'

const dark = ref(false)

function applyDark(val) {
  dark.value = val
  document.documentElement.classList.toggle('dark', val)
  localStorage.setItem('theme', val ? 'dark' : 'light')
}

export function useDarkMode() {
  function toggleDark() {
    applyDark(!dark.value)
  }

  function initDarkMode() {
    const saved = localStorage.getItem('theme')
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
    applyDark(saved ? saved === 'dark' : prefersDark)
  }

  return { dark, toggleDark, initDarkMode }
}
