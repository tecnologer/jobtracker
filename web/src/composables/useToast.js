import { ref } from 'vue'

// single shared error toast; new errors replace the current one
const toast = ref(null) // string | null
let timer

export function useToast() {
  function showError(message) {
    toast.value = message
    clearTimeout(timer)
    timer = setTimeout(() => { toast.value = null }, 6000)
  }

  function dismiss() {
    toast.value = null
    clearTimeout(timer)
  }

  return { toast, showError, dismiss }
}
