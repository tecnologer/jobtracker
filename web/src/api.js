import { dateToISO } from './utils/dates'

// every non-2xx response throws, so a failed request can never be mistaken for
// success; .status/.body let callers branch on typed errors (e.g. 409 duplicate)
export class ApiError extends Error {
  constructor(status, body, message) {
    super(message)
    this.status = status
    this.body = body
  }
}

async function fail(res) {
  const text = await res.text().catch(() => '')
  let body = null
  try { body = JSON.parse(text) } catch { /* plain-text error body (http.Error) */ }
  return new ApiError(res.status, body, body?.error || text.trim() || `HTTP ${res.status}`)
}

async function get(path) {
  const res = await fetch(path)
  if (!res.ok) throw await fail(res)
  return res.json()
}

async function request(path, method, body) {
  const res = await fetch(path, {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw await fail(res)
  return res
}

// shallow copy of a job payload with applied_at serialized to a timezone-aware timestamp
function jobBody(obj) {
  return { ...obj, applied_at: dateToISO(obj.applied_at) }
}

// jobs
export const fetchJobs = () => get('/api/jobs')
export const createJob = (job, allowDuplicate = false) => request(`/api/jobs${allowDuplicate ? '?allow_duplicate=1' : ''}`, 'POST', jobBody(job))
export const updateJob = (id, job) => request(`/api/jobs/${id}`, 'PUT', jobBody(job))
export const deleteJob = id => request(`/api/jobs/${id}`, 'DELETE')
export const setTopMatch = (id, topMatch) => request(`/api/jobs/${id}/top-match`, 'PUT', { top_match: topMatch })

// CSV import: multipart upload, so it bypasses the JSON `request` wrapper above.
export const importJobs = async file => {
  const body = new FormData()
  body.append('file', file)
  const res = await fetch('/api/jobs/import', { method: 'POST', body })
  if (!res.ok) throw await fail(res)
  return res.json()
}

// dashboard stats
export const fetchStats = () => get('/api/stats')

// build version (ldflags-injected; "dev" in local dev)
export const fetchVersion = () => get('/api/version')

// stage logs
export const fetchLogs = jobId => get(`/api/jobs/${jobId}/logs`)
export const addLog = (jobId, body) => request(`/api/jobs/${jobId}/logs`, 'POST', body)

// contacts
export const fetchContacts = jobId => get(`/api/jobs/${jobId}/contacts`)
export const addContact = (jobId, contact) => request(`/api/jobs/${jobId}/contacts`, 'POST', contact)
export const deleteContact = (jobId, contactId) => request(`/api/jobs/${jobId}/contacts/${contactId}`, 'DELETE')

// stages (job_id=0 rows are the default/template stages)
export const fetchDefaultStages = () => get('/api/stages')
export const addDefaultStage = body => request('/api/stages', 'POST', body)
export const fetchJobStages = jobId => get(`/api/jobs/${jobId}/stages`)
export const addJobStage = (jobId, body) => request(`/api/jobs/${jobId}/stages`, 'POST', body)
export const updateStage = stage => request(`/api/stages/${stage.id}`, 'PUT', { name: stage.name, sort_order: stage.sort_order })
export const deleteStage = id => request(`/api/stages/${id}`, 'DELETE')
export const swapStageOrder = (a, b) => Promise.all([
  request(`/api/stages/${a.id}`, 'PUT', { name: a.name, sort_order: b.sort_order }),
  request(`/api/stages/${b.id}`, 'PUT', { name: b.name, sort_order: a.sort_order }),
])

// meetings
export const fetchUpcomingMeetings = () => get('/api/meetings/upcoming')
export const fetchJobMeetings = jobId => get(`/api/jobs/${jobId}/meetings`)
export const addMeeting = (jobId, body) => request(`/api/jobs/${jobId}/meetings`, 'POST', body)
export const updateMeeting = (jobId, id, body) => request(`/api/jobs/${jobId}/meetings/${id}`, 'PUT', body)
export const deleteMeeting = (jobId, id) => request(`/api/jobs/${jobId}/meetings/${id}`, 'DELETE')
