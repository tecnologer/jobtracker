<template>
  <div
    :class="z"
    class="fixed inset-0 bg-black/50 flex items-center justify-center p-4"
  >
    <div
      :class="width"
      class="bg-white dark:bg-gray-800 rounded-xl shadow-xl p-6"
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
</script>

<script setup>
import { onMounted, onUnmounted } from 'vue'

defineProps({
  width: { type: String, default: 'w-full max-w-md' },
  z: { type: String, default: 'z-50' },
})

const emit = defineEmits(['close'])

const close = () => emit('close')

onMounted(() => escStack.push(close))
onUnmounted(() => {
  const i = escStack.indexOf(close)
  if (i >= 0) escStack.splice(i, 1)
})
</script>
