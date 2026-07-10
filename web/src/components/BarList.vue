<template>
  <div class="space-y-2">
    <div
      v-for="row in rows"
      :key="row.label"
      class="flex items-center gap-3"
    >
      <span class="w-36 shrink-0 truncate text-right text-sm capitalize text-gray-600 dark:text-gray-300">
        {{ row.label }}
      </span>
      <div class="flex-1 h-5 rounded bg-gray-100 dark:bg-gray-700 overflow-hidden">
        <div
          class="h-full rounded transition-all"
          :class="row.barClass ?? 'bg-blue-500'"
          :style="{ width: width(row) }"
        />
      </div>
      <span class="w-16 shrink-0 text-sm tabular-nums text-gray-800 dark:text-gray-100">
        {{ row.display ?? row.value }}
      </span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  // [{ label: String, value: Number, display: String?, barClass: String? }]
  rows: { type: Array, required: true },
})

const max = computed(() => Math.max(0, ...props.rows.map(r => r.value)))

function width(row) {
  // max === 0 guard: zero/empty data renders empty tracks, never NaN widths
  return max.value > 0 ? `${(row.value / max.value) * 100}%` : '0%'
}
</script>
