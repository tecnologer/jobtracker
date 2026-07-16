<template>
  <div class="pb-10">
    <p
      v-if="loading"
      class="py-12 text-center text-sm text-gray-400 dark:text-gray-500"
    >
      {{ t('dashboard.loading') }}
    </p>

    <div
      v-else-if="error"
      class="py-12 text-center"
    >
      <p class="mb-3 text-sm text-gray-600 dark:text-gray-300">
        {{ t('dashboard.failedToLoad') }}
      </p>
      <button
        class="text-sm px-3 py-1.5 rounded-lg border border-gray-300 dark:border-gray-600 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
        @click="load"
      >
        {{ t('dashboard.retry') }}
      </button>
    </div>

    <template v-else-if="stats">
      <div class="mb-8 grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
        <div
          v-for="kpi in kpis"
          :key="kpi.label"
          class="rounded-lg border border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-800"
        >
          <p class="text-xs uppercase text-gray-400 dark:text-gray-500">
            {{ kpi.label }}
          </p>
          <p class="mt-1 text-2xl font-bold text-gray-800 dark:text-gray-100">
            {{ kpi.value }}
          </p>
        </div>
      </div>

      <section class="mb-8">
        <h2 class="mb-3 text-sm font-semibold text-gray-800 dark:text-gray-100">
          {{ t('dashboard.statusBreakdown') }}
        </h2>
        <BarList :rows="statusRows" />
      </section>

      <section class="mb-8">
        <h2 class="mb-3 text-sm font-semibold text-gray-800 dark:text-gray-100">
          {{ t('dashboard.stageFunnel') }}
        </h2>
        <BarList :rows="funnelRows" />
      </section>

      <section class="mb-8">
        <h2 class="mb-3 text-sm font-semibold text-gray-800 dark:text-gray-100">
          {{ t('dashboard.avgDaysPerStage') }}
        </h2>
        <BarList
          v-if="stageTimeRows.length"
          :rows="stageTimeRows"
        />
        <p
          v-else
          class="text-sm text-gray-400 dark:text-gray-500"
        >
          {{ t('dashboard.noStageTimingData') }}
        </p>
      </section>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { fetchStats } from '../api'
import { statuses, statusBarClass } from '../constants'
import { useI18n } from '../composables/useI18n'
import BarList from './BarList.vue'

const { t, tStage } = useI18n()
const stats = ref(null)
const loading = ref(false)
const error = ref(false)

async function load() {
  loading.value = true
  error.value = false
  try {
    stats.value = await fetchStats()
  } catch {
    error.value = true
  } finally {
    loading.value = false
  }
}

onMounted(load)

const kpis = computed(() => [
  { label: t('dashboard.totalJobs'), value: stats.value.total_jobs },
  { label: t('dashboard.active'), value: stats.value.active_jobs },
  { label: t('dashboard.offers'), value: stats.value.offers },
  { label: t('dashboard.rejectionRate'), value: `${Math.round(stats.value.rejection_rate * 100)}%` },
  {
    label: t('dashboard.avgDaysToResponse'),
    value: stats.value.avg_days_to_first_response == null
      ? '—'
      : `${stats.value.avg_days_to_first_response.toFixed(1)}d`,
  },
])

const statusRows = computed(() => statuses.map(s => ({
  label: t('status.' + s),
  value: stats.value.status_breakdown[s] ?? 0,
  barClass: statusBarClass(s),
})))

const funnelRows = computed(() => stats.value.funnel.map(s => ({
  label: tStage(s.name),
  value: s.jobs_reached,
})))

// avg_days: null means "no data", not 0 days — omit those rows
const stageTimeRows = computed(() => stats.value.funnel
  .filter(s => s.avg_days != null)
  .map(s => ({
    label: tStage(s.name),
    value: s.avg_days,
    display: `${s.avg_days.toFixed(1)}d`,
  })))
</script>
