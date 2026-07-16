<template>
  <div class="mb-6 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl px-4 py-2">
    <div class="flex items-center gap-2 flex-wrap">
      <button
        class="min-h-11 md:min-h-0 flex items-center gap-2 text-sm font-medium text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200 transition-colors select-none py-1 flex-1"
        @click="filtersOpen = !filtersOpen"
      >
        <svg
          :class="filtersOpen ? 'rotate-90' : ''"
          class="w-4 h-4 shrink-0 transition-transform text-gray-500 dark:text-gray-400"
          viewBox="0 0 20 20"
          fill="currentColor"
          aria-hidden="true"
        ><path
          fill-rule="evenodd"
          d="M7.21 14.77a.75.75 0 01.02-1.06L11.168 10 7.23 6.29a.75.75 0 111.04-1.08l4.5 4.25a.75.75 0 010 1.08l-4.5 4.25a.75.75 0 01-1.06-.02z"
          clip-rule="evenodd"
        /></svg>
        {{ t('filters.filters') }}
        <span
          v-if="isFiltered"
          class="text-xs bg-blue-100 dark:bg-blue-900 text-blue-600 dark:text-blue-300 rounded-full px-1.5 py-0.5 font-medium leading-none"
        >{{ activeFilterCount }}</span>
      </button>
      <button
        :class="archivedOnly ? 'bg-amber-100 dark:bg-amber-900 text-amber-600 dark:text-amber-300 ring-2 ring-current' : 'bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400'"
        class="min-h-11 md:min-h-0 inline-flex items-center justify-center text-xs px-2.5 py-1 rounded-full font-medium transition-colors shrink-0"
        @click="toggleArchivedOnly"
      >
        {{ t('filters.archivedOnly') }}
      </button>
      <button
        :class="isActiveOnly ? 'bg-blue-100 dark:bg-blue-900 text-blue-600 dark:text-blue-300 ring-2 ring-current' : 'bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400'"
        class="min-h-11 md:min-h-0 inline-flex items-center justify-center text-xs px-2.5 py-1 rounded-full font-medium transition-colors shrink-0"
        @click="toggleActiveOnly"
      >
        {{ t('filters.activeOnly') }}
      </button>
      <button
        :class="topMatchOnly ? 'bg-amber-100 dark:bg-amber-900 text-amber-600 dark:text-amber-300 ring-2 ring-current' : 'bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400'"
        class="min-h-11 md:min-h-0 inline-flex items-center justify-center text-xs px-2.5 py-1 rounded-full font-medium transition-colors shrink-0"
        @click="toggleTopMatchOnly"
      >
        {{ t('filters.topMatches') }}
      </button>
    </div>

    <div
      v-show="filtersOpen"
      class="flex flex-col gap-2 mt-2 pb-2"
    >
      <div class="flex flex-col gap-1">
        <label class="text-xs font-medium text-gray-600 dark:text-gray-400">{{ t('filters.companyOrPosition') }}</label>
        <input
          v-model="filter.text"
          :placeholder="t('filters.companyOrPositionPlaceholder')"
          class="w-full border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-1.5 text-sm bg-white dark:bg-gray-800 text-gray-800 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
      </div>

      <!-- Row 2: status chips -->
      <div class="flex items-center gap-1.5 flex-wrap select-none">
        <span class="text-xs text-gray-400 dark:text-gray-500">{{ t('filters.statusColon') }}</span>
        <button
          v-for="s in statuses"
          :key="s"
          :class="filter.statuses.includes(s) ? statusClass(s) + ' ring-2 ring-offset-1 ring-current' : 'bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400'"
          class="min-h-11 md:min-h-0 inline-flex items-center justify-center px-2 py-0.5 rounded-full text-xs font-medium transition-colors cursor-pointer"
          @mousedown.prevent="chipMousedown(filter.statuses, s)"
          @mouseenter="chipMouseenter(filter.statuses, s)"
          @touchstart.prevent="chipTap(filter.statuses, s)"
          @dragstart.prevent
        >
          {{ t('status.' + s) }}
        </button>
      </div>

      <!-- Row 3: stage dropdown + applied date -->
      <div class="flex flex-col gap-2 md:flex-row md:flex-wrap md:items-center md:gap-3 lg:flex-nowrap">
        <div class="flex items-center gap-1.5">
          <span class="text-xs text-gray-400 dark:text-gray-500">{{ t('filters.stageColon') }}</span>
          <div class="relative">
            <button
              class="min-h-11 md:min-h-0 flex items-center justify-between gap-1 w-44 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-1 text-xs bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
              @click="stageDropdownOpen = !stageDropdownOpen"
            >
              {{ t('filters.stageButtonLabel') }}{{ filter.stages.length ? ` (${filter.stages.length})` : '' }}
              <span class="text-gray-400">▾</span>
            </button>
            <div
              v-if="stageDropdownOpen"
              class="absolute z-20 top-full mt-1 min-w-36 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 rounded-lg shadow-lg py-1"
            >
              <div
                v-if="allFilterStages.length"
                class="px-2 pt-1.5 pb-1 relative"
              >
                <input
                  v-model="stageSearch"
                  :placeholder="t('filters.filterStagesPlaceholder')"
                  class="w-full border border-gray-200 dark:border-gray-600 rounded px-2 py-1 pr-6 text-xs bg-white dark:bg-gray-700 text-gray-700 dark:text-gray-200 placeholder-gray-400 focus:outline-none focus:ring-1 focus:ring-purple-400"
                >
                <button
                  v-if="stageSearch"
                  class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 leading-none"
                  @click="stageSearch = ''"
                >
                  ✕
                </button>
              </div>
              <p
                v-if="!filteredDropdownStages.length"
                class="px-3 py-1.5 text-xs text-gray-400"
              >
                {{ allFilterStages.length ? t('filters.noMatch') : t('filters.noStagesConfigured') }}
              </p>
              <label
                v-for="s in filteredDropdownStages"
                :key="s.id"
                class="flex items-center gap-2 px-3 py-1.5 hover:bg-gray-50 dark:hover:bg-gray-700 cursor-pointer text-sm text-gray-700 dark:text-gray-200"
              >
                <input
                  v-model="filter.stages"
                  type="checkbox"
                  :value="s.name"
                  class="rounded accent-purple-500"
                >
                <span v-html="highlight(stageSearch, tStage(s.name))" />
              </label>
            </div>
            <div
              v-if="stageDropdownOpen"
              class="fixed inset-0 z-10"
              @click="stageDropdownOpen = false"
            />
          </div>
        </div>

        <div class="flex items-center gap-1.5 flex-1">
          <span class="text-xs text-gray-400 dark:text-gray-500 shrink-0">{{ t('filters.appliedColon') }}</span>
          <input
            v-model="filter.dateFrom"
            type="date"
            class="flex-1 min-w-0 border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-xs bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 focus:outline-none focus:ring-1 focus:ring-blue-500"
          >
          <span class="text-xs text-gray-400 shrink-0">–</span>
          <input
            v-model="filter.dateTo"
            type="date"
            class="flex-1 min-w-0 border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-xs bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 focus:outline-none focus:ring-1 focus:ring-blue-500"
          >
        </div>

        <button
          v-if="isFiltered"
          class="text-xs text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 underline transition-colors"
          @click="clearFilter"
        >
          {{ t('filters.clear') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted } from 'vue'
import { statuses, statusClass } from '../constants'
import { highlight } from '../utils/text'
import { useJobFilters } from '../composables/useJobFilters'
import { useI18n } from '../composables/useI18n'

const { t, tStage } = useI18n()
const {
  filter, filtersOpen, archivedOnly, topMatchOnly,
  isFiltered, activeFilterCount, isActiveOnly,
  stageDropdownOpen, stageSearch, allFilterStages, filteredDropdownStages,
  clearFilter, toggleActiveOnly, toggleArchivedOnly, toggleTopMatchOnly,
  chipTap, chipMousedown, chipMouseenter, chipMouseup,
} = useJobFilters()

function onEsc(e) {
  if (e.key === 'Escape') stageDropdownOpen.value = false
}

onMounted(() => {
  window.addEventListener('mouseup', chipMouseup)
  window.addEventListener('keydown', onEsc)
})

onUnmounted(() => {
  window.removeEventListener('mouseup', chipMouseup)
  window.removeEventListener('keydown', onEsc)
})
</script>
