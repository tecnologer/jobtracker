export function escHtml(s) {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

export function fuzzyMatch(query, target) {
  if (!query) return true
  const q = query.toLowerCase(), t = (target ?? '').toLowerCase()
  let qi = 0
  for (let i = 0; i < t.length && qi < q.length; i++) {
    if (t[i] === q[qi]) qi++
  }
  return qi === q.length
}

export function highlight(query, target) {
  if (!query || !target) return escHtml(target ?? '')
  const q = query.toLowerCase()
  let qi = 0
  return [...target].map(ch => {
    const safe = escHtml(ch)
    if (qi < q.length && ch.toLowerCase() === q[qi]) { qi++; return `<mark class="bg-yellow-200 dark:bg-yellow-800 rounded-sm">${safe}</mark>` }
    return safe
  }).join('')
}

export function truncateNotes(notes, max = 60) {
  if (!notes) return ''
  return notes.length > max ? `${notes.slice(0, max)}…` : notes
}
