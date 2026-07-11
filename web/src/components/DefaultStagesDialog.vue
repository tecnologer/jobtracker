<template>
  <BaseDialog
    width="w-full max-w-sm"
    @close="emit('close')"
  >
    <div class="flex justify-between items-center mb-1">
      <h3 class="font-semibold text-gray-800 dark:text-gray-100">
        Default Stages
      </h3>
      <button
        class="min-h-11 min-w-11 inline-flex items-center justify-center md:min-h-0 md:min-w-0 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 text-lg leading-none"
        @click="emit('close')"
      >
        ✕
      </button>
    </div>
    <p class="text-xs text-gray-400 dark:text-gray-500 mb-4">
      Copied to every new job on creation.
    </p>
    <StageListEditor
      :stages="defaultStages"
      @add="addDefaultStage"
      @rename="renameDefaultStage"
      @remove="removeDefaultStage"
      @reorder="reorderDefaultStages"
    />
  </BaseDialog>
</template>

<script setup>
import { useStages } from '../composables/useStages'
import BaseDialog from './BaseDialog.vue'
import StageListEditor from './StageListEditor.vue'

const emit = defineEmits(['close'])

const { defaultStages, addDefaultStage, renameDefaultStage, removeDefaultStage, reorderDefaultStages } = useStages()
</script>
