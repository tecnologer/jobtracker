import { ref } from 'vue'
import * as api from '../api'

// default/template stages (job_id=0), copied to every new job on creation
const defaultStages = ref([])

export function useStages() {
  async function loadDefaultStages() {
    defaultStages.value = await api.fetchDefaultStages()
  }

  async function addDefaultStage(name) {
    const maxOrder = defaultStages.value.length > 0 ? Math.max(...defaultStages.value.map(s => s.sort_order)) : 0
    await api.addDefaultStage({ name, sort_order: maxOrder + 1 })
    await loadDefaultStages()
  }

  async function renameDefaultStage(stage, name) {
    stage.name = name
    await api.updateStage(stage)
  }

  async function removeDefaultStage(id) {
    await api.deleteStage(id)
    await loadDefaultStages()
  }

  async function reorderDefaultStages(fromIdx, toIdx) {
    await api.swapStageOrder(defaultStages.value[fromIdx], defaultStages.value[toIdx])
    await loadDefaultStages()
  }

  return { defaultStages, loadDefaultStages, addDefaultStage, renameDefaultStage, removeDefaultStage, reorderDefaultStages }
}
