# Frontend Architecture: Progression Dashboard

Scope: the dashboard from `REQUIREMENTS.md` (KPI cards FR-03, status breakdown FR-04, stage funnel FR-05/06, avg time per stage FR-07), toggled inside the existing SPA (FR-01), fed by one `GET /api/stats` call (FR-02). No router, no state library, no new dependencies (NFR-02). Conventions follow the split documented in `web/ARCHITECTURE.md`: Composition API `<script setup>`, flat `components/`, module-scope composables only for genuinely shared state, `npm run lint` must pass (NFR-05).

The payload contract in section 4 is owned by the backend: it mirrors `ARCHITECTURE-dashboard-backend.md` §3 (repo root) exactly. If the two ever disagree, the backend doc wins and this one gets updated.

## 1. What gets added

```
web/src/
├── api.js                     # +1 line: fetchStats
├── constants.js               # +1 export: statusBarClass (solid bar colors per status)
└── components/
    ├── DashboardView.vue      # NEW feature component: owns the stats fetch and all four widgets
    └── BarList.vue            # NEW dumb component: labeled horizontal CSS bars (used 3x)
```

Two new files total. Justification per file:

- **`DashboardView.vue`** — the view itself; unavoidable.
- **`BarList.vue`** — status breakdown, stage funnel, and avg-time-per-stage are the same visual structure (label + proportional bar + value). One dumb component beats triplicating the markup. It joins the existing dumb tier (BaseDialog, ConfirmDialog, StageLogList, StageListEditor): props only, no composables, no api.js.

Deliberately NOT added:

- **No `useStats.js` composable.** Stats state is consumed by exactly one component tree and never mutated elsewhere. The module-scope-composable pattern (useJobs, useMeetings) exists for *shared* state; a single-consumer fetch belongs in the component. If a second consumer ever appears (e.g. a KPI badge in the header), lift the refs into `composables/useStats.js` then, it is a 10-line move.
- **No per-widget components** (`StatusBreakdown.vue`, `StageFunnel.vue`, ...). Each would be a heading plus one `<BarList>`, ~12 lines of wrapper. They stay as sections inside `DashboardView.vue`. Split only if a widget grows real logic.
- **No chart abstraction, no chart library** (NFR-02). Bars are divs with Tailwind widths.

`components/` stays flat at 12 files; still under the "revisit at 2x" threshold noted in `ARCHITECTURE.md`.

## 2. Header toggle (FR-01)

`App.vue` owns a single new ref; `AppHeader` gets one prop and one emit, matching its existing `open-job` emit style:

```
App.vue
  showDashboard = ref(false)          // resets on refresh: FR-01 needs no persistence
  <AppHeader :dashboard-open="showDashboard" @toggle-view="showDashboard = !showDashboard" />
  <main>
    <template v-if="!showDashboard">
      <JobFilters /> <JobsTable ... />
    </template>
    <DashboardView v-else />
  </main>
```

- New header button sits with the existing button group, same Tailwind classes as its siblings. Label flips: `Dashboard` / `Jobs` (driven by the `dashboardOpen` prop).
- `v-if`/`v-else`, not `v-show`: DashboardView remounts on every toggle, so `onMounted` refetches and the numbers are always fresh (matters for A-08, time-in-current-stage moves with "now"). Table state survives the round trip for free because filters and jobs live in module-scope composables, not in the unmounted components.
- Everything else in the header (Export CSV, meetings, stages, dark toggle) stays visible in both views; no conditional header.

## 3. Data fetch, loading, error

`api.js` gains one line, same shape as every other GET:

```js
export const fetchStats = () => get('/api/stats')
```

`DashboardView.vue` owns three local refs, no composable (see section 1):

```js
const stats = ref(null)
const loading = ref(false)
const error = ref(false)

async function load() {
  loading.value = true
  error.value = false
  try { stats.value = await api.fetchStats() }
  catch { error.value = true }
  finally { loading.value = false }
}
onMounted(load)
```

Template states, in order:

