<template>
  <div>
    <!-- Phone: card list (FR-01/02/03/04) -->
    <div class="md:hidden">
      <button
        class="min-h-11 w-full mb-3 inline-flex items-center justify-center gap-1.5 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg transition-colors"
        @click="addJobOpen = true"
      >
        + {{ t('jobs.addJob') }}
      </button>

      <p
        v-if="jobs.length === 0"
        class="text-center px-4 py-10 text-gray-400 dark:text-gray-500 bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700"
      >
        {{ totalCount === 0 ? t('jobs.noApplications') : t('jobs.noResultsFilter') }}
      </p>

      <ul
        v-else
        class="space-y-3"
      >
        <li
          v-for="job in jobs"
          :key="job.id"
          class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 p-4"
          @click="onCardTap($event, job)"
        >
          <div class="flex items-start justify-between gap-2 mb-2">
            <div class="flex items-center gap-1.5 min-w-0">
              <button
                :title="job.top_match ? t('jobs.removeTopMatch') : t('jobs.markTopMatch')"
                :aria-label="job.top_match ? t('jobs.removeTopMatch') : t('jobs.markTopMatch')"
                class="min-h-11 min-w-11 shrink-0 inline-flex items-center justify-center"
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
              <span
                class="font-medium text-gray-800 dark:text-gray-100 truncate"
                v-html="highlight(filterText, job.company)"
              />
            </div>
            <span
              :class="statusClass(job.status)"
              class="shrink-0 inline-block px-2 py-0.5 rounded-full text-xs font-semibold"
            >
              {{ t('status.' + job.status) }}
            </span>
          </div>

          <div class="mb-2 text-sm text-gray-700 dark:text-gray-300">
            <a
              v-if="job.url"
              :href="job.url"
              target="_blank"
              rel="noopener"
              class="text-blue-600 dark:text-blue-400 hover:underline"
              @click.stop
              v-html="highlight(filterText, job.position)"
            />
            <span
              v-else
              v-html="highlight(filterText, job.position)"
            />
          </div>

          <div class="flex flex-col gap-1 mb-3">
            <div
              class="h-1.5 w-full bg-gray-200 dark:bg-gray-600 rounded-full overflow-hidden"
              :title="tStage(job.stage?.name)"
            >
              <div
                v-if="job.stage_id"
                class="h-full bg-purple-500 rounded-full transition-all"
                :style="`width: ${stageProgress(job)}%`"
              />
            </div>
            <span class="text-xs text-gray-400 dark:text-gray-500">{{ tStage(job.stage?.name) }}</span>
          </div>

          <div class="flex items-center justify-between">
            <span class="text-xs text-gray-500 dark:text-gray-400">{{ formatDay(job.applied_at) }}</span>
            <div class="flex items-center gap-1">
              <button
                v-if="job.archived_at"
                :aria-label="t('jobs.unarchive')"
                :title="t('jobs.unarchive')"
                class="min-h-11 min-w-11 inline-flex items-center justify-center rounded text-amber-600 hover:text-amber-800 hover:bg-amber-50 dark:hover:bg-amber-900 focus:outline-none focus-visible:ring-2 focus-visible:ring-amber-500 transition-colors"
                @click.stop="setArchived(job, null)"
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
                :aria-label="t('jobs.archive')"
                :title="t('jobs.archive')"
                class="min-h-11 min-w-11 inline-flex items-center justify-center rounded text-amber-600 hover:text-amber-800 hover:bg-amber-50 dark:hover:bg-amber-900 focus:outline-none focus-visible:ring-2 focus-visible:ring-amber-500 transition-colors"
                @click.stop="confirmArchive = { open: true, job }"
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
                :aria-label="t('common.delete')"
                :title="t('common.delete')"
                class="min-h-11 min-w-11 inline-flex items-center justify-center rounded text-red-500 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-900 focus:outline-none focus-visible:ring-2 focus-visible:ring-red-500 transition-colors"
                @click.stop="confirmDelete = { open: true, id: job.id }"
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
          </div>
        </li>
      </ul>
    </div>

    <!-- Bulk selection action bar (md+ only, like the selection itself) -->
    <div
      v-if="selected.size"
      class="hidden md:flex flex-wrap items-center gap-2 mb-3 px-4 py-2 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg text-sm text-gray-700 dark:text-gray-200"
    >
      <span class="font-medium">{{ t('bulk.selected', { n: selected.size }) }}</span>
      <button
        :disabled="bulkBusy"
        class="px-2 py-1 rounded text-amber-600 hover:text-amber-800 hover:bg-amber-50 dark:hover:bg-amber-900 focus:outline-none focus-visible:ring-2 focus-visible:ring-amber-500 disabled:opacity-50 transition-colors"
        @click="bulkConfirm = { open: true, action: 'archive' }"
      >
        {{ t('jobs.archive') }}
      </button>
      <button
        :disabled="bulkBusy"
        class="px-2 py-1 rounded text-amber-600 hover:text-amber-800 hover:bg-amber-50 dark:hover:bg-amber-900 focus:outline-none focus-visible:ring-2 focus-visible:ring-amber-500 disabled:opacity-50 transition-colors"
        @click="bulkUnarchive"
      >
        {{ t('jobs.unarchive') }}
      </button>
      <select
        v-model="bulkStatus"
        :disabled="bulkBusy"
        :aria-label="t('stages.setStatus')"
        class="border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-1 focus:ring-blue-500 disabled:opacity-50"
        @change="onBulkStatusPick"
      >
        <option
          value=""
          disabled
        >
          {{ t('bulk.setStatus') }}
        </option>
        <option
          v-for="s in statuses"
          :key="s"
          :value="s"
        >
          {{ t('status.' + s) }}
        </option>
      </select>
      <a
        :href="bulkExportHref"
        class="px-2 py-1 rounded text-blue-600 hover:text-blue-800 hover:bg-blue-100 dark:hover:bg-blue-900 focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 transition-colors"
      >
        {{ t('header.exportCsv') }}
      </a>
      <button
        :disabled="bulkBusy"
        class="px-2 py-1 rounded text-red-500 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-900 focus:outline-none focus-visible:ring-2 focus-visible:ring-red-500 disabled:opacity-50 transition-colors"
        @click="bulkConfirm = { open: true, action: 'delete' }"
      >
        {{ t('common.delete') }}
      </button>
      <button
        class="ml-auto px-2 py-1 rounded text-gray-500 hover:text-gray-700 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700 focus:outline-none focus-visible:ring-2 focus-visible:ring-gray-400 transition-colors"
        @click="selected = new Set()"
      >
        {{ t('bulk.clearSelection') }}
      </button>
    </div>

    <!-- Bulk run outcome: "N succeeded, M failed" (FR-08) -->
    <div
      v-if="bulkResult"
      class="hidden md:flex items-center gap-2 mb-3 px-4 py-2 rounded-lg border text-sm"
      :class="bulkResult.failed ? 'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800 text-red-700 dark:text-red-300' : 'bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800 text-green-700 dark:text-green-300'"
    >
      <span>{{ t('bulk.result', { done: bulkResult.done, failed: bulkResult.failed }) }}</span>
      <button
        :aria-label="t('common.cancel')"
        class="ml-auto text-lg leading-none opacity-60 hover:opacity-100"
        @click="bulkResult = null"
      >
        ✕
      </button>
    </div>

    <!-- Tablet/desktop: table (FR-12/13) -->
    <div class="hidden md:block bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden">
      <table class="w-full table-fixed text-sm">
        <thead class="bg-gray-50 dark:bg-gray-700 border-b border-gray-200 dark:border-gray-600">
          <tr>
            <th class="w-10 px-2 py-3 text-center">
              <input
                type="checkbox"
                :checked="allSelected"
                :indeterminate="selected.size > 0 && !allSelected"
                :aria-label="t('bulk.selectAll')"
                class="w-4 h-4 accent-blue-600 cursor-pointer align-middle"
                @change="toggleSelectAll"
              >
            </th>
            <th class="w-[14%] lg:w-[12%] text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">
              {{ t('common.company') }}
            </th>
            <th class="w-[26%] lg:w-[20%] text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">
              {{ t('common.position') }}
            </th>
            <th class="w-[11%] lg:w-[9%] text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">
              {{ t('common.status') }}
            </th>
            <th class="w-[15%] lg:w-[12%] text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">
              {{ t('common.stage') }}
            </th>
            <th class="w-[18%] lg:w-[14%] text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">
              {{ t('common.applied') }}
            </th>
            <th class="hidden lg:table-cell lg:w-[22%] text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">
              {{ t('common.notes') }}
            </th>
            <th class="w-[16%] lg:w-[12%] px-2 py-3" />
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
          <tr class="bg-blue-50/40 dark:bg-blue-900/10 border-b-2 border-blue-100 dark:border-blue-800">
            <td class="px-2 py-2" />
            <td class="px-4 py-2">
              <input
                ref="companyInput"
                v-model="form.company"
                :placeholder="t('common.company')"
                class="w-full border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-300 dark:placeholder-gray-600 focus:outline-none focus:ring-1 focus:ring-blue-500"
                @keydown.enter.prevent="save"
              >
            </td>
            <td class="px-4 py-2">
              <input
                v-model="form.position"
                :placeholder="t('common.position')"
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
                  {{ t('status.' + s) }}
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
            <td class="hidden lg:table-cell px-4 py-2">
              <input
                v-model="form.notes"
                :placeholder="t('common.notes')"
                class="w-full border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 placeholder-gray-300 dark:placeholder-gray-600 focus:outline-none focus:ring-1 focus:ring-blue-500"
                @keydown.enter.prevent="save"
              >
            </td>
            <td class="px-4 py-2 text-right pr-3">
              <button
                :disabled="saving"
                class="px-3 py-1 text-sm bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white rounded"
                @click="save"
              >
                {{ t('common.add') }}
              </button>
            </td>
          </tr>
          <tr v-if="jobs.length === 0">
            <td
              colspan="8"
              class="text-center px-4 py-10 text-gray-400 dark:text-gray-500"
            >
              {{ totalCount === 0 ? t('jobs.noApplications') : t('jobs.noResultsFilter') }}
            </td>
          </tr>
          <tr
            v-for="job in jobs"
            :key="job.id"
            class="hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
            @click="onRowClick($event, job)"
            @dblclick="onRowDblClick($event, job)"
          >
            <td class="px-2 py-3 text-center">
              <input
                type="checkbox"
                :checked="selected.has(job.id)"
                :aria-label="t('bulk.selectRow')"
                class="w-4 h-4 accent-blue-600 cursor-pointer align-middle"
                @change="toggleSelect(job.id)"
              >
            </td>
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
                {{ t('status.' + job.status) }}
              </span>
            </td>
            <td class="px-4 py-3">
              <div class="flex flex-col gap-1 min-w-32 max-w-40">
                <div
                  class="h-1.5 w-full bg-gray-200 dark:bg-gray-600 rounded-full overflow-hidden"
                  :title="tStage(job.stage?.name)"
                >
                  <div
                    v-if="job.stage_id"
                    class="h-full bg-purple-500 rounded-full transition-all"
                    :style="`width: ${stageProgress(job)}%`"
                  />
                </div>
                <span class="text-xs text-gray-400 dark:text-gray-500">{{ tStage(job.stage?.name) }}</span>
              </div>
            </td>
            <td class="px-4 py-3 text-gray-500 dark:text-gray-400">
              {{ formatDay(job.applied_at) }}
            </td>
            <td
              class="hidden lg:table-cell px-4 py-3 text-gray-500 dark:text-gray-400"
              :title="job.notes"
            >
              <div class="truncate">
                {{ truncateNotes(job.notes) }}
              </div>
            </td>
            <td class="px-2 py-3">
              <div class="flex flex-nowrap items-center gap-1 justify-end whitespace-nowrap">
                <button
                  :aria-label="t('jobs.viewDetails')"
                  :title="t('jobs.viewDetails')"
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
                  :aria-label="t('jobs.unarchive')"
                  :title="t('jobs.unarchive')"
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
                  :aria-label="t('jobs.archive')"
                  :title="t('jobs.archive')"
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
                  :aria-label="t('common.delete')"
                  :title="t('common.delete')"
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
    </div>

    <ConfirmDialog
      v-if="confirmDelete.open"
      :title="t('jobs.deleteConfirmTitle')"
      :message="t('jobs.deleteConfirmMessage')"
      :confirm-label="t('common.delete')"
      tone="red"
      @confirm="doDelete"
      @close="confirmDelete.open = false"
    />
    <ConfirmDialog
      v-if="confirmArchive.open"
      :title="t('jobs.archiveConfirmTitle')"
      :message="t('jobs.archiveConfirmMessage')"
      :confirm-label="t('jobs.archive')"
      tone="amber"
      @confirm="doArchive"
      @close="confirmArchive.open = false"
    />
    <ConfirmDialog
      v-if="bulkConfirm.open"
      :title="bulkDialog.title"
      :message="bulkDialog.message"
      :confirm-label="bulkDialog.label"
      :tone="bulkDialog.tone"
      @confirm="bulkConfirmed"
      @close="closeBulkConfirm"
    />
    <ConfirmDialog
      v-if="confirmDuplicate.open"
      :title="t('jobs.duplicateConfirmTitle')"
      :message="duplicateMessage"
      :confirm-label="t('jobs.createAnyway')"
      tone="amber"
      @confirm="saveDuplicate"
      @close="confirmDuplicate = { open: false, duplicate: null }"
    />

    <!-- Phone: add-job dialog (FR-04), same fields/save() as the inline row -->
    <BaseDialog
      v-if="addJobOpen"
      width="w-full max-w-md"
      @close="addJobOpen = false"
    >
      <div class="flex justify-between items-center mb-4">
        <h3 class="font-semibold text-gray-800 dark:text-gray-100">
          {{ t('jobs.addJobDialogTitle') }}
        </h3>
        <button
          class="min-h-11 min-w-11 inline-flex items-center justify-center text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 text-lg leading-none"
          @click="addJobOpen = false"
        >
          ✕
        </button>
      </div>
      <form
        class="flex flex-col gap-3"
        @submit.prevent="save"
      >
        <div class="flex flex-col gap-1">
          <label class="text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('common.company') }}</label>
          <input
            v-model="form.company"
            :placeholder="t('common.company')"
            required
            class="border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
        </div>
        <div class="flex flex-col gap-1">
          <label class="text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('common.position') }}</label>
          <input
            v-model="form.position"
            :placeholder="t('common.position')"
            required
            class="border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
        </div>
        <div class="flex flex-col gap-1">
          <label class="text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('common.status') }}</label>
          <select
            v-model="form.status"
            class="border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option
              v-for="s in statuses"
              :key="s"
              :value="s"
            >
              {{ t('status.' + s) }}
            </option>
          </select>
        </div>
        <div class="flex flex-col gap-1">
          <label class="text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('common.applied') }}</label>
          <input
            v-model="form.applied_at"
            type="date"
            class="border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
        </div>
        <div class="flex flex-col gap-1">
          <label class="text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('common.notes') }}</label>
          <textarea
            v-model="form.notes"
            rows="2"
            :placeholder="t('common.notes')"
            class="border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
          />
        </div>
        <div class="flex gap-2 justify-end mt-2">
          <button
            type="button"
            class="min-h-11 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200 text-sm font-medium px-4 py-2 rounded-lg transition-colors"
            @click="addJobOpen = false"
          >
            {{ t('common.cancel') }}
          </button>
          <button
            type="submit"
            :disabled="saving"
            class="min-h-11 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors"
          >
            {{ t('common.save') }}
          </button>
        </div>
      </form>
    </BaseDialog>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import * as api from '../api'
