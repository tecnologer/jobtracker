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

export const closedStatuses = ['rejected', 'canceled']
export const activeStatuses = statuses.filter(s => !closedStatuses.includes(s))
