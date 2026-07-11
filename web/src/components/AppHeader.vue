<template>
  <header class="relative bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 px-3 md:px-6 py-3 md:py-4 mb-6 flex items-center justify-between gap-2">
    <div class="flex items-center gap-2 min-w-0 flex-1 md:flex-none">
      <!-- plain navigation to /: leaves the dashboard view and reloads when already home -->
      <a
        href="/"
        aria-label="Home"
        class="shrink-0"
      >
        <img
          src="/icon.svg"
          alt=""
          class="h-9 w-9 md:h-11 md:w-11 rounded-md"
        >
      </a>
      <h1 class="text-lg md:text-2xl font-bold text-gray-800 dark:text-gray-100 truncate min-w-0">
        Job Tracker
      </h1>
    </div>
    <div class="flex items-center gap-1.5 md:gap-2 shrink-0">
      <button
        :class="btn"
        :disabled="!jobs.length"
        :aria-label="dashboardOpen ? 'Jobs' : 'Dashboard'"
        @click="emit('toggle-view')"
      >
        <!-- briefcase (jobs) / chart-bar (dashboard) -->
        <svg
          v-if="dashboardOpen"
          class="h-5 w-5"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          stroke-width="1.5"
          stroke="currentColor"
          aria-hidden="true"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M20.25 14.15v4.25c0 1.094-.787 2.036-1.872 2.18-2.087.277-4.216.42-6.378.42s-4.291-.143-6.378-.42c-1.085-.144-1.872-1.086-1.872-2.18v-4.25m16.5 0a2.18 2.18 0 0 0 .75-1.661V8.706c0-1.081-.768-2.015-1.837-2.175a48.114 48.114 0 0 0-3.413-.387m4.5 8.006c-.194.165-.42.295-.673.38A23.978 23.978 0 0 1 12 15.75c-2.648 0-5.195-.429-7.577-1.22a2.016 2.016 0 0 1-.673-.38m0 0A2.18 2.18 0 0 1 3 12.489V8.706c0-1.081.768-2.015 1.837-2.175a48.111 48.111 0 0 1 3.413-.387m7.5 0V5.25A2.25 2.25 0 0 0 13.5 3h-3a2.25 2.25 0 0 0-2.25 2.25v.894m7.5 0a48.667 48.667 0 0 0-7.5 0M12 12.75h.008v.008H12v-.008Z"
          />
        </svg>
        <svg
          v-else
          class="h-5 w-5"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          stroke-width="1.5"
          stroke="currentColor"
          aria-hidden="true"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 0 1 3 19.875v-6.75ZM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 0 1-1.125-1.125V8.625ZM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 0 1-1.125-1.125V4.125Z"
          />
        </svg>
      </button>
      <a
        :class="[btn, !filteredJobs.length && 'opacity-50 pointer-events-none']"
        :href="filteredJobs.length ? exportHref : null"
        :aria-disabled="!filteredJobs.length || null"
        download
        aria-label="Export CSV"
      >
        <!-- arrow-down-tray -->
        <svg
          class="h-5 w-5"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          stroke-width="1.5"
          stroke="currentColor"
          aria-hidden="true"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3"
          />
        </svg>
      </a>
      <div class="relative">
        <button
          :class="btn"
          :disabled="!upcomingMeetings.length"
          aria-label="Upcoming meetings"
          @click="upcomingMeetingsOpen = !upcomingMeetingsOpen"
        >
          <!-- calendar-days -->
          <svg
            class="h-5 w-5"
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
            stroke-width="1.5"
            stroke="currentColor"
            aria-hidden="true"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5m-9-6h.008v.008H12v-.008ZM12 15h.008v.008H12V15Zm0 2.25h.008v.008H12v-.008ZM9.75 15h.008v.008H9.75V15Zm0 2.25h.008v.008H9.75v-.008ZM7.5 15h.008v.008H7.5V15Zm0 2.25h.008v.008H7.5v-.008Zm6.75-4.5h.008v.008h-.008v-.008Zm0 2.25h.008v.008h-.008V15Zm0 2.25h.008v.008h-.008v-.008Zm2.25-4.5h.008v.008H16.5v-.008Zm0 2.25h.008v.008H16.5V15Z"
            />
          </svg>
          <span
            v-if="upcomingMeetings.length"
            class="absolute -top-1 -right-1 inline-flex items-center justify-center text-xs bg-blue-600 text-white rounded-full h-4 min-w-4 px-1 leading-none"
          >
            {{ upcomingMeetings.length }}
          </span>
        </button>
        <div
          v-if="upcomingMeetingsOpen"
          class="absolute z-20 top-full mt-1 inset-x-4 md:inset-x-auto md:right-0 md:w-80 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 rounded-lg shadow-lg py-1 max-h-96 overflow-y-auto"
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
            class="min-h-11 md:min-h-0 w-full text-left px-3 py-2 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors border-b border-gray-100 dark:border-gray-700 last:border-b-0"
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
        :class="btn"
        aria-label="Default Stages"
        @click="defaultStagesMgmt = true"
      >
        <!-- queue-list -->
        <svg
          class="h-5 w-5"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          stroke-width="1.5"
          stroke="currentColor"
          aria-hidden="true"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M3.75 12h16.5m-16.5 3.75h16.5M3.75 19.5h16.5M5.625 4.5h12.75a1.875 1.875 0 0 1 0 3.75H5.625a1.875 1.875 0 0 1 0-3.75Z"
          />
        </svg>
      </button>
      <button
        :class="btn"
        :aria-label="dark ? 'Light mode' : 'Dark mode'"
        @click="toggleDark"
      >
        <!-- sun / moon: shows the mode you switch to -->
        <svg
          v-if="dark"
          class="h-5 w-5"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          stroke-width="1.5"
          stroke="currentColor"
          aria-hidden="true"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M12 3v2.25m6.364.386-1.591 1.591M21 12h-2.25m-.386 6.364-1.591-1.591M12 18.75V21m-4.773-4.227-1.591 1.591M5.25 12H3m4.227-4.773L5.636 5.636M15.75 12a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0Z"
          />
        </svg>
        <svg
          v-else
          class="h-5 w-5"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          stroke-width="1.5"
          stroke="currentColor"
          aria-hidden="true"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M21.752 15.002A9.72 9.72 0 0 1 18 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 0 0 3 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 0 0 9.002-5.998Z"
          />
        </svg>
      </button>
    </div>

    <DefaultStagesDialog
      v-if="defaultStagesMgmt"
      @close="defaultStagesMgmt = false"
    />
  </header>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { formatDate, isUrgent } from '../utils/dates'
