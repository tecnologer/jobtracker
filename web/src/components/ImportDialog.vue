<template>
  <BaseDialog
    width="w-full max-w-2xl"
    @close="emit('close')"
  >
    <div class="flex justify-between items-center mb-4">
      <h3 class="font-semibold text-gray-800 dark:text-gray-100">
        {{ t('import.title') }}
      </h3>
      <button
        class="min-h-11 min-w-11 inline-flex items-center justify-center md:min-h-0 md:min-w-0 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 text-lg leading-none"
        @click="emit('close')"
      >
        ✕
      </button>
    </div>

    <p class="text-sm text-gray-700 dark:text-gray-300 mb-3">
      {{ t('import.created', { n: result.created }) }}
      <template v-if="result.stages_created">
        {{ t('import.stagesCreated', { n: result.stages_created }) }}
      </template>
    </p>

    <div
      v-if="result.errors.length"
      class="mb-3"
    >
      <p class="text-xs font-medium text-red-600 dark:text-red-400 mb-1">
        {{ t('import.errors') }}
      </p>
      <ul class="text-xs text-gray-600 dark:text-gray-300 space-y-0.5 max-h-24 overflow-y-auto">
        <li
          v-for="err in result.errors"
          :key="`err-${err.row}`"
        >
          {{ t('import.rowMessage', { row: err.row, message: err.message }) }}
        </li>
      </ul>
    </div>

    <div
      v-if="result.warnings.length"
      class="mb-3"
    >
      <p class="text-xs font-medium text-amber-600 dark:text-amber-400 mb-1">
        {{ t('import.warnings') }}
      </p>
      <ul class="text-xs text-gray-600 dark:text-gray-300 space-y-0.5 max-h-24 overflow-y-auto">
        <li
          v-for="warn in result.warnings"
          :key="`warn-${warn.row}`"
        >
          {{ t('import.rowMessage', { row: warn.row, message: warn.message }) }}
        </li>
      </ul>
    </div>

    <div v-if="rows.length">
      <div class="flex items-center justify-between mb-1">
        <p class="text-xs font-medium text-gray-600 dark:text-gray-400">
          {{ t('import.duplicates') }}
        </p>
        <div class="flex items-center gap-1.5">
          <label class="text-xs text-gray-600 dark:text-gray-400">{{ t('import.applyToAll') }}</label>
          <select
            v-model="bulkAction"
            class="border border-gray-300 dark:border-gray-600 rounded px-1.5 py-1 text-xs bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100"
            @change="applyBulkAction"
          >
            <option value="skip">
              {{ t('import.actionSkip') }}
            </option>
            <option value="update">
              {{ t('import.actionUpdate') }}
            </option>
            <option value="import">
              {{ t('import.actionImport') }}
            </option>
          </select>
        </div>
      </div>
      <div class="max-h-64 overflow-y-auto border border-gray-200 dark:border-gray-700 rounded-lg">
        <table class="w-full text-xs">
          <thead class="bg-gray-50 dark:bg-gray-700 sticky top-0">
            <tr>
              <th class="text-left px-2 py-1.5 font-semibold text-gray-600 dark:text-gray-300">
                {{ t('import.colRow') }}
              </th>
              <th class="text-left px-2 py-1.5 font-semibold text-gray-600 dark:text-gray-300">
                {{ t('import.colJob') }}
              </th>
              <th class="text-left px-2 py-1.5 font-semibold text-gray-600 dark:text-gray-300">
                {{ t('import.colAction') }}
              </th>
              <th class="text-left px-2 py-1.5 font-semibold text-gray-600 dark:text-gray-300">
                {{ t('common.status') }}
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-gray-700">
            <tr
              v-for="row in rows"
              :key="row.row"
            >
              <td class="px-2 py-1.5 text-gray-500 dark:text-gray-400">
                {{ row.row }}
              </td>
              <td class="px-2 py-1.5 text-gray-700 dark:text-gray-300">
                {{ row.job.company }} — {{ row.job.position }}
              </td>
              <td class="px-2 py-1.5">
                <select
                  v-model="row.action"
                  :disabled="row.status === 'done'"
                  class="border border-gray-300 dark:border-gray-600 rounded px-1.5 py-1 text-xs bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100"
                >
                  <option value="skip">
                    {{ t('import.actionSkip') }}
                  </option>
                  <option value="update">
                    {{ t('import.actionUpdate') }}
                  </option>
                  <option value="import">
                    {{ t('import.actionImport') }}
                  </option>
                </select>
              </td>
              <td class="px-2 py-1.5 text-gray-500 dark:text-gray-400">
                {{ statusLabel(row) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="flex justify-end mt-3">
        <button
          :disabled="applying"
          class="min-h-11 md:min-h-0 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors"
          @click="applyAll"
        >
          {{ applying ? t('import.applying') : t('import.apply') }}
        </button>
      </div>
    </div>
  </BaseDialog>
</template>

<script setup>
import { ref } from 'vue'
import * as api from '../api'
import { useJobs } from '../composables/useJobs'
import { useI18n } from '../composables/useI18n'
import BaseDialog from './BaseDialog.vue'

const props = defineProps({
  result: { type: Object, required: true },
})

const emit = defineEmits(['close', 'applied'])

const { jobs } = useJobs()
const { t } = useI18n()

// local per-row resolution state; duplicates only ever live client-side while
// this dialog is open (see REQUIREMENTS.md "Duplicate state lives only in the client")
const rows = ref(props.result.duplicates.map(d => ({ ...d, action: 'skip', status: null })))
const applying = ref(false)
const bulkAction = ref('skip')

function applyBulkAction() {
  for (const row of rows.value) {
    if (row.status === null) row.action = bulkAction.value
  }
}

// case/trim-insensitive stage name match, mirroring the backend's matchStage
function matchStageID(stages, name) {
  if (!name) return null
  const target = name.trim().toLowerCase()
  const match = (stages ?? []).find(s => s.name.trim().toLowerCase() === target)
  return match ? match.id : null
}

async function applyUpdate(row) {
  const existing = jobs.value.find(j => j.id === row.existing.id)
  if (!existing) throw new Error(`job ${row.existing.id} no longer exists`)

  await api.updateJob(existing.id, {
    company: row.job.company,
    position: row.job.position,
    status: row.job.status,
    applied_at: row.job.applied_at,
    notes: row.job.notes,
    url: row.job.url,
    archived_at: row.job.archived ? new Date().toISOString() : null,
    stage_id: matchStageID(existing.stages, row.job.stage),
  })
  await api.setTopMatch(existing.id, row.job.top_match)
}

async function applyImport(row) {
  const res = await api.createJob({
    company: `${row.job.company} (csv)`,
    position: row.job.position,
    status: row.job.status,
    applied_at: row.job.applied_at,
    notes: row.job.notes,
    url: row.job.url,
    top_match: row.job.top_match,
    archived_at: row.job.archived ? new Date().toISOString() : null,
  }, true)

  if (!row.job.stage) return
  const created = await res.json()
  const stages = await api.fetchJobStages(created.id)
  const stageID = matchStageID(stages, row.job.stage)
  if (stageID) await api.addLog(created.id, { stage_id: stageID })
}

// row.status holds a machine code ('skipped'/'done') or a raw error message;
// translation happens here so labels follow language switches
function statusLabel(row) {
  if (row.status === 'done') return t('import.statusDone')
  if (row.status === 'skipped') return t('import.statusSkipped')
  return row.status ?? ''
}

async function applyAll() {
  applying.value = true
  for (const row of rows.value) {
    if (row.action === 'skip') {
      row.status = 'skipped'
      continue
    }
    try {
      if (row.action === 'update') await applyUpdate(row)
      else await applyImport(row)
      row.status = 'done'
    } catch (err) {
      row.status = err.message
    }
  }
  applying.value = false
  emit('applied')
}
</script>
