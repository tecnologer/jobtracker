<template>
  <BaseDialog
    width="w-full max-w-sm"
    @close="emit('close')"
  >
    <h3 class="font-semibold text-gray-800 dark:text-gray-100 mb-2">
      {{ title }}
    </h3>
    <p class="text-sm text-gray-500 dark:text-gray-400 mb-6">
      {{ message }}
    </p>
    <div class="flex justify-end gap-3">
      <button
        class="min-h-11 md:min-h-0 px-4 py-2 text-sm rounded-lg border border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
        @click="emit('close')"
      >
        {{ t('common.cancel') }}
      </button>
      <button
        :disabled="fired"
        :class="tone === 'amber' ? 'bg-amber-500 hover:bg-amber-600' : 'bg-red-500 hover:bg-red-600'"
        class="min-h-11 md:min-h-0 px-4 py-2 text-sm rounded-lg text-white font-medium transition-colors disabled:opacity-50"
        @click="confirm"
      >
        {{ confirmLabel }}
      </button>
    </div>
  </BaseDialog>
</template>

<script setup>
import { ref } from 'vue'
import BaseDialog from './BaseDialog.vue'
import { useI18n } from '../composables/useI18n'

const { t } = useI18n()

// one-shot: the destructive action must never fire twice from a double-click
const fired = ref(false)

function confirm() {
  if (fired.value) return
  fired.value = true
  emit('confirm')
}

defineProps({
  title: { type: String, required: true },
  message: { type: String, required: true },
  confirmLabel: { type: String, required: true },
  tone: { type: String, default: 'red' },
})

const emit = defineEmits(['confirm', 'close'])
</script>
