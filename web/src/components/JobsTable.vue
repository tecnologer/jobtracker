<template>
  <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden">
    <table class="w-full table-fixed text-sm">
      <thead class="bg-gray-50 dark:bg-gray-700 border-b border-gray-200 dark:border-gray-600">
        <tr>
          <th class="w-[12%] text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">
            Company
          </th>
          <th class="w-[20%] text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">
            Position
          </th>
          <th class="w-[9%] text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">
            Status
          </th>
          <th class="w-[12%] text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">
            Stage
          </th>
          <th class="w-[14%] text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">
            Applied
          </th>
          <th class="w-[22%] text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">
            Notes
          </th>
          <th class="w-[12%] px-2 py-3" />
        </tr>
      </thead>
      <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
        <tr class="bg-blue-50/40 dark:bg-blue-900/10 border-b-2 border-blue-100 dark:border-blue-800">
          <td class="px-4 py-2">
            <input
              ref="companyInput"
              v-model="form.company"
              placeholder="Company"
              class="w-full border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-300 dark:placeholder-gray-600 focus:outline-none focus:ring-1 focus:ring-blue-500"
              @keydown.enter.prevent="save"
            >
          </td>
          <td class="px-4 py-2">
            <input
              v-model="form.position"
              placeholder="Position"
              class="w-full border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-300 dark:placeholder-gray-600 focus:outline-none focus:ring-1 focus:ring-blue-500"
              @keydown.enter.prevent="save"
            >
          </td>
          <td class="px-4 py-2">
            <select
              v-model="form.status"
              class="w-full min-w-0 border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-1 focus:ring-blue-500"
              @keydown.enter.prevent="save"
            >
              <option
                v-for="s in statuses"
                :key="s"
                :value="s"
              >
                {{ s.replace('_', ' ') }}
              </option>
            </select>
          </td>
          <td class="px-4 py-2" />
          <td class="px-4 py-2">
            <input
              v-model="form.applied_at"
              type="date"
              class="w-full min-w-0 border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-1 focus:ring-blue-500"
              @keydown.enter.prevent="save"
            >
          </td>
          <td class="px-4 py-2">
            <input
              v-model="form.notes"
              placeholder="Notes"
              class="w-full border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-300 dark:placeholder-gray-600 focus:outline-none focus:ring-1 focus:ring-blue-500"
              @keydown.enter.prevent="save"
            >
          </td>
          <td class="px-4 py-2 text-right pr-3">
            <button
              class="px-3 py-1 text-sm bg-blue-600 hover:bg-blue-700 text-white rounded"
              @click="save"
            >
              Add
            </button>
          </td>
        </tr>
        <tr v-if="jobs.length === 0">
          <td
            colspan="7"
            class="text-center px-4 py-10 text-gray-400 dark:text-gray-500"
          >
            {{ totalCount === 0 ? 'No applications yet.' : 'No results match your filters.' }}
          </td>
        </tr>
        <tr
          v-for="job in jobs"
          :key="job.id"
          class="hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
          @dblclick="onRowDblClick($event, job)"
        >
          <td class="px-4 py-3 font-medium text-gray-800 dark:text-gray-100">
            <div class="flex items-center gap-1.5">
              <button
                :title="job.top_match ? 'Remove top match' : 'Mark as top match'"
                :aria-label="job.top_match ? 'Remove top match' : 'Mark as top match'"
                class="shrink-0"
                @click.stop="toggleTopMatch(job)"
              >
                <svg
                  :class="job.top_match ? 'text-amber-500 fill-current' : 'text-gray-400 dark:text-gray-500 fill-none stroke-current'"
                  class="w-4 h-4"
                  viewBox="0 0 20 20"
                  stroke-width="1.5"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"
                  />
                </svg>
              </button>
              <span v-html="highlight(filterText, job.company)" />
            </div>
          </td>
          <td class="px-4 py-3 text-gray-700 dark:text-gray-300">
            <a
              v-if="job.url"
              :href="job.url"
              target="_blank"
              rel="noopener"
              class="text-blue-600 dark:text-blue-400 hover:underline"
              v-html="highlight(filterText, job.position)"
            />
            <span
              v-else
              v-html="highlight(filterText, job.position)"
            />
          </td>
          <td class="px-4 py-3">
            <span
              :class="statusClass(job.status)"
              class="inline-block px-2 py-0.5 rounded-full text-xs font-semibold"
            >
              {{ job.status.replace('_', ' ') }}
            </span>
          </td>
          <td class="px-4 py-3">
            <div class="flex flex-col gap-1 min-w-32 max-w-40">
              <div
                class="h-1.5 w-full bg-gray-200 dark:bg-gray-600 rounded-full overflow-hidden"
                :title="job.stage?.name"
              >
                <div
                  v-if="job.stage_id"
                  class="h-full bg-purple-500 rounded-full transition-all"
                  :style="`width: ${stageProgress(job)}%`"
                />
              </div>
              <span class="text-xs text-gray-400 dark:text-gray-500">{{ job.stage?.name }}</span>
            </div>
          </td>
          <td class="px-4 py-3 text-gray-500 dark:text-gray-400">
            {{ formatDay(job.applied_at) }}
          </td>
          <td
            class="px-4 py-3 text-gray-500 dark:text-gray-400"
            :title="job.notes"
          >
            <div class="truncate">
              {{ truncateNotes(job.notes) }}
            </div>
          </td>
          <td class="px-2 py-3">
            <div class="flex flex-nowrap items-center gap-1 justify-end whitespace-nowrap">
              <button
                aria-label="View details"
                title="View details"
                class="inline-flex items-center justify-center p-2 rounded text-green-600 hover:text-green-800 hover:bg-green-50 dark:hover:bg-green-900 focus:outline-none focus-visible:ring-2 focus-visible:ring-green-500 transition-colors"
                @click="emit('view', job)"
              >
                <svg
                  class="w-4 h-4"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="1.5"
                  aria-hidden="true"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z"
                  />
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                  />
                </svg>
              </button>
              <button
                v-if="job.archived_at"
                aria-label="Unarchive"
                title="Unarchive"
                class="inline-flex items-center justify-center p-2 rounded text-amber-600 hover:text-amber-800 hover:bg-amber-50 dark:hover:bg-amber-900 focus:outline-none focus-visible:ring-2 focus-visible:ring-amber-500 transition-colors"
                @click="setArchived(job, null)"
              >
                <svg
                  class="w-4 h-4"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="1.5"
                  aria-hidden="true"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5M16.5 7.5L12 3m0 0L7.5 7.5M12 3v13.5"
                  />
                </svg>
              </button>
              <button
                v-else
                aria-label="Archive"
                title="Archive"
                class="inline-flex items-center justify-center p-2 rounded text-amber-600 hover:text-amber-800 hover:bg-amber-50 dark:hover:bg-amber-900 focus:outline-none focus-visible:ring-2 focus-visible:ring-amber-500 transition-colors"
                @click="confirmArchive = { open: true, job }"
              >
                <svg
                  class="w-4 h-4"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="1.5"
                  aria-hidden="true"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M20.25 7.5l-.625 10.632a2.25 2.25 0 01-2.247 2.118H6.622a2.25 2.25 0 01-2.247-2.118L3.75 7.5M10 11.25h4M3.375 7.5h17.25c.621 0 1.125-.504 1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125z"
                  />
                </svg>
              </button>
              <button
                aria-label="Delete"
                title="Delete"
                class="inline-flex items-center justify-center p-2 rounded text-red-500 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-900 focus:outline-none focus-visible:ring-2 focus-visible:ring-red-500 transition-colors"
                @click="confirmDelete = { open: true, id: job.id }"
              >
                <svg
                  class="w-4 h-4"
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
            </div>
          </td>
        </tr>
      </tbody>
    </table>

    <ConfirmDialog
      v-if="confirmDelete.open"
      title="Delete job application?"
      message="This action cannot be undone."
      confirm-label="Delete"
      tone="red"
      @confirm="doDelete"
      @close="confirmDelete.open = false"
    />
    <ConfirmDialog
      v-if="confirmArchive.open"
      title="Archive job application?"
      message="This marks the job as archived."
      confirm-label="Archive"
      tone="amber"
      @confirm="doArchive"
      @close="confirmArchive.open = false"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import * as api from '../api'
