<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900">
    <AppHeader @open-job="openJob" />

    <main class="max-w-6xl mx-auto px-6">
      <JobFilters />
      <JobsTable
        :jobs="filteredJobs"
        :filter-text="filter.text"
        :total-count="jobs.length"
        @view="detailJob = $event"
      />
    </main>

    <JobDetailDialog
      v-if="detailJob"
      :job="detailJob"
      @close="detailJob = null"
      @saved="onSaved"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useJobs } from './composables/useJobs'
import { useJobFilters } from './composables/useJobFilters'
import { useStages } from './composables/useStages'
import { useMeetings } from './composables/useMeetings'
import { useDarkMode } from './composables/useDarkMode'
import AppHeader from './components/AppHeader.vue'
import JobFilters from './components/JobFilters.vue'
import JobsTable from './components/JobsTable.vue'
import JobDetailDialog from './components/JobDetailDialog.vue'

const { jobs, loadJobs } = useJobs()
const { filter, filteredJobs } = useJobFilters()
const { loadDefaultStages } = useStages()
const { loadUpcomingMeetings } = useMeetings()
const { initDarkMode } = useDarkMode()

const detailJob = ref(null)

function openJob(jobId) {
  const job = jobs.value.find(j => j.id === jobId)
  if (job) detailJob.value = job
}

async function onSaved() {
  detailJob.value = null
  await loadJobs()
}

onMounted(() => {
  initDarkMode()
  loadJobs()
  loadDefaultStages()
  loadUpcomingMeetings()
})
</script>
