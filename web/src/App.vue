<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900">
    <header class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 px-6 py-4 mb-6 flex items-center justify-between">
      <h1 class="text-2xl font-bold text-gray-800 dark:text-gray-100">Job Tracker</h1>
      <div class="flex gap-2">
        <a href="/api/jobs/export" download
          class="text-sm px-3 py-1.5 rounded-lg border border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
          Export CSV
        </a>
        <button @click="defaultStagesMgmt = true"
          class="text-sm px-3 py-1.5 rounded-lg border border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
          Default Stages
        </button>
        <button @click="toggleDark"
          class="text-sm px-3 py-1.5 rounded-lg border border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
          {{ dark ? '☀ Light' : '☾ Dark' }}
        </button>
      </div>
    </header>

    <main class="max-w-6xl mx-auto px-6">
      <!-- Filters -->
      <div class="mb-6 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl px-4 py-2">
        <div class="flex items-center gap-2">
          <button @click="filtersOpen = !filtersOpen"
            class="flex items-center gap-2 text-sm font-medium text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200 transition-colors select-none py-1 flex-1">
            <svg :class="filtersOpen ? 'rotate-90' : ''" class="w-4 h-4 shrink-0 transition-transform text-gray-500 dark:text-gray-400" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true"><path fill-rule="evenodd" d="M7.21 14.77a.75.75 0 01.02-1.06L11.168 10 7.23 6.29a.75.75 0 111.04-1.08l4.5 4.25a.75.75 0 010 1.08l-4.5 4.25a.75.75 0 01-1.06-.02z" clip-rule="evenodd"/></svg>
            Filters
            <span v-if="isFiltered" class="text-xs bg-blue-100 dark:bg-blue-900 text-blue-600 dark:text-blue-300 rounded-full px-1.5 py-0.5 font-medium leading-none">{{ activeFilterCount }}</span>
          </button>
          <button @click="toggleArchivedOnly"
            :class="archivedOnly ? 'bg-amber-100 dark:bg-amber-900 text-amber-600 dark:text-amber-300 ring-2 ring-current' : 'bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400'"
            class="text-xs px-2.5 py-1 rounded-full font-medium transition-colors shrink-0">
            Archived only
          </button>
          <button @click="toggleActiveOnly"
            :class="isActiveOnly ? 'bg-blue-100 dark:bg-blue-900 text-blue-600 dark:text-blue-300 ring-2 ring-current' : 'bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400'"
            class="text-xs px-2.5 py-1 rounded-full font-medium transition-colors shrink-0">
            Active only
          </button>
        </div>

        <div v-show="filtersOpen" class="flex flex-col gap-2 mt-2 pb-2">
          <div class="flex flex-col gap-1">
            <label class="text-xs font-medium text-gray-600 dark:text-gray-400">Company or Position</label>
            <input v-model="filter.text" placeholder="Company or position…"
              class="w-full border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-1.5 text-sm bg-white dark:bg-gray-800 text-gray-800 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500" />
          </div>

          <!-- Row 2: status chips -->
          <div class="flex items-center gap-1.5 flex-wrap select-none">
            <span class="text-xs text-gray-400 dark:text-gray-500">Status:</span>
            <button v-for="s in statuses" :key="s"
              @mousedown.prevent="chipMousedown(filter.statuses, s)"
              @mouseenter="chipMouseenter(filter.statuses, s)"
              @dragstart.prevent
              :class="filter.statuses.includes(s) ? statusClass(s) + ' ring-2 ring-offset-1 ring-current' : 'bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400'"
              class="px-2 py-0.5 rounded-full text-xs font-medium transition-colors cursor-pointer">
              {{ s.replace('_', ' ') }}
            </button>
          </div>

          <!-- Row 3: stage dropdown + applied date -->
          <div class="flex gap-3 items-center">
            <div class="flex items-center gap-1.5">
              <span class="text-xs text-gray-400 dark:text-gray-500">Stage:</span>
              <div class="relative">
                <button @click="stageDropdownOpen = !stageDropdownOpen"
                  class="flex items-center justify-between gap-1 w-44 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-1 text-xs bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">
                  Stage{{ filter.stages.length ? ` (${filter.stages.length})` : '' }}
                  <span class="text-gray-400">▾</span>
                </button>
                <div v-if="stageDropdownOpen" class="absolute z-20 top-full mt-1 min-w-36 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 rounded-lg shadow-lg py-1">
                  <div v-if="allFilterStages.length" class="px-2 pt-1.5 pb-1 relative">
                    <input v-model="stageSearch" placeholder="Filter stages…"
                      class="w-full border border-gray-200 dark:border-gray-600 rounded px-2 py-1 pr-6 text-xs bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-200 placeholder-gray-400 focus:outline-none focus:ring-1 focus:ring-purple-400" />
                    <button v-if="stageSearch" @click="stageSearch = ''"
                      class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 leading-none">✕</button>
                  </div>
                  <p v-if="!filteredDropdownStages.length" class="px-3 py-1.5 text-xs text-gray-400">{{ allFilterStages.length ? 'No match' : 'No stages configured' }}</p>
                  <label v-for="s in filteredDropdownStages" :key="s.id"
                    class="flex items-center gap-2 px-3 py-1.5 hover:bg-gray-50 dark:hover:bg-gray-700 cursor-pointer text-sm text-gray-700 dark:text-gray-200">
                    <input type="checkbox" :value="s.name" v-model="filter.stages" class="rounded accent-purple-500" />
                    <span v-html="highlight(stageSearch, s.name)"></span>
                  </label>
                </div>
                <div v-if="stageDropdownOpen" class="fixed inset-0 z-10" @click="stageDropdownOpen = false" />
              </div>
            </div>

            <div class="flex items-center gap-1.5 flex-1">
              <span class="text-xs text-gray-400 dark:text-gray-500 shrink-0">Applied:</span>
              <input v-model="filter.dateFrom" type="date"
                class="flex-1 min-w-0 border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-xs bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 focus:outline-none focus:ring-1 focus:ring-blue-500" />
              <span class="text-xs text-gray-400 shrink-0">–</span>
              <input v-model="filter.dateTo" type="date"
                class="flex-1 min-w-0 border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-xs bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 focus:outline-none focus:ring-1 focus:ring-blue-500" />
            </div>

            <button v-if="isFiltered" @click="clearFilter"
              class="text-xs text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 underline transition-colors">
              Clear
            </button>
          </div>
        </div>
      </div>

      <!-- Table -->
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden">
        <table class="w-full table-fixed text-sm">
          <thead class="bg-gray-50 dark:bg-gray-700 border-b border-gray-200 dark:border-gray-600">
            <tr>
              <th class="w-[12%] text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Company</th>
              <th class="w-[18%] text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Position</th>
              <th class="w-[9%] text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Status</th>
              <th class="w-[12%] text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Stage</th>
              <th class="w-[14%] text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Applied</th>
              <th class="w-[19%] text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Notes</th>
              <th class="w-[17%] px-4 py-3"></th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
            <tr class="bg-blue-50/40 dark:bg-blue-900/10 border-b-2 border-blue-100 dark:border-blue-800">
              <td class="px-4 py-2">
                <input ref="companyInput" v-model="form.company" placeholder="Company" @keydown.enter.prevent="save"
                  class="w-full border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-300 dark:placeholder-gray-600 focus:outline-none focus:ring-1 focus:ring-blue-500" />
              </td>
              <td class="px-4 py-2">
                <input v-model="form.position" placeholder="Position" @keydown.enter.prevent="save"
                  class="w-full border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-300 dark:placeholder-gray-600 focus:outline-none focus:ring-1 focus:ring-blue-500" />
              </td>
              <td class="px-4 py-2">
                <select v-model="form.status" @keydown.enter.prevent="save"
                  class="w-full min-w-0 border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-1 focus:ring-blue-500">
                  <option v-for="s in statuses" :key="s" :value="s">{{ s.replace('_', ' ') }}</option>
                </select>
              </td>
              <td class="px-4 py-2"></td>
              <td class="px-4 py-2">
                <input v-model="form.applied_at" type="date" @keydown.enter.prevent="save"
                  class="w-full min-w-0 border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-1 focus:ring-blue-500" />
              </td>
              <td class="px-4 py-2">
                <input v-model="form.notes" placeholder="Notes" @keydown.enter.prevent="save"
                  class="w-full border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-300 dark:placeholder-gray-600 focus:outline-none focus:ring-1 focus:ring-blue-500" />
              </td>
              <td class="px-4 py-2 text-right pr-3">
                <button @click="save"
                  class="px-3 py-1 text-sm bg-blue-600 hover:bg-blue-700 text-white rounded">Add</button>
              </td>
            </tr>
            <tr v-if="filteredJobs.length === 0">
              <td colspan="7" class="text-center px-4 py-10 text-gray-400 dark:text-gray-500">
                {{ jobs.length === 0 ? 'No applications yet.' : 'No results match your filters.' }}
              </td>
            </tr>
            <tr v-for="job in filteredJobs" :key="job.id" class="hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">
              <td class="px-4 py-3 font-medium text-gray-800 dark:text-gray-100" v-html="highlight(filter.text, job.company)"></td>
              <td class="px-4 py-3 text-gray-700 dark:text-gray-300">
                <a v-if="job.url" :href="job.url" target="_blank" rel="noopener"
                  class="text-blue-600 dark:text-blue-400 hover:underline" v-html="highlight(filter.text, job.position)"></a>
                <span v-else v-html="highlight(filter.text, job.position)"></span>
              </td>
              <td class="px-4 py-3">
                <span :class="statusClass(job.status)"
                  class="inline-block px-2 py-0.5 rounded-full text-xs font-semibold">
                  {{ job.status.replace('_', ' ') }}
                </span>
              </td>
              <td class="px-4 py-3">
                <div class="flex flex-col gap-1 min-w-32 max-w-40">
                  <div class="h-1.5 w-full bg-gray-200 dark:bg-gray-600 rounded-full overflow-hidden" :title="job.stage?.name">
                    <div v-if="job.stage_id" class="h-full bg-purple-500 rounded-full transition-all"
                      :style="`width: ${stageProgress(job)}%`"></div>
                  </div>
                  <span class="text-xs text-gray-400 dark:text-gray-500">{{ job.stage?.name }}</span>
                </div>
              </td>
              <td class="px-4 py-3 text-gray-500 dark:text-gray-400">{{ formatDay(job.applied_at) }}</td>
              <td class="px-4 py-3 text-gray-500 dark:text-gray-400" :title="job.notes">
                <div class="truncate">{{ truncateNotes(job.notes) }}</div>
              </td>
              <td class="px-4 py-3">
                <div class="flex flex-nowrap gap-2 justify-end whitespace-nowrap">
                  <button @click="openDetail(job)"
                    class="text-green-600 hover:text-green-800 text-xs font-medium px-2 py-1 rounded hover:bg-green-50 dark:hover:bg-green-900 transition-colors">
                    Detail
                  </button>
                  <button v-if="job.archived_at" @click="unarchive(job)"
                    class="text-amber-600 hover:text-amber-800 text-xs font-medium px-2 py-1 rounded hover:bg-amber-50 dark:hover:bg-amber-900 transition-colors">
                    Unarchive
                  </button>
                  <button v-else @click="archive(job)"
                    class="text-amber-600 hover:text-amber-800 text-xs font-medium px-2 py-1 rounded hover:bg-amber-50 dark:hover:bg-amber-900 transition-colors">
                    Archive
                  </button>
                  <button @click="remove(job.id)"
                    class="text-red-500 hover:text-red-700 text-xs font-medium px-2 py-1 rounded hover:bg-red-50 dark:hover:bg-red-900 transition-colors">
                    Delete
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </main>

    <!-- Stage Update Dialog -->
    <div v-if="stageDialog.open" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl p-6 w-full max-w-md">
        <div class="flex justify-between items-center mb-4">
          <h3 class="font-semibold text-gray-800 dark:text-gray-100">
            {{ stageDialog.job.company }} — {{ stageDialog.job.position }}
          </h3>
          <button @click="stageDialog.open = false"
            class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 text-lg leading-none">✕</button>
        </div>
        <div class="flex flex-col gap-3">
          <div class="flex flex-col gap-1">
            <label class="text-xs font-medium text-gray-600 dark:text-gray-400">Stage</label>
            <select v-model="stageDialog.stageId"
              class="border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500">
              <option v-for="s in stages" :key="s.id" :value="s.id">{{ s.name }}</option>
            </select>
          </div>
          <div class="flex flex-col gap-1">
            <label class="text-xs font-medium text-gray-600 dark:text-gray-400">Notes</label>
            <textarea v-model="stageDialog.notes" rows="3" placeholder="How did it go?"
              class="border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"></textarea>
          </div>
          <button @click="submitStageUpdate"
            class="bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors">
            Save
          </button>
        </div>
        <div v-if="stageDialog.logs.length" class="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700">
          <p class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-3">History</p>
          <ul class="space-y-3 max-h-48 overflow-y-auto pr-1">
            <li v-for="log in stageDialog.logs" :key="log.id" class="flex gap-3 text-sm">
              <div class="w-1.5 h-1.5 rounded-full bg-purple-400 mt-1.5 shrink-0"></div>
              <div>
                <div class="flex items-center gap-2">
                  <span class="text-xs text-gray-400 dark:text-gray-500">{{ log.prev_stage?.name ?? 'No stage' }}</span>
                  <span class="text-gray-300 dark:text-gray-600">→</span>
                  <span class="font-medium text-gray-700 dark:text-gray-300">{{ log.stage?.name ?? 'No stage' }}</span>
                  <span class="text-xs text-gray-400 dark:text-gray-500">{{ formatDate(log.created_at) }}</span>
                </div>
                <p v-if="log.notes" class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{{ log.notes }}</p>
              </div>
            </li>
          </ul>
        </div>
      </div>
    </div>

    <!-- Detail Dialog -->
    <div v-if="detailDialog.open" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl p-6 w-[62vw] max-h-[90vh] overflow-y-auto">
        <div class="flex justify-between items-start mb-4 gap-3">
          <div class="flex-1 flex flex-col gap-2">
            <div class="flex gap-2">
              <div class="flex flex-col gap-1 flex-1">
                <label class="text-xs font-medium text-gray-600 dark:text-gray-400">Company</label>
                <input v-model="detailDialog.edit.company" placeholder="Company"
                  class="border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-1.5 text-sm font-semibold bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
              <div class="flex flex-col gap-1 flex-1">
                <label class="text-xs font-medium text-gray-600 dark:text-gray-400">Position</label>
                <input v-model="detailDialog.edit.position" placeholder="Position"
                  class="border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-1.5 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
            </div>
            <div class="flex items-end gap-2">
              <div class="flex flex-col gap-1">
                <label class="text-xs font-medium text-gray-600 dark:text-gray-400">Status</label>
                <select v-model="detailDialog.edit.status"
                  class="border border-gray-300 dark:border-gray-600 rounded-lg px-2 py-1 text-xs bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500">
                  <option v-for="s in statuses" :key="s" :value="s">{{ s.replace('_', ' ') }}</option>
                </select>
              </div>
              <div class="flex flex-col gap-1">
                <label class="text-xs font-medium text-gray-600 dark:text-gray-400">Stage</label>
                <div class="flex items-center gap-1">
                  <select v-model="detailDialog.edit.stage_id"
                    class="border border-gray-300 dark:border-gray-600 rounded-lg px-2 py-1 text-xs bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500">
                    <option :value="0">No stage</option>
                    <option v-for="s in stages" :key="s.id" :value="s.id">{{ s.name }}</option>
                  </select>
                  <button @click="stagesMgmt = true" title="Manage stages"
                    class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors leading-none">
                    &#9881;
                  </button>
                </div>
              </div>
              <div class="flex flex-col gap-1">
                <label class="text-xs font-medium text-gray-600 dark:text-gray-400">Applied</label>
                <input v-model="detailDialog.edit.applied_at" type="date"
                  class="border border-gray-300 dark:border-gray-600 rounded-lg px-2 py-1 text-xs bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
            </div>
          </div>
          <button @click="detailDialog.open = false"
            class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 text-lg leading-none shrink-0">✕</button>
        </div>

        <div class="mb-4 pb-4 border-b border-gray-100 dark:border-gray-700 flex flex-col gap-2">
          <div class="flex flex-col gap-1">
            <label class="text-xs font-medium text-gray-600 dark:text-gray-400">Notes</label>
            <textarea v-model="detailDialog.edit.notes" rows="2" placeholder="Notes"
              class="border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"></textarea>
          </div>
          <div class="flex flex-col gap-1">
            <label class="text-xs font-medium text-gray-600 dark:text-gray-400">Job URL</label>
            <div class="flex gap-2 items-center">
              <input v-model="detailDialog.edit.url" type="url" placeholder="https://..."
                class="flex-1 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500" />
              <a :href="detailDialog.edit.url || '#'" target="_blank" rel="noopener"
                :class="detailDialog.edit.url ? 'text-blue-500 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 cursor-pointer' : 'text-blue-500 dark:text-blue-400 opacity-25 pointer-events-none'"
                :aria-disabled="!detailDialog.edit.url"
                title="Open job URL"
                class="text-lg leading-none transition-colors shrink-0">&#128279;</a>
            </div>
          </div>
        </div>

        <!-- Stage log -->
        <div class="mb-4 pb-4 border-b border-gray-100 dark:border-gray-700">
          <p class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-3">Stage History</p>
          <ul v-if="detailDialog.logs.length" class="space-y-3 max-h-40 overflow-y-auto pr-1">
            <li v-for="log in detailDialog.logs" :key="log.id" class="flex gap-3 text-sm">
              <div class="w-1.5 h-1.5 rounded-full bg-purple-400 mt-1.5 shrink-0"></div>
              <div>
                <div class="flex items-center gap-2">
                  <span class="text-xs text-gray-400 dark:text-gray-500">{{ log.prev_stage?.name ?? 'No stage' }}</span>
                  <span class="text-gray-300 dark:text-gray-600">→</span>
                  <span class="font-medium text-gray-700 dark:text-gray-300">{{ log.stage?.name ?? 'No stage' }}</span>
                  <span class="text-xs text-gray-400 dark:text-gray-500">{{ formatDate(log.created_at) }}</span>
                </div>
                <p v-if="log.notes" class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{{ log.notes }}</p>
              </div>
            </li>
          </ul>
          <p v-else class="text-sm text-gray-400 dark:text-gray-500">No stage history yet.</p>
        </div>

        <!-- Contacts -->
        <div>
          <p class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-3">Contacts</p>
          <ul v-if="detailDialog.contacts.length || pendingContacts.length" class="space-y-2 mb-4 max-h-40 overflow-y-auto pr-1">
            <li v-for="c in detailDialog.contacts" :key="c.id"
              class="flex items-start justify-between gap-2 text-sm border border-gray-100 dark:border-gray-700 rounded-lg px-3 py-2">
              <div>
                <span class="font-medium text-gray-800 dark:text-gray-100">{{ c.name }}</span>
                <span v-if="c.role" class="text-xs text-gray-500 dark:text-gray-400 ml-1">({{ c.role }})</span>
                <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5 space-x-2">
                  <span v-if="c.email">{{ c.email }}</span>
                  <span v-if="c.phone">{{ c.phone }}</span>
                </div>
              </div>
              <button @click="removeContact(c.id)"
                class="text-red-400 hover:text-red-600 text-xs shrink-0">✕</button>
            </li>
            <li v-for="(c, i) in pendingContacts" :key="'p'+i"
              class="flex items-start justify-between gap-2 text-sm border border-dashed border-blue-300 dark:border-blue-600 rounded-lg px-3 py-2 opacity-70">
              <div>
                <span class="font-medium text-gray-800 dark:text-gray-100">{{ c.name }}</span>
                <span v-if="c.role" class="text-xs text-gray-500 dark:text-gray-400 ml-1">({{ c.role }})</span>
                <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5 space-x-2">
                  <span v-if="c.email">{{ c.email }}</span>
                  <span v-if="c.phone">{{ c.phone }}</span>
                </div>
              </div>
              <button @click="pendingContacts.splice(i, 1)"
                class="text-red-400 hover:text-red-600 text-xs shrink-0">✕</button>
            </li>
          </ul>
          <p v-else class="text-sm text-gray-400 dark:text-gray-500 mb-4">No contacts yet.</p>
          <form @submit.prevent="addContact" class="flex flex-col gap-2">
            <div class="flex gap-2">
              <input v-model="newContact.name" placeholder="Name" required
                class="flex-1 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500" />
              <input v-model="newContact.role" placeholder="Role"
                class="flex-1 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <div class="flex gap-2">
              <input v-model="newContact.email" placeholder="Email" type="email"
                class="flex-1 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500" />
              <input v-model="newContact.phone" placeholder="Phone"
                class="flex-1 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
            <button type="submit"
              class="bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors">
              Add Contact
            </button>
          </form>
        </div>

        <div class="flex gap-2 justify-end mt-6 pt-4 border-t border-gray-100 dark:border-gray-700">
          <button @click="detailDialog.open = false"
            class="bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200 text-sm font-medium px-4 py-2 rounded-lg transition-colors">
            Cancel
          </button>
          <button @click="saveDetail"
            class="bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors">
            Save
          </button>
        </div>
      </div>
    </div>

    <!-- Stage Comment Dialog -->
    <div v-if="stageComment.open" class="fixed inset-0 bg-black/50 flex items-center justify-center z-[60] p-4">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl p-6 w-full max-w-sm">
        <h3 class="font-semibold text-gray-800 dark:text-gray-100 mb-1">Stage changed</h3>
        <p class="text-xs text-gray-500 dark:text-gray-400 mb-3">Add an optional comment for this transition.</p>
        <textarea v-model="stageComment.notes" rows="3" placeholder="How did it go? (optional)"
          class="w-full border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none mb-3"></textarea>
        <div v-if="stageComment.isLastStage" class="flex flex-col gap-1 mb-4">
          <label class="text-xs font-medium text-gray-600 dark:text-gray-400">Set status</label>
          <select v-model="stageComment.newStatus"
            class="border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500">
            <option v-for="s in statuses" :key="s" :value="s">{{ s.replace('_', ' ') }}</option>
          </select>
        </div>
        <div class="flex gap-2 justify-end">
          <button @click="stageComment.open = false"
            class="bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200 text-sm font-medium px-4 py-2 rounded-lg transition-colors">
            Cancel
          </button>
          <button @click="confirmStageComment"
            class="bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors">
            Save
          </button>
        </div>
      </div>
    </div>

    <!-- Manage Stages Dialog -->
    <div v-if="stagesMgmt" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl p-6 w-full max-w-sm">
        <div class="flex justify-between items-center mb-4">
          <h3 class="font-semibold text-gray-800 dark:text-gray-100">Manage Stages</h3>
          <button @click="stagesMgmt = false"
            class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 text-lg leading-none">✕</button>
        </div>
        <ul class="space-y-2 mb-4">
          <li v-for="(stage, idx) in stages" :key="stage.id"
            draggable="true"
            @dragstart="dragIdx = idx"
            @dragover.prevent="dragOverIdx = idx"
            @dragleave="dragOverIdx = null"
            @drop.prevent="dropStage(idx)"
            :class="dragOverIdx === idx ? 'ring-2 ring-blue-400 rounded' : ''"
            class="flex items-center gap-2 cursor-grab">
            <span class="text-gray-300 dark:text-gray-600 select-none text-sm">⠿</span>
            <input v-model="stage.name" @blur="updateStage(stage)"
              class="flex-1 border border-gray-300 dark:border-gray-600 rounded px-2 py-1.5 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-1 focus:ring-blue-500" />
            <button @click="removeStage(stage.id)"
              class="text-red-400 hover:text-red-600 px-1 text-base leading-none">✕</button>
          </li>
        </ul>
        <form @submit.prevent="addStage" class="flex gap-2">
          <input v-model="newStageName" placeholder="New stage name" required
            class="flex-1 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500" />
          <button type="submit"
            class="bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium px-3 py-2 rounded-lg transition-colors">
            Add
          </button>
        </form>
      </div>
    </div>
  </div>

    <!-- Default Stages Dialog -->
    <div v-if="defaultStagesMgmt" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl p-6 w-full max-w-sm">
        <div class="flex justify-between items-center mb-1">
          <h3 class="font-semibold text-gray-800 dark:text-gray-100">Default Stages</h3>
          <button @click="defaultStagesMgmt = false"
            class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 text-lg leading-none">✕</button>
        </div>
        <p class="text-xs text-gray-400 dark:text-gray-500 mb-4">Copied to every new job on creation.</p>
        <ul class="space-y-2 mb-4">
          <li v-for="(stage, idx) in defaultStages" :key="stage.id"
            draggable="true"
            @dragstart="dragDefaultIdx = idx"
            @dragover.prevent="dragOverDefaultIdx = idx"
            @dragleave="dragOverDefaultIdx = null"
            @drop.prevent="dropDefaultStage(idx)"
            :class="dragOverDefaultIdx === idx ? 'ring-2 ring-blue-400 rounded' : ''"
            class="flex items-center gap-2 cursor-grab">
            <span class="text-gray-300 dark:text-gray-600 select-none text-sm">⠿</span>
            <input v-model="stage.name" @blur="updateStage(stage)"
              class="flex-1 border border-gray-300 dark:border-gray-600 rounded px-2 py-1.5 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-1 focus:ring-blue-500" />
            <button @click="removeDefaultStage(stage.id)"
              class="text-red-400 hover:text-red-600 px-1 text-base leading-none">✕</button>
          </li>
        </ul>
        <form @submit.prevent="addDefaultStage" class="flex gap-2">
          <input v-model="newDefaultStageName" placeholder="New stage name" required
            class="flex-1 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500" />
          <button type="submit"
            class="bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium px-3 py-2 rounded-lg transition-colors">
            Add
          </button>
        </form>
      </div>
    </div>

    <!-- Delete Confirm Dialog -->
    <div v-if="confirmDelete.open" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl p-6 w-full max-w-sm">
        <h3 class="font-semibold text-gray-800 dark:text-gray-100 mb-2">Delete job application?</h3>
        <p class="text-sm text-gray-500 dark:text-gray-400 mb-6">This action cannot be undone.</p>
        <div class="flex justify-end gap-3">
          <button @click="confirmDelete.open = false"
            class="px-4 py-2 text-sm rounded-lg border border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
            Cancel
          </button>
          <button @click="doDelete"
            class="px-4 py-2 text-sm rounded-lg bg-red-500 hover:bg-red-600 text-white font-medium transition-colors">
            Delete
          </button>
        </div>
      </div>
    </div>

    <!-- Archive Confirm Dialog -->
    <div v-if="confirmArchive.open" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl p-6 w-full max-w-sm">
        <h3 class="font-semibold text-gray-800 dark:text-gray-100 mb-2">Archive job application?</h3>
        <p class="text-sm text-gray-500 dark:text-gray-400 mb-6">This marks the job as archived.</p>
        <div class="flex justify-end gap-3">
          <button @click="confirmArchive.open = false"
            class="px-4 py-2 text-sm rounded-lg border border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
            Cancel
          </button>
          <button @click="doArchive"
            class="px-4 py-2 text-sm rounded-lg bg-amber-500 hover:bg-amber-600 text-white font-medium transition-colors">
            Archive
          </button>
        </div>
      </div>
    </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'