1. `loading` → centered "Loading…" text. No skeletons; the payload is small (NFR-01: <200 ms).
2. `error` → short message + Retry button calling `load()`. This is the first explicit error state in the app; it stays local to the dashboard and does not introduce a global toast/error layer (out of scope, per `ARCHITECTURE.md` section 5).
3. `stats` → the four widgets.

Empty data (FR-09) is NOT a frontend branch: the server returns zeros/empty structures and nulls, the widgets render them (see sections 4 and 6). The only frontend guard is BarList's divide-by-zero handling.

## 4. Stats payload shape (contract with `GET /api/stats`)

Verbatim from `ARCHITECTURE-dashboard-backend.md` §3: flat KPI fields (no nested `kpis` object) and one merged `funnel` array feeding both the funnel widget and the avg-days widget.

```json
{
  "total_jobs": 42,
  "active_jobs": 17,
  "offers": 3,
  "rejection_rate": 0.45,
  "avg_days_to_first_response": 6.2,
  "status_breakdown": {
    "prospect": 5, "applied": 10, "in_progress": 8, "on_hold": 2,
    "negotiating": 1, "accepted": 1, "rejected": 12, "canceled": 3
  },
  "funnel": [
    { "name": "Phone Screen",        "sort_order": 1, "jobs_reached": 20, "avg_days": 4.0 },
    { "name": "Technical Interview", "sort_order": 2, "jobs_reached": 12, "avg_days": 6.5 },
    { "name": "Code Challenge",      "sort_order": 3, "jobs_reached": 7,  "avg_days": null },
    { "name": "Final Round",         "sort_order": 4, "jobs_reached": 4,  "avg_days": 2.1 },
    { "name": "Offer",               "sort_order": 5, "jobs_reached": 3,  "avg_days": 1.0 },
    { "name": "Other",               "sort_order": 6, "jobs_reached": 2,  "avg_days": 3.0 }
  ]
}
```

What the frontend relies on:

- **KPIs are top-level fields** (FR-03). Counts are integers. `rejection_rate` is always a float 0–1 (the server sends `0`, not null, when the denominator is 0). `avg_days_to_first_response` is a float or `null` (no responses yet, FR-08); `null` renders as `—`. The frontend computes nothing itself (FR-02) and owns all rounding (backend sends raw floats).
- **`status_breakdown`** (FR-04): map keyed by status, all 8 keys always present, zeros included. The frontend iterates `constants.js` `statuses` for order and `statusBarClass` for color; the server does not own presentation order.
- **`funnel`** (FR-05/06/07): one server-ordered array serving BOTH bar widgets. All default stages always present (zeros included, FR-09); `"Other"` appended last only when `jobs_reached > 0`. `avg_days` is a float or `null` (no time samples for that bucket). `sort_order` is informational; the frontend renders array order and ignores it.
- Empty DB (FR-09): zeros, `avg_days_to_first_response: null`, breakdown all zeros, funnel = the default stages with `jobs_reached: 0` and `avg_days: null`, no "Other". The dashboard renders zero-width bars; no NaN paths exist because the frontend does no arithmetic on the payload except bar-width scaling.

Day-difference math (`applied_at` wall date vs `created_at` instant) is entirely server-side; the frontend never touches dates in this payload.

## 5. Component hierarchy and contracts

```
App.vue                          + showDashboard ref, v-if/v-else in <main>
├── AppHeader                    + prop: dashboardOpen (Boolean)  + emit: toggle-view
└── DashboardView                no props, no emits; owns fetch + loading/error
    ├── [inline] KPI cards       grid of 5 cards, v-for over a computed array
    ├── [inline] Status section  <h2> + <BarList :rows="statusRows">
    ├── [inline] Funnel section  <h2> + <BarList :rows="funnelRows">
    └── [inline] Time section    <h2> + <BarList :rows="stageTimeRows">
        BarList (×3)             props: rows [{ label, value, display, barClass? }]
```

`DashboardView` responsibilities:

