<template>
  <BaseDialog
    width="w-full max-w-sm"
    z="z-[60]"
    @close="emit('close')"
  >
    <h3 class="font-semibold text-gray-800 dark:text-gray-100 mb-1">
      Stage changed
    </h3>
    <p class="text-xs text-gray-500 dark:text-gray-400 mb-3">
      Add an optional comment for this transition.
    </p>
    <textarea
      v-model="notes"
      rows="3"
      placeholder="How did it go? (optional)"
      class="w-full border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none mb-3"
    />
    <div
      v-if="isLastStage"
      class="flex flex-col gap-1 mb-4"
    >
      <label class="text-xs font-medium text-gray-600 dark:text-gray-400">Set status</label>
      <select
        v-model="newStatus"
        class="border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
      >
        <option
          v-for="s in statuses"
          :key="s"
          :value="s"
        >
          {{ s.replace('_', ' ') }}
        </option>
      </select>
    </div>
    <div class="flex gap-2 justify-end">
      <button
        class="bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200 text-sm font-medium px-4 py-2 rounded-lg transition-colors"
        @click="emit('close')"
      >
        Cancel
      </button>
      <button
        class="bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors"
        @click="emit('confirm', { notes, newStatus })"
      >
        Save
      </button>
    </div>
  </BaseDialog>
</template>

<script setup>
import { ref } from 'vue'
import { statuses } from '../constants'
import BaseDialog from './BaseDialog.vue'

const props = defineProps({
  isLastStage: { type: Boolean, default: false },
  initialStatus: { type: String, default: '' },
})

const emit = defineEmits(['confirm', 'close'])

const notes = ref('')
const newStatus = ref(props.initialStatus)
</script>
