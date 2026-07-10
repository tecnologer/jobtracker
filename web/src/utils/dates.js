// applied_at is edited as a local YYYY-MM-DD (date picker) but stored server-side as a
// timezone-aware RFC3339 timestamp. These helpers convert between the two representations.
export function todayLocal() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

// applied_at is a calendar date. Take the wall-clock date straight off the stored
// value (bare YYYY-MM-DD or the date part of an RFC3339 timestamp) instead of routing
// it through `new Date()`, which would re-project the instant into the viewer's timezone
// and shift the day for values stored at UTC midnight (e.g. migrated legacy rows).
export function isoToDate(iso) {
  if (!iso) return ''
  if (iso.length >= 10 && iso[4] === '-') return iso.slice(0, 10)
  return ''
}

export function dateToISO(dateStr) {
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

export function formatDate(iso) {
  if (!iso) return ''
  // treat a bare YYYY-MM-DD as a local wall date (not UTC midnight); full RFC3339
  // timestamps (real instants, e.g. StageLog.created_at, Meeting.scheduled_at)
  // render in the browser's local timezone, date and time
  const d = new Date(iso.length === 10 && iso[4] === '-' ? `${iso}T00:00:00` : iso)
  if (isNaN(d)) return ''
  return d.toLocaleString(undefined, { month: 'short', day: 'numeric', year: 'numeric', hour: 'numeric', minute: '2-digit' })
}

// applied_at is a calendar date, not an instant: render the stored wall date as-is
// (see isoToDate) so it never shifts a day for the viewer's timezone.
export function formatDay(iso) {
  const date = isoToDate(iso)
  if (!date) return ''
  return new Date(`${date}T00:00:00`).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })
}

// scheduled_at is a real instant (unlike applied_at's wall date): the datetime-local
// input value is serialized with the browser's own UTC offset via Date/toISOString,
// and rendered back through formatDate (viewer-local) — never the dateToISO/isoToDate
// wall-date helpers used for applied_at.
export function toRFC3339(dtLocal) {
  if (!dtLocal) return null
  const d = new Date(dtLocal)
  if (isNaN(d)) return null
  return d.toISOString()
}

export function toDatetimeLocal(iso) {
  if (!iso) return ''
  const d = new Date(iso)
  if (isNaN(d)) return ''
  const pad = n => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

// a meeting is "urgent" when it is upcoming and less than 24h away (FR-08)
export function isUrgent(iso) {
  if (!iso) return false
  const diffMs = new Date(iso).getTime() - Date.now()
  return diffMs >= 0 && diffMs <= 24 * 60 * 60 * 1000
}