const statuses = ['prospect', 'applied', 'in_progress', 'on_hold', 'negotiating', 'accepted', 'rejected', 'canceled']

const statusColors = {
  prospect:    'bg-gray-100 text-gray-600',
  applied:     'bg-blue-100 text-blue-700',
  in_progress: 'bg-purple-100 text-purple-700',
  on_hold:     'bg-amber-100 text-amber-700',
  negotiating: 'bg-indigo-100 text-indigo-700',
  accepted:    'bg-green-100 text-green-700',
  rejected:    'bg-red-100 text-red-600',
  canceled:    'bg-yellow-100 text-yellow-700',
}

function statusClass(s) {
  return statusColors[s] ?? 'bg-gray-100 text-gray-600'
}

const dark = ref(false)

function applyDark(val) {
  dark.value = val
  document.documentElement.classList.toggle('dark', val)
  localStorage.setItem('theme', val ? 'dark' : 'light')
}

function toggleDark() {
  applyDark(!dark.value)
}

const jobs = ref([])
const stages = ref([])
const filter = ref({ text: '', statuses: [], stages: [], dateFrom: '', dateTo: '' })

function escHtml(s) {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

function fuzzyMatch(query, target) {
  if (!query) return true
  const q = query.toLowerCase(), t = (target ?? '').toLowerCase()
  let qi = 0
  for (let i = 0; i < t.length && qi < q.length; i++) {
    if (t[i] === q[qi]) qi++
  }
  return qi === q.length
}

function highlight(query, target) {
  if (!query || !target) return escHtml(target ?? '')
  const q = query.toLowerCase()
  let qi = 0
  return [...target].map(ch => {
    const safe = escHtml(ch)
    if (qi < q.length && ch.toLowerCase() === q[qi]) { qi++; return `<mark class="bg-yellow-200 dark:bg-yellow-800 rounded-sm">${safe}</mark>` }
    return safe
  }).join('')
}

const filteredJobs = computed(() => {
  const { text, statuses, stages: stageIds, dateFrom, dateTo } = filter.value
  return jobs.value.filter(j => {
    if (text && !fuzzyMatch(text, j.company) && !fuzzyMatch(text, j.position)) return false
    if (statuses.length && !statuses.includes(j.status)) return false
    if (stageIds.length && !stageIds.includes(j.stage?.name)) return false
    const appliedDate = isoToDate(j.applied_at)
    if (dateFrom && appliedDate && appliedDate < dateFrom) return false
    if (dateTo && appliedDate && appliedDate > dateTo) return false
    if (archivedOnly.value ? !j.archived_at : (isActiveOnly.value && j.archived_at)) return false
    return true
  })
})

const isFiltered = computed(() =>
  filter.value.text || filter.value.statuses.length || filter.value.stages.length || filter.value.dateFrom || filter.value.dateTo
)

const activeFilterCount = computed(() => {
  const f = filter.value
  return (f.text ? 1 : 0) + f.statuses.length + f.stages.length + (f.dateFrom ? 1 : 0) + (f.dateTo ? 1 : 0)
})

const filtersOpen = ref(false)

function toggleFilter(arr, val) {
  const i = arr.indexOf(val)
  if (i >= 0) arr.splice(i, 1)
  else arr.push(val)
}

function clearFilter() {
  filter.value = { text: '', statuses: [], stages: [], dateFrom: '', dateTo: '' }
}

const closedStatuses = ['rejected', 'canceled']
const activeStatuses = statuses.filter(s => !closedStatuses.includes(s))
const isActiveOnly = computed(() =>
  filter.value.statuses.length === activeStatuses.length && activeStatuses.every(s => filter.value.statuses.includes(s))
)
function toggleActiveOnly() {
  const next = !isActiveOnly.value
  filter.value.statuses = next ? [...activeStatuses] : []
  localStorage.setItem('activeOnly', String(next))
  if (next) archivedOnly.value = false
}
if (localStorage.getItem('activeOnly') !== 'false') filter.value.statuses = [...activeStatuses]

const archivedOnly = ref(false)
function toggleArchivedOnly() {
  archivedOnly.value = !archivedOnly.value
  if (archivedOnly.value) {
    filter.value.statuses = []
    localStorage.setItem('activeOnly', 'false')
  }
}

const stageDropdownOpen = ref(false)
const stageSearch = ref('')
watch(stageDropdownOpen, open => { if (!open) stageSearch.value = '' })

const allFilterStages = computed(() => {
  const byName = new Map(defaultStages.value.map(s => [s.name, s]))
  for (const j of jobs.value) {
    for (const s of (j.stages ?? [])) {
      if (!byName.has(s.name)) byName.set(s.name, s)
    }
  }
  return [...byName.values()]
})

const filteredDropdownStages = computed(() => {
  if (!stageSearch.value) return allFilterStages.value
  return allFilterStages.value.filter(s => fuzzyMatch(stageSearch.value, s.name) || filter.value.stages.includes(s.name))
})

const chipDrag = ref({ active: false, action: null, arr: null })

function chipMousedown(arr, val) {
  const adding = !arr.includes(val)
  toggleFilter(arr, val)
  chipDrag.value = { active: true, action: adding ? 'add' : 'remove', arr }
}

function chipMouseenter(arr, val) {
  if (!chipDrag.value.active || chipDrag.value.arr !== arr) return
  const i = arr.indexOf(val)
  if (chipDrag.value.action === 'add' && i < 0) arr.push(val)
  else if (chipDrag.value.action === 'remove' && i >= 0) arr.splice(i, 1)
}

function chipMouseup() { chipDrag.value.active = false }
const form = ref(emptyForm())
const companyInput = ref(null)
const stagesMgmt = ref(false)
const newStageName = ref('')
const dragIdx = ref(null)
const dragOverIdx = ref(null)
const defaultStagesMgmt = ref(false)
const defaultStages = ref([])
const newDefaultStageName = ref('')
const dragDefaultIdx = ref(null)
const dragOverDefaultIdx = ref(null)
const stageDialog = ref({ open: false, job: null, stageId: null, notes: '', logs: [] })
const detailDialog = ref({ open: false, job: null, contacts: [], logs: [], edit: null })
const stageComment = ref({ open: false, notes: '', jobId: null, stageId: null, edit: null })
const confirmDelete = ref({ open: false, id: null })
const confirmArchive = ref({ open: false, job: null })
const newContact = ref({ name: '', role: '', email: '', phone: '' })
const pendingContacts = ref([])
const deletedContactIds = ref([])

// applied_at is edited as a local YYYY-MM-DD (date picker) but stored server-side as a
// timezone-aware RFC3339 timestamp. These helpers convert between the two representations.
function todayLocal() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

// applied_at is a calendar date. Take the wall-clock date straight off the stored
// value (bare YYYY-MM-DD or the date part of an RFC3339 timestamp) instead of routing
// it through `new Date()`, which would re-project the instant into the viewer's timezone
// and shift the day for values stored at UTC midnight (e.g. migrated legacy rows).
function isoToDate(iso) {
  if (!iso) return ''
  if (iso.length >= 10 && iso[4] === '-') return iso.slice(0, 10)
  return ''
}

function dateToISO(dateStr) {
  if (!dateStr) return null
  if (dateStr.includes('T')) return dateStr // already a full timestamp
  const d = new Date(`${dateStr}T00:00:00`) // local midnight
  if (isNaN(d)) return null
  const off = -d.getTimezoneOffset() // minutes east of UTC
  const sign = off >= 0 ? '+' : '-'
  const abs = Math.abs(off)
  const hh = String(Math.floor(abs / 60)).padStart(2, '0')
  const mm = String(abs % 60).padStart(2, '0')
  return `${dateStr}T00:00:00${sign}${hh}:${mm}`
}

// shallow copy of a job payload with applied_at serialized to a timezone-aware timestamp
function jobBody(obj) {
  return { ...obj, applied_at: dateToISO(obj.applied_at) }
}

function emptyForm() {
  return { id: null, company: '', position: '', status: 'applied', applied_at: todayLocal(), notes: '', url: '' }
}

async function load() {
  const res = await fetch('/api/jobs')
  jobs.value = await res.json()
}

async function loadStages(jobId) {
  const res = await fetch(`/api/jobs/${jobId}/stages`)
  stages.value = await res.json()
}

async function save() {
  if (!form.value.company || !form.value.position) return
  if (form.value.id) {
    await fetch(`/api/jobs/${form.value.id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(jobBody(form.value)),
    })
  } else {
    await fetch('/api/jobs', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(jobBody(form.value)),
    })
  }
  reset()
  await load()
  companyInput.value?.focus()
}

function remove(id) {
  confirmDelete.value = { open: true, id }
}

async function doDelete() {
  await fetch(`/api/jobs/${confirmDelete.value.id}`, { method: 'DELETE' })
  confirmDelete.value = { open: false, id: null }
  load()
}

function archive(job) {
  confirmArchive.value = { open: true, job }
}

async function setArchived(job, archivedAt) {
  await fetch(`/api/jobs/${job.id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      company: job.company, position: job.position, status: job.status,
      applied_at: dateToISO(job.applied_at), notes: job.notes, url: job.url,
      archived_at: archivedAt,
    }),
  })
  load()
}

