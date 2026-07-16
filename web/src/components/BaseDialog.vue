<template>
  <div
    :class="[z, sheet ? 'p-0 md:p-4 items-stretch md:items-center' : 'p-4 items-center']"
    class="fixed inset-0 bg-black/50 flex justify-center"
    @keydown="trapTab"
  >
    <div
      ref="panel"
      role="dialog"
      aria-modal="true"
      tabindex="-1"
      :class="[width, sheet ? 'rounded-none md:rounded-xl' : 'rounded-xl']"
      class="bg-white dark:bg-gray-800 shadow-xl p-6 focus:outline-none"
    >
      <slot />
    </div>
  </div>
</template>

<script>
// module-level stack of open dialogs: Esc closes only the most recently opened one
// (last-opened-wins, matching the old onEsc priority chain)
const escStack = []

window.addEventListener('keydown', e => {
  if (e.key === 'Escape' && escStack.length) escStack[escStack.length - 1]()
})

// body scroll lock is shared across nested dialogs: unlock only when the last closes
let openCount = 0

const FOCUSABLE = 'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'
</script>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'

defineProps({
  width: { type: String, default: 'w-full max-w-md' },
  z: { type: String, default: 'z-50' },
  sheet: { type: Boolean, default: false },
})

const emit = defineEmits(['close'])

const close = () => emit('close')

const panel = ref(null)
let opener // element focused before the dialog opened, restored on close

// keep Tab inside the dialog, wrapping at both ends
function trapTab(e) {
  if (e.key !== 'Tab' || e.defaultPrevented) return
  const nodes = panel.value.querySelectorAll(FOCUSABLE)
  if (!nodes.length) { e.preventDefault(); return }
  const first = nodes[0]
  const last = nodes[nodes.length - 1]
  const active = document.activeElement
  if (e.shiftKey && (active === first || active === panel.value)) {
    e.preventDefault()
    last.focus()
  } else if (!e.shiftKey && active === last) {
    e.preventDefault()
    first.focus()
  }
}

onMounted(() => {
  escStack.push(close)
  opener = document.activeElement
  panel.value.focus()
  if (++openCount === 1) document.body.style.overflow = 'hidden'
})

onUnmounted(() => {
  const i = escStack.indexOf(close)
  if (i >= 0) escStack.splice(i, 1)
  if (--openCount === 0) document.body.style.overflow = ''
  if (opener?.isConnected) opener.focus()
})
</script>
