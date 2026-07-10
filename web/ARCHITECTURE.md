# Frontend Architecture: splitting App.vue

Target: break `web/src/App.vue` (2024 lines) into an idiomatic Vue 3 structure with the smallest reasonable diff. No router, no Pinia, no new dependencies. Composition API with `<script setup>` throughout. `eslint-plugin-vue` `flat/recommended` must pass after every step.

## 1. Directory structure

```
web/src/
├── main.js                      # unchanged
├── style.css                    # unchanged: Tailwind entry + dark custom-variant (global)
├── App.vue                      # thin shell: layout + dialog orchestration (~80 lines)
├── api.js                       # fetch helper + one function per endpoint
├── constants.js                 # statuses, statusColors, statusClass, closedStatuses, activeStatuses
├── utils/
│   ├── dates.js                 # wall-date and instant helpers (moved VERBATIM, comments included)
│   └── text.js                  # escHtml, fuzzyMatch, highlight, truncateNotes
├── composables/
│   ├── useJobs.js               # shared jobs list + job mutations
│   ├── useJobFilters.js         # filter state + filteredJobs
│   ├── useStages.js             # shared defaultStages + stage API ops
│   ├── useMeetings.js           # shared upcomingMeetings
│   └── useDarkMode.js           # dark ref + applyDark/toggleDark
└── components/
    ├── BaseDialog.vue           # overlay shell shared by every dialog + Esc handling
    ├── ConfirmDialog.vue        # generic confirm (replaces delete + archive dialogs)
    ├── StageLogList.vue         # stage-history <ul> (dumb, props only)
    ├── StageListEditor.vue      # drag-reorder/rename/delete/add stage list (dumb, props/emits)
    ├── AppHeader.vue            # title, Export CSV, upcoming-meetings dropdown, default-stages button, dark toggle
    ├── DefaultStagesDialog.vue  # BaseDialog + StageListEditor wired to /api/stages
    ├── JobFilters.vue           # collapsible filter panel + quick-filter chips
    ├── JobsTable.vue            # table incl. inline add-row and per-row actions + confirm dialogs
    ├── JobDetailDialog.vue      # edit form, stage history, contacts, meetings, nested dialogs
    └── StageCommentDialog.vue   # "Stage changed" comment + optional status prompt
```

Flat `components/` on purpose: one view, ten components. Feature folders start paying off around 2x this size; revisit only if a second view appears.

## 2. What moves where (by current name in App.vue)

### `utils/dates.js`
Moved **verbatim, comments included** (CLAUDE.md wall-date rule, do not "improve"):
- `todayLocal`, `isoToDate`, `dateToISO` (applied_at wall-date helpers)
- `formatDay` (wall-date rendering for `applied_at`)
- `formatDate` (real instants only: `StageLog.created_at`, `Meeting.scheduled_at`)
- `toRFC3339`, `toDatetimeLocal` (meeting datetime-local helpers, with their comment)
- `isUrgent` (meeting <24h check)

### `utils/text.js`
- `escHtml`, `fuzzyMatch`, `highlight` (used by filters, table, stage dropdown)
- `truncateNotes`

### `constants.js`
- `statuses`, `statusColors`, `statusClass`
- `closedStatuses`, `activeStatuses`

### `api.js`
One `request(path, { method, body })` helper wrapping `fetch` + JSON headers + `res.json()`, then thin named functions replacing every inline `fetch`:

- Jobs: `fetchJobs`, `createJob(body)`, `updateJob(id, body)`, `deleteJob(id)`, `setTopMatch(id, next)`
  - `jobBody(obj)` (the `applied_at: dateToISO(...)` serializer) moves here; `createJob`/`updateJob` apply it so callers pass the raw form.
- Logs: `fetchLogs(jobId)`, `addLog(jobId, { stage_id, notes })`
- Contacts: `fetchContacts(jobId)`, `addContact(jobId, c)`, `deleteContact(jobId, cid)`
- Stages: `fetchDefaultStages`, `addDefaultStage(body)`, `fetchJobStages(jobId)`, `addJobStage(jobId, body)`, `updateStage(stage)`, `deleteStage(id)`, `swapStageOrder(a, b)` (the paired sort_order PUTs currently duplicated in `dropStage`/`dropDefaultStage`)
- Meetings: `fetchUpcomingMeetings`, `fetchJobMeetings(jobId)`, `addMeeting(jobId, body)`, `updateMeeting(jobId, id, body)`, `deleteMeeting(jobId, id)`

No error handling layer beyond what exists today (none). Add a toast/error path only when the app actually needs it.