import { statuses, statusClass } from '../constants'
import { todayLocal, formatDay } from '../utils/dates'
import { highlight, truncateNotes } from '../utils/text'
import { useJobs } from '../composables/useJobs'
import ConfirmDialog from './ConfirmDialog.vue'

defineProps({
  jobs: { type: Array, required: true },
  filterText: { type: String, default: '' },
  totalCount: { type: Number, required: true },
})

const emit = defineEmits(['view'])

const { loadJobs, removeJob, setArchived, toggleTopMatch } = useJobs()

function emptyForm() {
  return { company: '', position: '', status: 'applied', applied_at: todayLocal(), notes: '', url: '' }
}

const form = ref(emptyForm())
const companyInput = ref(null)
const confirmDelete = ref({ open: false, id: null })
const confirmArchive = ref({ open: false, job: null })

async function save() {
  if (!form.value.company || !form.value.position) return
  await api.createJob(form.value)
  form.value = emptyForm()
  await loadJobs()
  companyInput.value?.focus()
}

async function doDelete() {
  await removeJob(confirmDelete.value.id)
  confirmDelete.value = { open: false, id: null }
}

async function doArchive() {
  await setArchived(confirmArchive.value.job, new Date().toISOString())
  confirmArchive.value = { open: false, job: null }
}

function stageProgress(job) {
  const list = job.stages ?? []
  if (!job.stage_id || list.length === 0) return 0
  const idx = list.findIndex(s => s.id === job.stage_id)
  return idx < 0 ? 0 : Math.round((idx + 1) / list.length * 100)
}

function onRowDblClick(event, job) {
  // ignore double-clicks on the row's own controls (star, URL, action icons)
  if (event.target.closest('button, a')) return
  emit('view', job)
}

onMounted(() => companyInput.value?.focus())
</script>