async function doArchive() {
  await setArchived(confirmArchive.value.job, new Date().toISOString())
  confirmArchive.value = { open: false, job: null }
}

function unarchive(job) {
  setArchived(job, null)
}

function reset() {
  form.value = emptyForm()
}

function nextStageId(job) {
  const list = job.stages ?? []
  if (!job.stage_id || list.length === 0) return list[0]?.id ?? null
  const idx = list.findIndex(s => s.id === job.stage_id)
  if (idx < 0 || idx === list.length - 1) return job.stage_id
  return list[idx + 1].id
}

function stageProgress(job) {
  const list = job.stages ?? []
  if (!job.stage_id || list.length === 0) return 0
  const idx = list.findIndex(s => s.id === job.stage_id)
  return idx < 0 ? 0 : Math.round((idx + 1) / list.length * 100)
}

function formatDate(iso) {
  if (!iso) return ''
  // treat a bare YYYY-MM-DD as a local wall date (not UTC midnight); full RFC3339
  // timestamps render in the browser's local timezone
  const d = new Date(iso.length === 10 && iso[4] === '-' ? `${iso}T00:00:00` : iso)
  if (isNaN(d)) return ''
  return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })
}

// applied_at is a calendar date, not an instant: render the stored wall date as-is
// (see isoToDate) so it never shifts a day for the viewer's timezone.
function formatDay(iso) {
  const date = isoToDate(iso)
  if (!date) return ''
  return new Date(`${date}T00:00:00`).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })
}