### `composables/useJobs.js`
Module-scope shared state (see section 4):
- `jobs` ref, `load` (renamed `loadJobs`)
- `save` logic split: `createJob`-side lives in JobsTable's handler, but the shared mutations live here: `removeJob(id)` (was `doDelete` body), `setArchived(job, archivedAt)`, `toggleTopMatch(job)` (keep its mutate-in-place comment: the detail dialog holds the same object reference)

### `composables/useJobFilters.js`
- `filter` ref, `filtersOpen`, `archivedOnly`, `topMatchOnly`
- `filteredJobs`, `isFiltered`, `activeFilterCount`, `isActiveOnly`
- `toggleFilter`, `clearFilter`, `toggleActiveOnly` (incl. localStorage `activeOnly` persistence and initial read), `toggleArchivedOnly`, `toggleTopMatchOnly`
- `stageDropdownOpen`, `stageSearch`, its `watch`, `allFilterStages` (imports `defaultStages` from useStages), `filteredDropdownStages`
- `chipDrag`, `chipMousedown`, `chipMouseenter`, `chipMouseup`
- Note: the `window.addEventListener('mouseup', chipMouseup)` registration does NOT go in the composable (it is called from more than one component); JobFilters.vue registers/removes it in its own `onMounted`/`onUnmounted`.

### `composables/useStages.js`
- `defaultStages` ref (shared: used by DefaultStagesDialog and by `allFilterStages` in useJobFilters)
- `loadDefaultStages`, `addDefaultStage`, `removeDefaultStage`, `dropDefaultStage` reorder logic (via `api.swapStageOrder`)
- Per-job `stages` do NOT live here: they are local to JobDetailDialog (loaded per opened job).

### `composables/useMeetings.js`
- `upcomingMeetings` ref, `loadUpcomingMeetings`
- Per-job meetings are local to JobDetailDialog.

### `composables/useDarkMode.js`
- `dark` ref, `applyDark`, `toggleDark`, plus an `initDarkMode()` for the `localStorage`/`prefers-color-scheme` bootstrap currently in `onMounted`.

### `components/BaseDialog.vue`
- The repeated overlay shell: `fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4` + white/dark panel. Props: `width` (Tailwind class, e.g. `max-w-md`), optional `z` for the stage-comment `z-[60]`. Default slot for content, emits `close`.
- Owns Esc handling: a module-level stack of open dialogs; on mount push `emit('close')`, on unmount pop; a single `keydown` listener closes the top of the stack. This replaces the `onEsc` priority chain for dialogs (StageCommentDialog mounts after JobDetailDialog, so last-opened-wins for free). The two non-dialog dropdowns (stage filter, upcoming meetings) keep tiny local Esc handlers in their components to preserve current behavior.

### `components/ConfirmDialog.vue`
- Replaces both the Delete Confirm and Archive Confirm dialogs. Props: `title`, `message`, `confirmLabel`, `tone` (`red`/`amber`). Emits `confirm`, `close`. Dumb: no API calls.

### `components/StageLogList.vue`
- The stage-history `<ul>` markup (prev_stage → stage, `formatDate(log.created_at)`, notes), currently duplicated in the Stage Update Dialog and the Detail Dialog. Prop: `logs`. Dumb.

### `components/StageListEditor.vue`
- The draggable stage list + add form, currently duplicated in "Manage Stages" and "Default Stages" dialogs. Owns local drag state (`dragIdx`/`dragOverIdx`, was also `dragDefaultIdx`/`dragOverDefaultIdx`). Props: `stages`. Emits: `add(name)`, `rename(stage)` (was `updateStage` on blur), `remove(id)`, `reorder(fromIdx, toIdx)`. Dumb: parents perform the API calls.

### `components/AppHeader.vue`
- Header markup: title, Export CSV link, upcoming-meetings dropdown (`upcomingMeetingsOpen`, `openMeetingFromDropdown`, uses `useMeetings` + `formatDate` + `isUrgent`), Default Stages button, dark toggle (uses `useDarkMode`).
- Owns `defaultStagesMgmt` and renders `DefaultStagesDialog` itself (self-contained). Emits only `open-job(jobId)` (was `openMeetingFromDropdown` finding the job; App resolves the id against `jobs` and opens the detail dialog).

### `components/DefaultStagesDialog.vue`
- BaseDialog + StageListEditor wired to `useStages`: `addDefaultStage`, `updateStage`, `removeDefaultStage`, `dropDefaultStage` handlers. Includes the "Copied to every new job on creation" copy. Emits `close`.

