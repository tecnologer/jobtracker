<template>
  <footer class="mt-auto border-t border-gray-200 dark:border-gray-700 px-3 md:px-6 py-2 text-center text-xs text-gray-400 dark:text-gray-500">
    <a
      href="https://github.com/tecnologer/jobtracker/blob/main/LICENSE"
      target="_blank"
      rel="noopener"
      class="hover:underline"
    >
      {{ t('footer.license') }}
    </a>
    ·
    <a
      v-if="version !== 'dev'"
      :href="`https://github.com/tecnologer/jobtracker/releases/tag/${version}`"
      target="_blank"
      rel="noopener"
      class="hover:underline"
    >
      {{ version }}
    </a>
    <span v-else>{{ version }}</span>
  </footer>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { fetchVersion } from '../api'
import { useI18n } from '../composables/useI18n'

const { t } = useI18n()
const version = ref('dev')

onMounted(async () => {
  try {
    version.value = (await fetchVersion()).version
  } catch {
    // offline/dev backend missing: keep the "dev" fallback
  }
})
</script>
