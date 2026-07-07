-- Convert legacy bare-date applied_at values ("YYYY-MM-DD") into timezone-aware
-- RFC3339 timestamps at local midnight ("YYYY-MM-DDT00:00:00-06:00").
--
-- Context: applied_at used to be stored as a bare date string. It is now a
-- timezone-aware timestamp (see store.Job.AppliedAt *time.Time). New rows are
-- written by the app as RFC3339 with an offset, e.g. "2026-07-05T00:00:00-06:00".
-- This backfills old rows to match, treating the stored day as *local* midnight.
--
-- applied_at is a calendar date: the app renders the wall-clock date straight off
-- the string (App.vue isoToDate/formatDay), so the day never shifts for the viewer.
-- The offset below is cosmetic for display but keeps stored instants consistent with
-- fresh creates. It is hardcoded to this deployment's timezone (UTC-6); change it if
-- you run elsewhere.
--
-- Safe to run more than once: it only touches bare 10-char dates and skips rows
-- already carrying a time component. Back up jobs.db first (e.g. `cp jobs.db jobs.db.bak`).

BEGIN TRANSACTION;

-- Normalize empty strings to NULL (unset).
UPDATE jobs
SET applied_at = NULL
WHERE applied_at = '';

-- Bare "YYYY-MM-DD" -> "YYYY-MM-DDT00:00:00-06:00" (local midnight, UTC-6).
UPDATE jobs
SET applied_at = substr(applied_at, 1, 10) || 'T00:00:00-06:00'
WHERE applied_at IS NOT NULL
  AND length(applied_at) = 10
  AND applied_at NOT LIKE '%T%';

COMMIT;

-- Verify:
--   SELECT id, applied_at FROM jobs;