import { statuses, statusClass } from '../constants'
import { todayLocal, formatDay } from '../utils/dates'
import { highlight, truncateNotes } from '../utils/text'
import { useJobs } from '../composables/useJobs'
import { useI18n } from '../composables/useI18n'
import ConfirmDialog from './ConfirmDialog.vue'
import BaseDialog from './BaseDialog.vue'

const props = defineProps({
  jobs: { type: Array, required: true },
  filterText: { type: String, default: '' },
  totalCount: { type: Number, required: true },
})

const emit = defineEmits(['view'])

const { loadJobs, removeJob, setArchived, toggleTopMatch, jobPayload } = useJobs()
const { t, tStage } = useI18n()

function emptyForm() {
  return { company: '', position: '', status: 'applied', applied_at: todayLocal(), notes: '', url: '' }
}

const form = ref(emptyForm())
const saving = ref(false)
const companyInput = ref(null)
const confirmDelete = ref({ open: false, id: null })
const confirmArchive = ref({ open: false, job: null })
const confirmDuplicate = ref({ open: false, duplicate: null })
const addJobOpen = ref(false)

const duplicateMessage = computed(() => {
  const dup = confirmDuplicate.value.duplicate
  if (!dup) return ''
  const applied = dup.applied_at ? t('jobs.duplicateAppliedOn', { date: formatDay(dup.applied_at) }) : ''
  return t('jobs.duplicatePrompt', { company: dup.company, position: dup.position, applied })
})

