# Backend Architecture: Dashboard Stats Endpoint

Scope: `GET /api/stats` (FR-02), serving KPI cards (FR-03), status breakdown (FR-04), stage funnel with "Other" bucket (FR-05, FR-06), and average time per stage (FR-07). Read-only, no new dependencies, no schema changes.

## 1. Overview

One new route in `handler/routes.go`, one thin handler in `handler/`, one new store method `Store.Stats(now time.Time)` in a new file `store/stats.go`. The store method issues three read-only SELECTs (jobs, default stages, stage logs joined to stage names) and computes every aggregate in a single Go pass over the log rows.

```
GET /api/stats
  └─ handler.Stats            (thin: call store, writeJSON)
       └─ store.Stats(now)    (3 SELECTs + one in-memory pass)
            ├─ SELECT jobs (id, status, applied_at, archived_at)
            ├─ SELECT default stages (reuse ListDefaultStages)
            └─ SELECT logs JOIN jobs (filter soft-deleted)
                            LEFT JOIN stages (stage name)
```

## 2. Architecture drivers

- **NFR-01**: <200 ms at 500 jobs / 5,000 logs. At this scale, everything fits in memory trivially; three indexed-free full scans of tiny tables cost single-digit milliseconds.
- **NFR-02/NFR-04**: no new dependencies, SELECTs only.
- **Constraint (REQUIREMENTS §6)**: stage identity across jobs is name-only; cross-job aggregation must be a new store method.
- **Constraint (REQUIREMENTS §6, CLAUDE.md)**: `applied_at` is a wall date; `StageLog.created_at` is an instant. Day math must respect that split.
- **Codebase convention**: store owns all query and aggregation logic (see `NextMeetingTimes`), handlers stay thin.

## 3. Response shape

`GET /api/stats` → `200 application/json`:

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

Semantics:

