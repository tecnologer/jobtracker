# Requirements: Job Application Progression Dashboard

## 1. Context and problem

jobtracker is a single-user job application tracker (Go REST API, SQLite, Vue 3 SPA behind basic auth). It already captures everything needed for progression analysis: jobs with status and applied date, per-job interview stages, and `StageLog` rows recording every stage transition with a timestamp. What it lacks is any aggregate view: the user can only see one job at a time, so questions like "where do my applications die", "how long does each stage take", and "what is my rejection rate" require manual counting or the CSV export plus a spreadsheet.

The dashboard answers the core question at a glance: how far do applications get through the funnel, and where do they stall or die. It is a read-only analytics view over existing data; it introduces no new user-entered data.

Design constraint carried from the codebase philosophy: minimal, YAGNI-driven. One new backend endpoint, zero new frontend dependencies, no router.

## 2. Actors

| Actor | Description | Volume |
|---|---|---|
| Job seeker (owner) | Sole user of the tool, technical, authenticated via basic auth | 1 user, dataset of one person's job search (tens to low hundreds of jobs, hundreds of stage logs) |

## 3. v1 scope / No-goals

**In scope (v1):**
- KPI cards, status breakdown, stage funnel, average time per stage (see FRs).
- One aggregate endpoint `GET /api/stats`.
- Dashboard as a toggled view inside the existing SPA.

**No-goals (explicitly out of v1):**
- Time range filters (all-time only).
- Time series charts (applications per week, activity trends).
- Chart library; period-over-period comparisons; goal setting.
- Meeting/interview analytics.
- Dashboard export (CSV export already exists for raw data).
- Router or state library.
- Any write operation from the dashboard.

## 4. Functional requirements

- **FR-01: Dashboard view toggle.** A control in the header switches the main area between the jobs table and the dashboard, within the existing SPA (no router). Acceptance: clicking the toggle shows the dashboard; clicking again returns to the table; page refresh returns to the table view (no persistence required).

- **FR-02: Stats endpoint.** `GET /api/stats` returns all dashboard aggregates in a single JSON payload, computed server-side (Go/SQL). No query parameters in v1. Acceptance: one request supplies every number the dashboard renders; the dashboard makes no per-job calls.

- **FR-03: KPI cards.** The dashboard shows: total jobs, active jobs (status not in {rejected, canceled} and not archived), offers (jobs whose logs include reaching the "Offer" stage, plus status `accepted`), rejection rate (rejected / total with non-prospect status), and average days from applied to first response. Acceptance: values match a manual count against the database for a seeded dataset.

- **FR-04: Status breakdown.** Count of jobs per status (all 8 statuses), rendered as labeled horizontal bars using the existing status colors. Includes archived and soft-active jobs; excludes soft-deleted. Acceptance: per-status counts sum to total jobs and match the table's filter counts.

- **FR-05: Stage funnel.** For each default stage (ordered by `sort_order`), the number of distinct jobs that ever reached it, derived from `StageLog` entries grouped by stage name (per-job stages are name-matched clones of the templates). Archived and closed jobs are included. Acceptance: a job with logs Phone Screen → Technical Interview counts once in each of those two bars; a job with no logs counts in none.

- **FR-06: "Other" stage bucket.** Stage log entries whose stage name does not match any current default stage name are aggregated into a single "Other" funnel row, placed after the default stages. Acceptance: a job whose custom stage "VP Chat" was logged appears in "Other", not dropped.

- **FR-07: Average time per stage.** For each default stage, the average days between entering the stage (log where `stage_id` names it) and leaving it (next log for the same job). Stages a job never left (its current stage) contribute time up to now. Acceptance: for a seeded job that spent exactly 3 and 5 days in a stage across two jobs, the stage shows 4 days.

- **FR-08: Definitions applied consistently.**
  - "Response" = the job's first `StageLog` of any kind. Average time to first response = days from `applied_at` to that log's `created_at`.
  - A rejection with zero stage logs counts as "no interview" and does not count as a response.
  - Jobs with `applied_at` null (typically prospects) are excluded from all time-based metrics but included in the status breakdown.
  Acceptance: a prospect with no applied date never affects avg-days KPIs; its status still shows in FR-04.

- **FR-09: Empty and sparse data.** With zero jobs, or jobs but zero stage logs, the dashboard renders without errors, showing zeros and an empty funnel rather than NaN or a crash. Acceptance: fresh database renders the dashboard cleanly.

## 5. Non-functional requirements

- **NFR-01: Latency.** `GET /api/stats` responds in under 200 ms with 500 jobs and 5,000 stage logs on the production host (SQLite, single connection). Verify with a seeded database and `curl -w '%{time_total}'`.
- **NFR-02: No new dependencies.** Zero new Go modules and zero new npm packages; charts are plain CSS/HTML bars styled with the existing Tailwind setup. Verify: `go.mod` and `package.json` diffs are empty.
- **NFR-03: Auth parity.** `/api/stats` sits behind the same basic auth middleware as every other route. Verify: unauthenticated request returns 401.
- **NFR-04: Read-only.** The endpoint performs no writes; safe to call concurrently with edits (single-connection SQLite already serializes). Verify: endpoint handler contains only SELECTs.
- **NFR-05: Frontend conventions.** New `.vue` files pass `npm run lint` (eslint-plugin-vue flat/recommended), dark mode supported via existing Tailwind dark classes. Verify: lint passes, dashboard is legible in dark mode.

## 6. Constraints

- Stack is fixed: Go stdlib `net/http` routing (Go 1.22 method patterns), GORM + glebarez/sqlite, Vue 3 SPA served from `web/dist`, Tailwind.
- No router or state library (existing architectural decision).
- Stage identity across jobs exists only by name: per-job stages are clones of `job_id = 0` templates, so all cross-job stage aggregation must group by stage name.
- Stage logs are currently only queryable per job; cross-job aggregation must be a new store method, not N+1 API calls.
- `applied_at` is a wall date, not an instant: day-difference math must use its `YYYY-MM-DD` component (see CLAUDE.md date rules), while `StageLog.created_at` is a real instant.
- Deployment: single binary on Railway, `healthz` unauthenticated, everything else behind basic auth.

## 7. Assumptions and risks

All items below were defaulted by the analyst and accepted by the client with "defaults are fine"; flagged here because none was independently specified.

- **A-01:** Funnel plus KPIs is the core need; time series deferred. Risk: if weekly activity tracking turns out to be the real question, v1 will not answer it.
- **A-02:** The four widgets in scope (KPIs, status breakdown, funnel, time per stage) are sufficient; nothing else was requested.
- **A-03:** "Response" defined as first stage log. Risk: a rejection email with no logged stage looks like "no response"; acceptable for a personal tool where the user controls logging discipline.
- **A-04:** Renamed/custom stages fragment into "Other". Risk: heavy per-job stage customization makes the funnel less informative; mitigated by the user mostly keeping default stage names.
- **A-05:** All-time window is enough at this data volume.
- **A-06:** Deleted default stages: logs referencing a stage name no longer in the templates fall into "Other". This changes historical funnel shape when templates are edited; accepted.
- **A-07:** Stage logs with null `stage_id` (stage cleared) end the previous stage's duration but add no funnel entry.
- **A-08:** Time-per-stage for a job's current stage counts up to "now", so the number moves between page loads; accepted as more honest than excluding in-progress stages.

## 8. Open questions

- None blocking. One deferred decision: whether the "offers" KPI should also require status `accepted`/`negotiating` or purely the Offer stage log; v1 uses "reached Offer stage OR status accepted" (FR-03) and can be tightened after real use.