### `components/JobFilters.vue`
- The whole Filters panel: collapse toggle, Archived only / Active only / Top matches chips, text input, status chips (chip drag select), stage multiselect dropdown with search, applied date range, Clear. Consumes `useJobFilters()` directly (shared state, no prop drilling of ~10 v-models). Registers the window `mouseup` listener.

### `components/JobsTable.vue`
- Table markup: header row, inline add-row (`form` ref, `emptyForm`, `companyInput` ref + focus, `save` create-path, `reset`), empty-state row, job rows (`onRowDblClick`, star `toggleTopMatch`, status pill, stage progress bar via local `stageProgress(job)`, `formatDay`, `truncateNotes`, view/archive/unarchive/delete buttons).
- Owns `confirmDelete`/`confirmArchive` state and renders two `ConfirmDialog`s; on confirm calls `useJobs().removeJob` / `setArchived` (was `remove`/`doDelete`/`archive`/`doArchive`/`unarchive`).
- Props: `jobs` (the filtered list), `filterText` (for `highlight`), `totalCount` (for the empty-state message distinction). Emits: `view(job)`.
- `stageProgress` stays local here (only consumer).

### `components/JobDetailDialog.vue`
- Prop: `job`. Emits: `close`, `saved`. On mount loads its own data (the `Promise.all` from `openDetail`): `contacts`, `logs`, `stages` (per-job), `meetings`, and builds `edit` (using `isoToDate` for applied_at).
- Contains: the edit form (company/position/status/stage select + gear/applied/notes/url), top-match star, `StageLogList`, the Contacts section (`newContact`, `pendingContacts`, `deletedContactIds`, `addContact`, `removeContact`), the Meetings section (`newMeeting`, `editingMeeting`, `sortedDialogMeetings`, `addMeeting`, `editMeeting`, `cancelEditMeeting`, `saveMeetingEdit`, `removeMeeting`, `refreshDetailMeetings` which also calls `useMeetings().loadUpcomingMeetings`), `saveDetail`, `confirmStageComment` logic.
- Nests two dialogs it alone triggers:
  - Manage Stages (was `stagesMgmt`): BaseDialog + StageListEditor wired to the per-job stage endpoints (`addStage`, `updateStage`, `removeStage`, `dropStage` via `api.swapStageOrder`). Inlined here, not a separate file: it is 15 lines of wiring.
  - `StageCommentDialog` (see below), opened from `saveDetail` when `stage_id` changed.
- On successful save emits `saved`; App calls `loadJobs()`.

### `components/StageCommentDialog.vue`
- Props: `isLastStage`, `initialStatus`. Local `notes` and `newStatus`. Emits `confirm({ notes, newStatus })`, `close`. JobDetailDialog performs the API fan-out (`addLog` + `updateJob` + pending/deleted contacts), which is today's `confirmStageComment`.

### `App.vue` (after)
- Template: `AppHeader @open-job` → `JobFilters` → `JobsTable :jobs="filteredJobs" :filter-text="filter.text" :total-count="jobs.length" @view` → `JobDetailDialog v-if="detailJob" :job="detailJob" @close @saved`.
- Script: `detailJob` ref, `onMounted` bootstrap (`initDarkMode()`, `loadJobs()`, `loadDefaultStages()`, `loadUpcomingMeetings()`), open-job resolution for the header event.

### Deleted (dead code)
`stageDialog`, `openStageDialog`, `submitStageUpdate`, `nextStageId`, and the "Stage Update Dialog" template block (lines 514-595) have no trigger anywhere in the template. Delete them during the migration; git history keeps them, and if the flow returns it is a ~40-line component on top of BaseDialog + StageLogList.

## 3. Component hierarchy and contracts

```
App.vue
├── AppHeader                     emits: open-job(jobId)
│   └── DefaultStagesDialog       emits: close
│       └── BaseDialog + StageListEditor
├── JobFilters                    (no props/emits: consumes useJobFilters)
├── JobsTable                     props: jobs, filterText, totalCount   emits: view(job)
│   └── ConfirmDialog ×2          props: title, message, confirmLabel, tone   emits: confirm, close
└── JobDetailDialog (v-if job)    props: job   emits: close, saved
    ├── BaseDialog                props: width, z   emits: close
    ├── StageLogList              props: logs
    ├── [inline] Manage Stages    BaseDialog + StageListEditor
    │   └── StageListEditor       props: stages   emits: add(name), rename(stage), remove(id), reorder(from, to)
    └── StageCommentDialog        props: isLastStage, initialStatus   emits: confirm({notes,newStatus}), close
```

Rules:
- Dumb components (BaseDialog, ConfirmDialog, StageLogList, StageListEditor): props/emits only, no composables, no api.js imports.
- Feature components (AppHeader, JobFilters, JobsTable, JobDetailDialog, DefaultStagesDialog, StageCommentDialog) may import composables, api.js, utils, constants.
- Composables never import components.