| Field | Definition | Source |
|---|---|---|
| `total_jobs` | all non-soft-deleted jobs | FR-03 |
| `active_jobs` | status not in {rejected, canceled} AND `archived_at IS NULL` | FR-03 |
| `offers` | distinct jobs with a log whose stage name is exactly `"Offer"`, union jobs with status `accepted` | FR-03, §8 |
| `rejection_rate` | rejected / count(status != prospect); `0` when denominator is 0 | FR-03, FR-09 |
| `avg_days_to_first_response` | mean over responded jobs of wall-date diff (`applied_at` date → first log's `created_at` date, each in its own stored offset); `null` when no job qualifies | FR-08 |
| `status_breakdown` | map with all 8 statuses always present (zeros included); frontend owns ordering and colors | FR-04 |
| `funnel` | default stages in `sort_order`, always all present even at zero; `"Other"` row appended only when `jobs_reached > 0`, with `sort_order` = last default + 1 | FR-05, FR-06, FR-09 |
| `funnel[].jobs_reached` | distinct jobs with ≥1 log naming that bucket (a job counts once per bucket regardless of repeat visits) | FR-05 |
| `funnel[].avg_days` | mean stint length in days (float, from instants); `null` when the bucket has no time samples | FR-07 |

Numbers are raw floats; rounding is the frontend's job (one formatting rule, not two).

## 4. Decisions

**AD-01: Aggregate in Go, not SQL.** The store loads three result sets and computes everything in one pass.
- Rationale: the hard parts (LEAD-style "next log per job" pairing, wall-date vs instant day math, name-vs-template bucketing, distinct-per-job funnel counts) are window functions plus string date arithmetic in SQL, but ~40 lines of obvious Go. The glebarez driver returns aggregate datetime columns as raw text (see `NextMeetingTimes`), so SQL date math would reintroduce fragile string parsing. At 5k rows, Go-side aggregation is well under NFR-01 with two orders of magnitude of headroom.
- Rejected: pure-SQL aggregation (CTEs + `LEAD() OVER`), correct but unreadable and fights the driver's text-typed aggregates; per-widget SQL queries (5+ round trips on a single-connection DB for no benefit).
- Consequence: if the dataset ever grows 100x, this method is the one to push into SQL. It is self-contained in `store/stats.go`, so that swap touches one file.

**AD-02: One store method, `Stats(now time.Time) (*Stats, error)`.** Not separate methods per widget.
- Rationale: FR-02 mandates one payload; the widgets share the same log scan, so one pass computes all of them. `now` is a parameter so FR-07's "current stage counts up to now" is deterministic in tests (matches the seeded 3-day/5-day acceptance case).
- Rejected: `ListAllStageLogs()` + aggregation in the handler (violates thin-handler convention; `NextMeetingTimes` already sets the precedent that computation lives in store).

**AD-03: New file `store/stats.go` (same package), not a new package.**
- Rationale: it needs the models and `s.db`; a separate package would force exporting internals. `store.go` is already ~370 lines; a sibling file keeps the feature deletable in one `rm`.

**AD-04: Exact, case-sensitive name match for funnel buckets.** Per-job stages are byte-for-byte clones of templates, so exact match is correct by construction; any rename (template or per-job) diverges into "Other", exactly as A-04/A-06 accept.
- Rejected: fuzzy/case-insensitive matching (invents equivalences the requirements did not ask for).

**AD-05: Logs with no resolvable stage name (null `stage_id`, or `stage_id` pointing at a hard-deleted stage row) end the previous stint and add no funnel entry.** A-07 defines this for null `stage_id`; a dangling FK after `DeleteStage` (hard delete, `Stage` has no soft delete) is indistinguishable after the LEFT JOIN, so it gets the same treatment. Flagged as an assumption: FR-06's "Other" applies to logs with a *name* that matches no template, not to logs with no name at all.

**AD-06: FR-07 applied literally: every job's last log accrues time up to `now`, including closed (rejected/canceled/accepted) jobs.** A-08 accepts the moving number. Consequence (negative): a rejected job's final stage grows without bound and will slowly inflate that stage's average. If that skews the widget in practice, the one-line fix is to cap the open stint at `updated_at` or exclude terminal-status jobs; noted in §9, not built now.

**AD-07: `status_breakdown` is a JSON object keyed by status, all 8 keys always present.** The frontend already owns status order and colors; an ordered array would duplicate that knowledge server-side.

## 5. Store: types and method

In `store/stats.go`:

```go
type StageStat struct {
    Name        string   `json:"name"`
    SortOrder   int      `json:"sort_order"`
    JobsReached int      `json:"jobs_reached"`
    AvgDays     *float64 `json:"avg_days"` // nil = no time samples for this bucket
}

type Stats struct {
    TotalJobs              int                       `json:"total_jobs"`
    ActiveJobs             int                       `json:"active_jobs"`
    Offers                 int                       `json:"offers"`
    RejectionRate          float64                   `json:"rejection_rate"`
    AvgDaysToFirstResponse *float64                  `json:"avg_days_to_first_response"` // nil = no responses yet
    StatusBreakdown        map[ApplicationStatus]int `json:"status_breakdown"`
    Funnel                 []StageStat               `json:"funnel"`
}

// Stats computes all dashboard aggregates in one pass. Read-only.
// `now` closes the open stint of each job's current stage (FR-07 / A-08).
func (s *Store) Stats(now time.Time) (*Stats, error)
```

### Queries (all SELECT, NFR-04)

1. **Jobs**: `s.db.Model(&Job{}).Select("id", "status", "applied_at", "archived_at").Find(&jobs)`. GORM's soft-delete scope excludes `deleted_at` rows automatically.
2. **Default stages**: reuse `s.ListDefaultStages()` (order and names for the funnel skeleton).
3. **Logs with names**, ordered for the single pass:

```go
type statLogRow struct {
    JobID     uint
    StageName *string   // nil: null stage_id or dangling stage row (AD-05)
    CreatedAt time.Time // plain column, scans as time.Time (unlike aggregates)
}

s.db.Model(&StageLog{}).
    Select("stage_logs.job_id, stages.name AS stage_name, stage_logs.created_at").
    Joins("JOIN jobs ON jobs.id = stage_logs.job_id AND jobs.deleted_at IS NULL").
    Joins("LEFT JOIN stages ON stages.id = stage_logs.stage_id").
    Order("stage_logs.job_id, stage_logs.created_at").
    Find(&rows)
```

### Aggregation pass

Precompute: `defaultNames` set from query 2; `bucket(name) = name` if in set, else `"Other"`; per-job map of `applied_at` and status from query 1 (KPIs and breakdown fall straight out of that slice).

Then iterate `rows`, which arrive grouped by job and ordered by time. For each job's run of logs:

- **Funnel** (`jobs_reached`): for each log with a non-nil name, mark `seen[bucket][jobID]`; distinct by construction. Jobs with nil `applied_at` ARE included (FR-08 excludes them from time math only).
- **Time per stage** (`avg_days`): stint for log *i* runs from `rows[i].CreatedAt` to `rows[i+1].CreatedAt` (next log of the same job, any name including nil), or to `now` for the job's last log. Attribute `hours/24` to `bucket(rows[i].StageName)`; skip logs with nil name (AD-05) and skip entirely if the job's `applied_at` is nil (FR-08). `"Other"` gets an average too, it falls out of the same loop.
- **First response**: the first log of each job with non-nil `applied_at` yields one sample: calendar-day difference between `applied_at.Format("2006-01-02")` and `CreatedAt.Format("2006-01-02")`, each rendered in its own stored offset (the CLAUDE.md wall-date rule; never `In(location)`-shift `applied_at`). Rejections with zero logs contribute nothing (FR-08).

Assemble: funnel = default stages in `sort_order` (always emitted, zeros for empty, FR-09), then `"Other"` if reached. All divisions guard `count > 0`; empty averages are `nil`, never NaN (FR-09).

## 6. Handler and route

`handler/handler.go`, matching existing handlers exactly:

```go
func (h *Handler) Stats(w http.ResponseWriter, _ *http.Request) {
    stats, err := h.store.Stats(time.Now())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    writeJSON(w, http.StatusOK, stats)
}
```

`handler/routes.go`, one line among the other GETs:

```go
mux.HandleFunc("GET /api/stats", h.Stats)
```

The route lives inside the mux that `main.go` wraps with `basicAuth`, so NFR-03 (401 unauthenticated) holds with zero new code.

## 7. File map

| File | Change |
|---|---|
| `store/stats.go` | new: `Stats`, `StageStat`, `Store.Stats`, aggregation helpers |
| `store/stats_test.go` | new: store-level tests (below) |
| `handler/handler.go` | +1 handler method |
| `handler/routes.go` | +1 route |
| `CLAUDE.md` | update the architecture tree/API list to mention `/api/stats` and `store/stats.go` |

No migrations, no model changes, no `main.go` changes.

## 8. Test plan (`store/stats_test.go`)

Conventions from `seed_test.go` and repo rules: `t.Parallel()`, `filepath.Join(t.TempDir(), "jobs.db")` + `Open`, `testify/require`, table-driven with `t.Run(test.name, ...)` and snake_case scenario names, helpers last with `t.Helper()`. Seeding goes through the public API (`Create`, `AddStageLog`, `CreateStage`); deterministic timestamps by setting `StageLog.CreatedAt` before `AddStageLog` (GORM keeps a non-zero `CreatedAt`) and passing a fixed `now` to `Stats`.

- **TestStats_Empty** (FR-09): fresh DB → all KPIs zero, `AvgDaysToFirstResponse == nil`, breakdown has all 8 keys at 0, funnel has exactly the 5 default stages with `JobsReached == 0` and `AvgDays == nil`, no "Other".
- **TestStats_KPIs** (FR-03): seed jobs across statuses plus one archived; assert `total`, `active` (archived and rejected/canceled excluded), `offers` counts a job with an "Offer"-named log AND a status-`accepted` job with no such log, without double counting a job that is both; `rejection_rate` = rejected / non-prospect.
- **TestStats_StatusBreakdown** (FR-04): counts per status sum to `total_jobs`; a soft-deleted job (`Delete`) appears nowhere, including the funnel (its logs are filtered by the join).
- **TestStats_Funnel** (FR-05, FR-06, AD-04, AD-05): table-driven — `two_stage_job_counts_once_per_bar`, `no_logs_counts_nowhere`, `custom_stage_goes_to_other` (via `CreateStage` "VP Chat" + log), `revisited_stage_still_counts_once`, `null_stage_log_adds_no_entry`, `other_row_absent_when_unused`.
- **TestStats_AvgTimePerStage** (FR-07, A-08): the acceptance case, two jobs spending exactly 3 and 5 days in "Phone Screen" (closed by a following log) → `AvgDays == 4.0`; plus `current_stage_counts_to_now`: last log at `now.Add(-48h)` → 2.0.
- **TestStats_FirstResponseAndNullAppliedAt** (FR-08): mean of wall-date diffs for responded jobs; a rejected job with zero logs is not a response; a nil-`applied_at` job with logs is excluded from `avg_days_to_first_response` and from every `AvgDays`, yet still counts in the funnel and breakdown.

Handler gets no dedicated test: it is four lines of the same shape as 20 existing untested handlers; store tests cover all logic. NFR-01 is verified manually per the requirement's own method (`curl -w '%{time_total}'` against a seeded DB), not with a CI benchmark.

## 9. Evolution notes

- **If A-08's moving averages mislead** (terminal jobs inflating their last stage forever, AD-06): cap the open stint at the job's `updated_at` for terminal statuses. One conditional inside the stint loop; payload unchanged.
- **If data outgrows in-memory aggregation** (AD-01, roughly >50k logs): rewrite `Store.Stats` internals as SQL window functions. Signature, payload, and tests stay; only `store/stats.go` internals change.
