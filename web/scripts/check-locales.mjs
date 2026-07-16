// web/scripts/check-locales.mjs — compare every catalog's keys against en.json
import { readdirSync, readFileSync } from 'node:fs'

const dir = new URL('../src/locales/', import.meta.url)
const load = f => JSON.parse(readFileSync(new URL(f, dir), 'utf8'))
const en = Object.keys(load('en.json'))
let failed = false

for (const file of readdirSync(dir).filter(f => f.endsWith('.json') && f !== 'en.json')) {
  const keys = new Set(Object.keys(load(file)))
  const missing = en.filter(k => !keys.has(k))
  const extra = [...keys].filter(k => !en.includes(k))
  if (missing.length) console.warn(`WARN ${file}: ${missing.length} missing key(s) (will fall back to English):\n  ${missing.join('\n  ')}`)
  if (extra.length) { console.error(`FAIL ${file}: unknown key(s), typo or removed from en.json:\n  ${extra.join('\n  ')}`); failed = true }
}
process.exit(failed ? 1 : 0)
