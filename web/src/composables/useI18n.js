import { ref } from 'vue'

const catalogs = Object.fromEntries(
  Object.entries(import.meta.glob('../locales/*.json', { eager: true, import: 'default' }))
    .map(([path, messages]) => [path.match(/([\w-]+)\.json$/)[1], messages]),
)

export const locale = ref('en') // exported for utils/dates.js

// available languages for the picker: [{ code: 'en', name: 'English' }, ...]
const languages = Object.entries(catalogs)
  .map(([code, messages]) => ({ code, name: messages['_meta.name'] ?? code }))
  .sort((a, b) => a.code.localeCompare(b.code))

function lookup(key) {
  return catalogs[locale.value]?.[key] ?? catalogs.en[key]
}

function t(key, params) {
  let msg
  if (params && 'n' in params) {
    // plural: try key.<category> (one/other/few/many...), fall back to key.other, then key
    const category = new Intl.PluralRules(locale.value).select(params.n)
    msg = lookup(`${key}.${category}`) ?? lookup(`${key}.other`) ?? lookup(key)
  } else {
    msg = lookup(key)
  }
  if (msg === undefined) return key // last-resort fallback: show the key, never blank
  if (params) msg = msg.replace(/\{(\w+)\}/g, (_, name) => params[name] ?? `{${name}}`)
  return msg
}

// DB-sourced names (stages): translate when a key exists, otherwise show the raw value
function tStage(name) {
  return (name && lookup(`stage.${name}`)) || name || ''
}

function setLocale(code) {
  if (!catalogs[code]) return
  locale.value = code
  localStorage.setItem('lang', code)
  document.documentElement.lang = code
}

function initLocale() {
  const saved = localStorage.getItem('lang')
  if (saved && catalogs[saved]) { setLocale(saved); return }
  // browser auto-detect: first navigator language whose base tag has a catalog
  for (const tag of navigator.languages ?? [navigator.language]) {
    const base = tag.toLowerCase().split('-')[0]
    if (catalogs[base]) { locale.value = base; document.documentElement.lang = base; return }
  }
}

export function useI18n() {
  return { locale, languages, t, tStage, setLocale, initLocale }
}
