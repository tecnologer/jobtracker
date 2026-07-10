import pluginVue from 'eslint-plugin-vue'

export default [
  { ignores: ['dist/**', 'node_modules/**', 'coverage/**'] },
  ...pluginVue.configs['flat/recommended'],
]
