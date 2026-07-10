<template>
  <header class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 px-6 py-4 mb-6 flex items-center justify-between">
    <h1 class="text-2xl font-bold text-gray-800 dark:text-gray-100">
      Job Tracker
    </h1>
    <div class="flex gap-2">
      <button
        class="text-sm px-3 py-1.5 rounded-lg border border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
        @click="emit('toggle-view')"
      >
        {{ dashboardOpen ? 'Jobs' : 'Dashboard' }}
      </button>
      <a
        href="/api/jobs/export"
        download
        class="text-sm px-3 py-1.5 rounded-lg border border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
      >
        Export CSV
      </a>
      <div class="relative">
        <button
          class="text-sm px-3 py-1.5 rounded-lg border border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
          @click="upcomingMeetingsOpen = !upcomingMeetingsOpen"
        >
          Upcoming meetings
          <span
            v-if="upcomingMeetings.length"
            class="ml-1 inline-flex items-center justify-center text-xs bg-blue-600 text-white rounded-full h-4 min-w-4 px-1 leading-none align-middle"
          >
            {{ upcomingMeetings.length }}
          </span>
        </button>
        <div
          v-if="upcomingMeetingsOpen"
          class="absolute z-20 top-full mt-1 right-0 w-80 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 rounded-lg shadow-lg py-1 max-h-96 overflow-y-auto"
        >
          <p
            v-if="!upcomingMeetings.length"
            class="px-3 py-2 text-xs text-gray-400 dark:text-gray-500"
          >
            No upcoming meetings
          </p>
          <button
            v-for="m in upcomingMeetings"
            :key="m.id"
            :class="isUrgent(m.scheduled_at) ? 'bg-amber-50 dark:bg-amber-900/30' : ''"
            class="w-full text-left px-3 py-2 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors border-b border-gray-100 dark:border-gray-700 last:border-b-0"
            @click="openMeeting(m)"
          >
            <div class="text-sm font-medium text-gray-800 dark:text-gray-100">
              {{ m.job?.company }} — {{ m.job?.position }}
            </div>
            <div class="text-xs text-gray-600 dark:text-gray-300">
              {{ m.title }}
            </div>
            <div class="text-xs text-gray-400 dark:text-gray-500">
              {{ formatDate(m.scheduled_at) }}
            </div>
          </button>
        </div>
        <div
          v-if="upcomingMeetingsOpen"
          class="fixed inset-0 z-10"
          @click="upcomingMeetingsOpen = false"
        />
      </div>
      <button
        class="text-sm px-3 py-1.5 rounded-lg border border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
        @click="defaultStagesMgmt = true"
      >
        Default Stages
      </button>
      <button
        class="text-sm px-3 py-1.5 rounded-lg border border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
        @click="toggleDark"
      >
        {{ dark ? '☀ Light' : '☾ Dark' }}
      </button>
    </div>

    <DefaultStagesDialog
      v-if="defaultStagesMgmt"
      @close="defaultStagesMgmt = false"
    />
  </header>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { formatDate, isUrgent } from '../utils/dates'
import { useMeetings } from '../composables/useMeetings'
import { useDarkMode } from '../composables/useDarkMode'
import DefaultStagesDialog from './DefaultStagesDialog.vue'

defineProps({
  dashboardOpen: { type: Boolean, default: false },
})

const emit = defineEmits(['open-job', 'toggle-view'])

const { upcomingMeetings } = useMeetings()
const { dark, toggleDark } = useDarkMode()

const upcomingMeetingsOpen = ref(false)
const defaultStagesMgmt = ref(false)

function openMeeting(m) {
  upcomingMeetingsOpen.value = false
  emit('open-job', m.job_id)
}

function onEsc(e) {
  if (e.key === 'Escape') upcomingMeetingsOpen.value = false
}

onMounted(() => window.addEventListener('keydown', onEsc))
onUnmounted(() => window.removeEventListener('keydown', onEsc))
</script>
