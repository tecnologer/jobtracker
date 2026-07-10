import { ref } from 'vue'
import * as api from '../api'

const jobs = ref([])

export function useJobs() {
  async function loadJobs() {
    jobs.value = await api.fetchJobs()
  }

  async function removeJob(id) {
    await api.deleteJob(id)
    await loadJobs()
  }

  async function setArchived(job, archivedAt) {
    await api.updateJob(job.id, {
      company: job.company, position: job.position, status: job.status,
      applied_at: job.applied_at, notes: job.notes, url: job.url,
      archived_at: archivedAt,
    })
    await loadJobs()
  }

  // top_match is toggled through a dedicated endpoint (not the general PUT /api/jobs/{id})
  // so it can never be clobbered by a full-body save that omits it. Mutate the job object
  // directly so a dialog open on the same job (same reference) reflects the change even
  // though loadJobs() below replaces jobs.value with freshly-fetched objects.
  async function toggleTopMatch(job) {
    const next = !job.top_match
    await api.setTopMatch(job.id, next)
    job.top_match = next
    await loadJobs()
  }

  return { jobs, loadJobs, removeJob, setArchived, toggleTopMatch }
}