## 4. Shared state

Mechanism: **module-scope refs inside composables** (singleton by ES module semantics). Example shape:

```js
// composables/useJobs.js
const jobs = ref([])            // module scope = shared
export function useJobs() {
  async function loadJobs() { jobs.value = await api.fetchJobs() }
  ...
  return { jobs, loadJobs, removeJob, setArchived, toggleTopMatch }
}
```

| State | Owner | Consumers |
|---|---|---|
| `jobs` | useJobs | App (bootstrap, open-job lookup), JobsTable (via App prop), useJobFilters |
| `filter`, `archivedOnly`, `topMatchOnly`, `filteredJobs`, dropdown state | useJobFilters | JobFilters, App (passes `filteredJobs`/`filter.text` to JobsTable) |
| `defaultStages` | useStages | DefaultStagesDialog, useJobFilters (`allFilterStages`) |
| `upcomingMeetings` | useMeetings | AppHeader, JobDetailDialog (`refreshDetailMeetings`) |
| `dark` | useDarkMode | AppHeader, App (init) |
| `form` (add row), confirm dialogs | JobsTable local | — |
| `contacts`, `logs`, per-job `stages`, `meetings`, `edit`, pending/deleted contacts, meeting forms | JobDetailDialog local | its nested dialogs |

Constraint carried over from today's code: composables holding module-scope state must not register lifecycle hooks or window listeners (they are called from multiple components); listeners belong to exactly one component (`mouseup` → JobFilters, Esc → BaseDialog's single module-level listener).

## 5. Decisions

- **No router.** One view; dialogs are transient UI state, not navigable locations. Deep-linking a job detail would be the first reason to add vue-router; nothing today asks for it.
- **No Pinia.** Shared state is four refs with trivial mutations. Module-scope refs in composables give identical singleton semantics with zero dependencies and zero boilerplate. Pinia earns its keep with devtools/time-travel needs or many cross-cutting stores; not here.
- **Modals: `v-if` + BaseDialog, parent owns visibility.** Same pattern as today, minus 6 copies of the overlay markup. Ownership follows the trigger: App owns the detail dialog, JobsTable owns its confirms, AppHeader owns default stages, JobDetailDialog owns its two nested dialogs. No global modal manager, no Teleport needed (overlays are already viewport-fixed and stacking works via z-index as today).
- **CSS: stays Tailwind-in-template, `style.css` stays global.** It is only the Tailwind entry + dark variant. No scoped styles are needed for this split; if a component ever needs real CSS, it goes in a `<style scoped>` block in that component, never in style.css.
- **Data refetch stays "mutate then reload"** (`load()` after each write), as today. It is correct and simple at this data size; no client cache/invalidation layer.
- **Dead Stage Update Dialog is deleted**, not migrated (see section 2).
- **Not extracted (deliberately):** SVG icons (the trash icon repeats across 3 files after the split; an icon component is optional follow-up, not part of this migration), a JobRow component (rows stay in JobsTable), a toast/error layer (no error handling exists today; adding one is a feature, not a refactor).

## 6. Migration order

Each step compiles, lints (`cd web && npm run lint`), and works on its own; commit per step.

1. **`utils/dates.js`, `utils/text.js`, `constants.js`** — pure moves, App.vue imports them. Date helpers verbatim with comments (CLAUDE.md rule).
2. **`api.js`** — replace all ~30 inline `fetch` calls; `jobBody` moves here. Behavior-identical.
3. **Composables** — `useDarkMode`, `useMeetings`, `useStages`, `useJobs`, `useJobFilters` (in that order: least to most entangled). App.vue still renders everything but its script shrinks to wiring. Delete the dead stage-dialog code here.
4. **Dumb components** — `BaseDialog` (with the Esc stack, remove `onEsc`), `ConfirmDialog`, `StageLogList`, `StageListEditor`. Template blocks in App.vue swap to them.
5. **Dialogs** — `StageCommentDialog`, `DefaultStagesDialog`, then `JobDetailDialog` (the big one: detail data loading moves from `openDetail` into its `onMounted`).
6. **Layout components** — `AppHeader`, `JobFilters`, `JobsTable` (with its confirm dialogs).
7. **Slim App.vue** — reduce to shell + bootstrap; final lint + manual smoke test (add job, filter, detail edit with stage change + comment, contacts pending/delete, meetings CRUD, stage reorder both dialogs, archive/unarchive, delete, dark toggle, Esc closes topmost, `applied_at` renders the same wall date).