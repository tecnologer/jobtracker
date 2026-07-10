import { ref, computed, watch } from 'vue'
import { activeStatuses } from '../constants'
import { fuzzyMatch } from '../utils/text'
import { isoToDate } from '../utils/dates'
import { useJobs } from './useJobs'
import { useStages } from './useStages'

const { jobs } = useJobs()
const { defaultStages } = useStages()

const filter = ref({ text: '', statuses: [], stages: [], dateFrom: '', dateTo: '' })
const filtersOpen = ref(false)
const archivedOnly = ref(false)
const topMatchOnly = ref(false)
const stageDropdownOpen = ref(false)
const stageSearch = ref('')

watch(stageDropdownOpen, open => { if (!open) stageSearch.value = '' })

if (localStorage.getItem('activeOnly') !== 'false') filter.value.statuses = [...activeStatuses]

const filteredJobs = computed(() => {
  const { text, statuses, stages: stageNames, dateFrom, dateTo } = filter.value
  return jobs.value.filter(j => {
    if (text && !fuzzyMatch(text, j.company) && !fuzzyMatch(text, j.position)) return false
    if (statuses.length && !statuses.includes(j.status)) return false
    if (stageNames.length && !stageNames.includes(j.stage?.name)) return false
    const appliedDate = isoToDate(j.applied_at)
    if (dateFrom && appliedDate && appliedDate < dateFrom) return false
    if (dateTo && appliedDate && appliedDate > dateTo) return false
    if (archivedOnly.value ? !j.archived_at : (isActiveOnly.value && j.archived_at)) return false
    if (topMatchOnly.value && !j.top_match) return false
    return true
  })
})

const isFiltered = computed(() =>
  filter.value.text || filter.value.statuses.length || filter.value.stages.length || filter.value.dateFrom || filter.value.dateTo || topMatchOnly.value
)

const activeFilterCount = computed(() => {
  const f = filter.value
  return (f.text ? 1 : 0) + f.statuses.length + f.stages.length + (f.dateFrom ? 1 : 0) + (f.dateTo ? 1 : 0) + (topMatchOnly.value ? 1 : 0)
})

const isActiveOnly = computed(() =>
  filter.value.statuses.length === activeStatuses.length && activeStatuses.every(s => filter.value.statuses.includes(s))
)

const allFilterStages = computed(() => {
  const byName = new Map(defaultStages.value.map(s => [s.name, s]))
  for (const j of jobs.value) {
    for (const s of (j.stages ?? [])) {
      if (!byName.has(s.name)) byName.set(s.name, s)
    }
  }
  return [...byName.values()]
})

const filteredDropdownStages = computed(() => {
  if (!stageSearch.value) return allFilterStages.value
  return allFilterStages.value.filter(s => fuzzyMatch(stageSearch.value, s.name) || filter.value.stages.includes(s.name))
})

const chipDrag = ref({ active: false, action: null, arr: null })

export function useJobFilters() {
  function toggleFilter(arr, val) {
    const i = arr.indexOf(val)
    if (i >= 0) arr.splice(i, 1)
    else arr.push(val)
  }

  function clearFilter() {
    filter.value = { text: '', statuses: [], stages: [], dateFrom: '', dateTo: '' }
    topMatchOnly.value = false
  }

  function toggleActiveOnly() {
    const next = !isActiveOnly.value
    filter.value.statuses = next ? [...activeStatuses] : []
    localStorage.setItem('activeOnly', String(next))
    if (next) archivedOnly.value = false
  }

  function toggleArchivedOnly() {
    archivedOnly.value = !archivedOnly.value
    if (archivedOnly.value) {
      filter.value.statuses = []
      localStorage.setItem('activeOnly', 'false')
    }
  }

  function toggleTopMatchOnly() {
    topMatchOnly.value = !topMatchOnly.value
  }

  function chipMousedown(arr, val) {
    const adding = !arr.includes(val)
    toggleFilter(arr, val)
    chipDrag.value = { active: true, action: adding ? 'add' : 'remove', arr }
  }

  function chipMouseenter(arr, val) {
    if (!chipDrag.value.active || chipDrag.value.arr !== arr) return
    const i = arr.indexOf(val)
    if (chipDrag.value.action === 'add' && i < 0) arr.push(val)
    else if (chipDrag.value.action === 'remove' && i >= 0) arr.splice(i, 1)
  }

  // the window mouseup listener belongs to the one component using chip drag (JobFilters)
  function chipMouseup() { chipDrag.value.active = false }

  return {
    filter, filtersOpen, archivedOnly, topMatchOnly,
    filteredJobs, isFiltered, activeFilterCount, isActiveOnly,
    stageDropdownOpen, stageSearch, allFilterStages, filteredDropdownStages,
    clearFilter, toggleActiveOnly, toggleArchivedOnly, toggleTopMatchOnly,
    chipMousedown, chipMouseenter, chipMouseup,
  }
}