import { useJobs } from '../composables/useJobs'
import { useJobFilters } from '../composables/useJobFilters'
import { useMeetings } from '../composables/useMeetings'
import { useDarkMode } from '../composables/useDarkMode'
import DefaultStagesDialog from './DefaultStagesDialog.vue'

defineProps({
  dashboardOpen: { type: Boolean, default: false },
})

const emit = defineEmits(['open-job', 'toggle-view'])

// shared icon-button style; .tip renders the CSS tooltip from aria-label
const btn = 'tip relative min-h-11 min-w-11 md:min-h-9 md:min-w-9 inline-flex items-center justify-center rounded-lg border border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors disabled:opacity-50 disabled:pointer-events-none'

const { jobs } = useJobs()
const { filteredJobs } = useJobFilters()
const { upcomingMeetings } = useMeetings()

// export exactly what the grid shows; filtering is client-side, so we hand the backend the visible IDs
const exportHref = computed(() =>
  `/api/jobs/export?ids=${filteredJobs.value.map(j => j.id).join(',')}`)
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

<style scoped>
/* CSS-only tooltip sourced from aria-label: hover/focus on desktop,
   press-and-hold (:active) on touch, where title never renders */
.tip {
  -webkit-touch-callout: none;
  user-select: none;
}

.tip::after {
  content: attr(aria-label);
  position: absolute;
  top: calc(100% + 0.375rem);
  right: 0;
  z-index: 30;
  padding: 0.25rem 0.5rem;
  border-radius: 0.375rem;
  background: rgb(17 24 39);
  color: #fff;
  font-size: 0.75rem;
  line-height: 1rem;
  white-space: nowrap;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.15s ease 0.1s;
}

.tip:hover::after,
.tip:focus-visible::after,
.tip:active::after {
  opacity: 1;
}
</style>
