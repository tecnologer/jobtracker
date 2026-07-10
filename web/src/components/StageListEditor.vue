<template>
  <div>
    <ul class="space-y-2 mb-4">
      <li
        v-for="(stage, idx) in stages"
        :key="stage.id"
        draggable="true"
        :class="dragOverIdx === idx ? 'ring-2 ring-blue-400 rounded' : ''"
        class="flex items-center gap-2 cursor-grab"
        @dragstart="dragIdx = idx"
        @dragover.prevent="dragOverIdx = idx"
        @dragleave="dragOverIdx = null"
        @drop.prevent="drop(idx)"
      >
        <span class="text-gray-300 dark:text-gray-600 select-none text-sm">⠿</span>
        <input
          :value="stage.name"
          class="flex-1 border border-gray-300 dark:border-gray-600 rounded px-2 py-1.5 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-1 focus:ring-blue-500"
          @blur="emit('rename', stage, $event.target.value)"
        >
        <button
          aria-label="Delete stage"
          title="Delete stage"
          class="inline-flex items-center justify-center p-1.5 rounded text-red-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900 focus:outline-none focus-visible:ring-2 focus-visible:ring-red-500 transition-colors"
          @click="emit('remove', stage.id)"
        >
          <svg
            class="w-3.5 h-3.5"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
            aria-hidden="true"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0"
            />
          </svg>
        </button>
      </li>
    </ul>
    <form
      class="flex gap-2"
      @submit.prevent="add"
    >
      <input
        v-model="newName"
        placeholder="New stage name"
        required
        class="flex-1 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
      >
      <button
        type="submit"
        class="bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium px-3 py-2 rounded-lg transition-colors"
      >
        Add
      </button>
    </form>
  </div>
</template>

<script setup>
import { ref } from 'vue'

defineProps({
  stages: { type: Array, required: true },
})

const emit = defineEmits(['add', 'rename', 'remove', 'reorder'])

const newName = ref('')
const dragIdx = ref(null)
const dragOverIdx = ref(null)

function drop(toIdx) {
  const fromIdx = dragIdx.value
  dragIdx.value = null
  dragOverIdx.value = null
  if (fromIdx === null || fromIdx === toIdx) return
  emit('reorder', fromIdx, toIdx)
}

function add() {
  emit('add', newName.value)
  newName.value = ''
}
</script>