function truncateNotes(notes, max = 60) {
  if (!notes) return ''
  return notes.length > max ? `${notes.slice(0, max)}…` : notes
}

async function openStageDialog(job) {
  const logs = await fetch(`/api/jobs/${job.id}/logs`).then(r => r.json())
  stageDialog.value = { open: true, job, stageId: nextStageId(job), notes: '', logs }
}

async function submitStageUpdate() {
  await fetch(`/api/jobs/${stageDialog.value.job.id}/logs`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ stage_id: stageDialog.value.stageId, notes: stageDialog.value.notes }),
  })
  stageDialog.value.open = false
  load()
}

async function openDetail(job) {
  const [contacts, logs, stagesData] = await Promise.all([
    fetch(`/api/jobs/${job.id}/contacts`).then(r => r.json()),
    fetch(`/api/jobs/${job.id}/logs`).then(r => r.json()),
    fetch(`/api/jobs/${job.id}/stages`).then(r => r.json()),
  ])
  stages.value = stagesData
  const edit = { company: job.company, position: job.position, status: job.status, applied_at: isoToDate(job.applied_at), notes: job.notes, url: job.url, stage_id: job.stage_id }
  detailDialog.value = { open: true, job, contacts, logs, edit }
  newContact.value = { name: '', role: '', email: '', phone: '' }
  pendingContacts.value = []
  deletedContactIds.value = []
}

