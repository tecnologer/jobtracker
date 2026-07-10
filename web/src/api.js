import { dateToISO } from './utils/dates'

function get(path) {
  return fetch(path).then(r => r.json())
}

function request(path, method, body) {
  return fetch(path, {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
}

// shallow copy of a job payload with applied_at serialized to a timezone-aware timestamp
function jobBody(obj) {
  return { ...obj, applied_at: dateToISO(obj.applied_at) }
}

// jobs
export const fetchJobs = () => get('/api/jobs')
export const createJob = job => request('/api/jobs', 'POST', jobBody(job))
export const updateJob = (id, job) => request(`/api/jobs/${id}`, 'PUT', jobBody(job))
export const deleteJob = id => request(`/api/jobs/${id}`, 'DELETE')
export const setTopMatch = (id, topMatch) => request(`/api/jobs/${id}/top-match`, 'PUT', { top_match: topMatch })

// dashboard stats
export const fetchStats = () => get('/api/stats')

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