async function save(allowDuplicate = false) {
  if (!form.value.company || !form.value.position) return
  if (saving.value) return
  saving.value = true
  try {
    await api.createJob(form.value, allowDuplicate === true)
  } catch (err) {
    if (err.status === 409) {
      confirmDuplicate.value = { open: true, duplicate: err.body?.duplicate }
      return
    }
    throw err // form state stays intact; global handler shows the toast
  } finally {
    saving.value = false
  }
  confirmDuplicate.value = { open: false, duplicate: null }
  form.value = emptyForm()
  await loadJobs()
  addJobOpen.value = false
  companyInput.value?.focus()
}

function saveDuplicate() {
  return save(true)
}

// close the dialog before awaiting so a double-click can't fire the action twice
async function doDelete() {
  const { id } = confirmDelete.value
  confirmDelete.value = { open: false, id: null }
  await removeJob(id)
}

async function doArchive() {
  const { job } = confirmArchive.value
  confirmArchive.value = { open: false, job: null }
  await setArchived(job, new Date().toISOString())
}

// --- bulk selection (md+ table only; REQUIREMENTS-bulk-actions.md) ---
const selected = ref(new Set()) // reassigned (never mutated) so reactivity is trivial
const bulkBusy = ref(false)
const bulkResult = ref(null) // { done, failed } after a bulk run
const bulkConfirm = ref({ open: false, action: null }) // 'delete' | 'archive' | 'status'
const bulkStatus = ref('')