async function saveDetail() {
  const { id } = detailDialog.value.job
  const edit = detailDialog.value.edit
  const newStage = edit.stage_id || null
  const oldStage = detailDialog.value.job.stage_id || null
  if (newStage !== oldStage) {
    const isLastStage = stages.value.length > 0 && stages.value[stages.value.length - 1]?.id === newStage
    stageComment.value = { open: true, notes: '', jobId: id, stageId: newStage, edit, isLastStage, newStatus: edit.status }
    return
  }
  await Promise.all([
    fetch(`/api/jobs/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(jobBody(edit)),
    }),
    ...pendingContacts.value.map(c => fetch(`/api/jobs/${id}/contacts`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(c),
    })),
    ...deletedContactIds.value.map(cid => fetch(`/api/jobs/${id}/contacts/${cid}`, { method: 'DELETE' })),
  ])
  detailDialog.value.open = false
  load()
}

async function confirmStageComment() {
  const { jobId, stageId, notes, edit, isLastStage, newStatus } = stageComment.value
  const { stage_id, ...rest } = edit
  if (isLastStage && newStatus) rest.status = newStatus
  await Promise.all([
    fetch(`/api/jobs/${jobId}/logs`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ stage_id: stageId || null, notes }),
    }),
    fetch(`/api/jobs/${jobId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(jobBody(rest)),
    }),
    ...pendingContacts.value.map(c => fetch(`/api/jobs/${jobId}/contacts`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(c),
    })),
    ...deletedContactIds.value.map(cid => fetch(`/api/jobs/${jobId}/contacts/${cid}`, { method: 'DELETE' })),
  ])
  stageComment.value.open = false
  detailDialog.value.open = false
  load()
}