- Fetch + state (section 3).
- Three tiny computeds mapping payload to BarList rows, this is where formatting lives, keeping BarList dumb. The funnel and time widgets are two projections of the same `stats.funnel` array:
  - `statusRows`: iterate `statuses` from `constants.js`, `label` from the status name (same capitalization treatment the table uses), `value: stats.status_breakdown[s]`, `barClass: statusBarClass(s)`.
  - `funnelRows`: `stats.funnel.map(s => ({ label: s.name, value: s.jobs_reached, display: String(s.jobs_reached) }))`, single accent color.
  - `stageTimeRows`: `stats.funnel.filter(s => s.avg_days != null).map(s => ({ label: s.name, value: s.avg_days, display: s.avg_days.toFixed(1) + 'd' }))`. Buckets with `avg_days: null` are omitted from this widget rather than shown as zero, a null average means "no data", not "0 days". "Other" appears here too when the server sends an average for it.
- KPI formatting: `rejection_rate` → `Math.round(rate * 100) + '%'` (always a number, no null branch); `avg_days_to_first_response == null ? '—' : v.toFixed(1) + 'd'`.

KPI cards are a `v-for` over a computed `[{ label, value }]` array inside DashboardView, styled like the existing white/dark panels (`bg-white dark:bg-gray-800 border ... rounded-lg`). No `KpiCard.vue`: it is a div with two text nodes.

## 6. CSS bars: `BarList.vue`

Decision: **one reusable dumb component**, not per-widget markup. Three widgets share the exact structure; the component is ~30 lines and removes two copies.

Props:

```js
rows: Array   // [{ label: String, value: Number, display: String?, barClass: String? }]
              // display defaults to String(value); barClass defaults to 'bg-blue-500'
```

No emits, no slots in v1. Markup per row (illustrative):

```html
<div v-for="row in rows" :key="row.label" class="flex items-center gap-3">
  <span class="w-36 shrink-0 truncate text-sm text-gray-600 dark:text-gray-300 text-right">
    {{ row.label }}
  </span>
  <div class="flex-1 h-5 rounded bg-gray-100 dark:bg-gray-700 overflow-hidden">
    <div
      class="h-full rounded transition-all"
      :class="row.barClass ?? 'bg-blue-500'"
      :style="{ width: width(row) }"
    />
  </div>
  <span class="w-16 shrink-0 text-sm tabular-nums text-gray-800 dark:text-gray-100">
    {{ row.display ?? row.value }}
  </span>
</div>
```

Scaling: `width(row) = max > 0 ? (row.value / max * 100) + '%' : '0%'` where `max` is a computed over `rows`. The `max === 0` guard is the FR-09 divide-by-zero protection. Zero-value rows render label + empty track + `0`, they are not hidden (FR-04 shows all 8 statuses; FR-09 funnel shows all default stages at zero).

Colors: `constants.js` `statusColors` are pill classes (`bg-blue-100 text-blue-700`), too pale for solid bars and meaningless `text-*` on a div. Add one parallel export to `constants.js`, same hues at bar strength, satisfying "existing status colors" (FR-04) without repurposing pill classes:

```js
const statusBarColors = {
  prospect: 'bg-gray-400',    applied: 'bg-blue-500',
  in_progress: 'bg-purple-500', on_hold: 'bg-amber-500',
  negotiating: 'bg-indigo-500', accepted: 'bg-green-500',
  rejected: 'bg-red-500',     canceled: 'bg-yellow-500',
}
export function statusBarClass(s) { return statusBarColors[s] ?? 'bg-gray-400' }
```

Solid 400/500 shades are legible on both `bg-gray-100` (light track) and `dark:bg-gray-700` (dark track), covering NFR-05 dark mode without dark-variant bar classes.

## 7. Rules recap (unchanged from ARCHITECTURE.md, restated for the new files)

- `BarList.vue` is dumb: props only, no composables, no `api.js`, no constants imports (colors arrive via `barClass`).
- `DashboardView.vue` is a feature component: may import `api.js`, `constants.js`; imports no composables today (fetch is local).
- No new window/lifecycle listeners anywhere in this feature.
- Both new `.vue` files must pass `cd web && npm run lint` before done (CLAUDE.md rule); multi-word names (`DashboardView`, `BarList`) satisfy `vue/multi-word-component-names`.