const allSelected = computed(() => props.jobs.length > 0 && props.jobs.every(j => selected.value.has(j.id)))
const bulkExportHref = computed(() => `/api/jobs/export?ids=${[...selected.value].join(',')}`)

const bulkDialog = computed(() => {
  const n = selected.value.size
  switch (bulkConfirm.value.action) {
    case 'delete':
      return { title: t('bulk.deleteConfirmTitle'), message: t('bulk.deleteConfirmMessage', { n }), label: t('common.delete'), tone: 'red' }
    case 'archive':
      return { title: t('bulk.archiveConfirmTitle'), message: t('bulk.archiveConfirmMessage', { n }), label: t('jobs.archive'), tone: 'amber' }
    default:
      return { title: t('bulk.statusConfirmTitle'), message: t('bulk.statusConfirmMessage', { n, status: t('status.' + bulkStatus.value) }), label: t('common.save'), tone: 'amber' }
  }
})

function toggleSelect(id) {
  const next = new Set(selected.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  selected.value = next
}

function toggleSelectAll() {
  selected.value = allSelected.value ? new Set() : new Set(props.jobs.map(j => j.id))
}

// selection never outlives visibility: filter changes and deletions both flow
// through the jobs prop, so pruning here covers FR-03 without watching filters
watch(() => props.jobs, list => {
  if (!selected.value.size) return
  const visible = new Set(list.map(j => j.id))
  selected.value = new Set([...selected.value].filter(id => visible.has(id)))
})

// sequential per-id calls, best-effort, one refresh at the end (FR-07/FR-08)
async function runBulk(fn) {
  bulkBusy.value = true
  const targets = props.jobs.filter(j => selected.value.has(j.id))
  let failed = 0
  for (const job of targets) {
    try {
      await fn(job) // api.js throws on any non-2xx
    } catch {
      failed++
    }
  }
  await loadJobs()
  bulkResult.value = { done: targets.length - failed, failed }
  selected.value = new Set()
  bulkBusy.value = false
}

function onBulkStatusPick() {
  if (bulkStatus.value) bulkConfirm.value = { open: true, action: 'status' }
}

function closeBulkConfirm() {
  bulkConfirm.value = { open: false, action: null }
  bulkStatus.value = ''
}

function bulkConfirmed() {
  const action = bulkConfirm.value.action
  const status = bulkStatus.value
  closeBulkConfirm()
  if (action === 'delete') return runBulk(job => api.deleteJob(job.id))
  if (action === 'archive') return runBulk(job => api.updateJob(job.id, { ...jobPayload(job), archived_at: new Date().toISOString() }))
  return runBulk(job => api.updateJob(job.id, { ...jobPayload(job), status }))
}

function bulkUnarchive() {
  return runBulk(job => api.updateJob(job.id, { ...jobPayload(job), archived_at: null }))
}

function stageProgress(job) {
  const list = job.stages ?? []
  if (!job.stage_id || list.length === 0) return 0
  const idx = list.findIndex(s => s.id === job.stage_id)
  return idx < 0 ? 0 : Math.round((idx + 1) / list.length * 100)
}

function onRowDblClick(event, job) {
  // ignore double-clicks on the row's own controls (star, URL, action icons, checkbox)
  if (event.target.closest('button, a, input')) return
  emit('view', job)
}

// tablet-only tap-to-open (FR-13): desktop (>= 1024px) keeps dblclick-only behavior
function onRowClick(event, job) {
  if (event.target.closest('button, a, input')) return
  if (window.innerWidth >= 1024) return
  emit('view', job)
}

// phone card tap-to-open (FR-02): ignore taps on the card's own controls/links
function onCardTap(event, job) {
  if (event.target.closest('button, a')) return
  emit('view', job)
}

onMounted(() => companyInput.value?.focus())
</script>
