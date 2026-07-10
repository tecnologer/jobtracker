<template>
  <ul class="space-y-3 max-h-40 overflow-y-auto pr-1">
    <li
      v-for="log in logs"
      :key="log.id"
      class="flex gap-3 text-sm"
    >
      <div class="w-1.5 h-1.5 rounded-full bg-purple-400 mt-1.5 shrink-0" />
      <div>
        <div class="flex items-center gap-2">
          <span class="text-xs text-gray-400 dark:text-gray-500">{{ log.prev_stage?.name ?? 'No stage' }}</span>
          <span class="text-gray-300 dark:text-gray-600">→</span>
          <span class="font-medium text-gray-700 dark:text-gray-300">{{ log.stage?.name ?? 'No stage' }}</span>
          <span class="text-xs text-gray-400 dark:text-gray-500">{{ formatDate(log.created_at) }}</span>
        </div>
        <p
          v-if="log.notes"
          class="text-xs text-gray-500 dark:text-gray-400 mt-0.5"
        >
          {{ log.notes }}
        </p>
      </div>
    </li>
  </ul>
</template>

<script setup>
import { formatDate } from '../utils/dates'

defineProps({
  logs: { type: Array, required: true },
})
</script>