function addContact() {
  pendingContacts.value.push({ ...newContact.value })
  newContact.value = { name: '', role: '', email: '', phone: '' }
}

function removeContact(id) {
  detailDialog.value.contacts = detailDialog.value.contacts.filter(c => c.id !== id)
  deletedContactIds.value.push(id)
}

async function addStage() {
  const maxOrder = stages.value.length > 0 ? Math.max(...stages.value.map(s => s.sort_order)) : 0
  const jobId = detailDialog.value.job.id
  await fetch(`/api/jobs/${jobId}/stages`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name: newStageName.value, sort_order: maxOrder + 1 }),
  })
  newStageName.value = ''
  await loadStages(jobId)
}

async function updateStage(stage) {
  await fetch(`/api/stages/${stage.id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name: stage.name, sort_order: stage.sort_order }),
  })
}

async function dropStage(toIdx) {
  const fromIdx = dragIdx.value
  dragIdx.value = null
  dragOverIdx.value = null
  if (fromIdx === null || fromIdx === toIdx) return
  const a = stages.value[fromIdx]
  const b = stages.value[toIdx]
  await Promise.all([
    fetch(`/api/stages/${a.id}`, { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ name: a.name, sort_order: b.sort_order }) }),
    fetch(`/api/stages/${b.id}`, { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ name: b.name, sort_order: a.sort_order }) }),
  ])
  await loadStages(detailDialog.value.job.id)
}

async function removeStage(id) {
  await fetch(`/api/stages/${id}`, { method: 'DELETE' })
  await loadStages(detailDialog.value.job.id)
}

async function loadDefaultStages() {
  defaultStages.value = await fetch('/api/stages').then(r => r.json())
}

async function addDefaultStage() {
  const maxOrder = defaultStages.value.length > 0 ? Math.max(...defaultStages.value.map(s => s.sort_order)) : 0
  await fetch('/api/stages', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name: newDefaultStageName.value, sort_order: maxOrder + 1 }),
  })
  newDefaultStageName.value = ''
  await loadDefaultStages()
}

async function dropDefaultStage(toIdx) {
  const fromIdx = dragDefaultIdx.value
  dragDefaultIdx.value = null
  dragOverDefaultIdx.value = null
  if (fromIdx === null || fromIdx === toIdx) return
  const a = defaultStages.value[fromIdx]
  const b = defaultStages.value[toIdx]
  await Promise.all([
    fetch(`/api/stages/${a.id}`, { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ name: a.name, sort_order: b.sort_order }) }),
    fetch(`/api/stages/${b.id}`, { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ name: b.name, sort_order: a.sort_order }) }),
  ])
  await loadDefaultStages()
}

async function removeDefaultStage(id) {
  await fetch(`/api/stages/${id}`, { method: 'DELETE' })
  await loadDefaultStages()
}

function onEsc(e) {
  if (e.key !== 'Escape') return
  if (confirmDelete.value.open) { confirmDelete.value.open = false; return }
  if (confirmArchive.value.open) { confirmArchive.value.open = false; return }
  if (stageComment.value.open) { stageComment.value.open = false; return }
  if (stagesMgmt.value) { stagesMgmt.value = false; return }
  if (defaultStagesMgmt.value) { defaultStagesMgmt.value = false; return }
  if (detailDialog.value.open) { detailDialog.value.open = false; return }
  if (stageDialog.value.open) { stageDialog.value.open = false; return }
  if (stageDropdownOpen.value) { stageDropdownOpen.value = false; return }
}

onMounted(() => {
  const saved = localStorage.getItem('theme')
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
  applyDark(saved ? saved === 'dark' : prefersDark)
  load()
  loadDefaultStages()
  companyInput.value?.focus()
  window.addEventListener('mouseup', chipMouseup)
  window.addEventListener('keydown', onEsc)
})

onUnmounted(() => {
  window.removeEventListener('mouseup', chipMouseup)
  window.removeEventListener('keydown', onEsc)
})
</script>
