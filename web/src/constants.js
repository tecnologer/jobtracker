export const statuses = ['prospect', 'applied', 'in_progress', 'on_hold', 'negotiating', 'accepted', 'rejected', 'canceled']

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

export function statusClass(s) {
  return statusColors[s] ?? 'bg-gray-100 text-gray-600'
}

// solid shades for dashboard bar fills — the pill classes above are too pale
// for bars and their text-* half is meaningless on a div
const statusBarColors = {
  prospect:    'bg-gray-400',
  applied:     'bg-blue-500',
  in_progress: 'bg-purple-500',
  on_hold:     'bg-amber-500',
  negotiating: 'bg-indigo-500',
  accepted:    'bg-green-500',
  rejected:    'bg-red-500',
  canceled:    'bg-yellow-500',
}

export function statusBarClass(s) {
  return statusBarColors[s] ?? 'bg-gray-400'
}

export const closedStatuses = ['rejected', 'canceled']
export const activeStatuses = statuses.filter(s => !closedStatuses.includes(s))
