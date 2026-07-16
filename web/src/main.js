import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import { useToast } from './composables/useToast'
import { useI18n } from './composables/useI18n'

const { showError } = useToast()
const { t } = useI18n()

// global net for failed requests: api.js throws on any non-2xx, and errors from
// event handlers / lifecycle hooks land here, so no save can fail silently
function onError(err) {
  console.error(err)
  const detail = err?.message ? ` (${err.message})` : ''
  showError(t('common.requestFailed') + detail)
}

const app = createApp(App)
app.config.errorHandler = onError
window.addEventListener('unhandledrejection', e => onError(e.reason))
app.mount('#app')
