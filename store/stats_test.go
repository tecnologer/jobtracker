package store

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var statsNow = time.Date(2026, 1, 20, 12, 0, 0, 0, time.UTC)

func TestStatsEmpty(t *testing.T) {
	t.Parallel()

	s := newStatsStore(t)

	stats, err := s.Stats(statsNow)
	require.NoError(t, err)

	require.Zero(t, stats.TotalJobs)
	require.Zero(t, stats.ActiveJobs)
	require.Zero(t, stats.Offers)
	require.Zero(t, stats.RejectionRate)
	require.Nil(t, stats.AvgDaysToFirstResponse)
	require.Len(t, stats.StatusBreakdown, 8)
	for status, count := range stats.StatusBreakdown {
		require.Zero(t, count, "status %s", status)
	}
	require.Len(t, stats.Funnel, 5, "default stages only, no Other")
	for _, stage := range stats.Funnel {
		require.Zero(t, stage.JobsReached, "stage %s", stage.Name)
		require.Nil(t, stage.AvgDays, "stage %s", stage.Name)
	}
}

func TestStatsKPIs(t *testing.T) {
	t.Parallel()

	s := newStatsStore(t)
	applied := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)

	seedStatsJob(t, s, StatusApplied, &applied, false)
	seedStatsJob(t, s, StatusRejected, &applied, false)
	seedStatsJob(t, s, StatusCanceled, &applied, false)
	accepted := seedStatsJob(t, s, StatusAccepted, &applied, false)
	offered := seedStatsJob(t, s, StatusApplied, &applied, false)
	seedStatsJob(t, s, StatusApplied, &applied, true) // archived
	seedStatsJob(t, s, StatusProspect, nil, false)

	// accepted job also reached Offer: must not be double-counted
	addStatsLog(t, s, accepted, stageIDByName(t, s, accepted, "Offer"), statsNow.AddDate(0, 0, -5))
	addStatsLog(t, s, offered, stageIDByName(t, s, offered, "Offer"), statsNow.AddDate(0, 0, -3))

	stats, err := s.Stats(statsNow)
	require.NoError(t, err)

	require.Equal(t, 7, stats.TotalJobs)
	require.Equal(t, 4, stats.ActiveJobs, "archived, rejected, canceled excluded")
	require.Equal(t, 2, stats.Offers, "offer-log job + accepted job, no double count")
	require.InDelta(t, 1.0/6.0, stats.RejectionRate, 1e-9, "rejected / non-prospect")
}

func TestStatsStatusBreakdown(t *testing.T) {
	t.Parallel()

	s := newStatsStore(t)
	applied := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)

	seedStatsJob(t, s, StatusApplied, &applied, false)
	seedStatsJob(t, s, StatusApplied, &applied, false)
	seedStatsJob(t, s, StatusInProgress, &applied, false)
	seedStatsJob(t, s, StatusProspect, nil, false)

	deleted := seedStatsJob(t, s, StatusApplied, &applied, false)
	addStatsLog(t, s, deleted, stageIDByName(t, s, deleted, "Phone Screen"), statsNow.AddDate(0, 0, -2))
	require.NoError(t, s.Delete(deleted))

	stats, err := s.Stats(statsNow)
	require.NoError(t, err)

	require.Equal(t, 4, stats.TotalJobs, "soft-deleted job excluded")
	sum := 0
	for _, count := range stats.StatusBreakdown {
		sum += count
	}
	require.Equal(t, stats.TotalJobs, sum)
	require.Zero(t, funnelStage(t, stats, "Phone Screen").JobsReached, "deleted job's logs excluded from funnel")
}

func TestStatsFunnel(t *testing.T) {
	t.Parallel()

	applied := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name  string
		seed  func(t *testing.T, s *Store)
		check func(t *testing.T, stats *Stats)
	}{
		{
			name: "two_stage_job_counts_once_per_bar",
			seed: func(t *testing.T, s *Store) {
				t.Helper()
				jobID := seedStatsJob(t, s, StatusInProgress, &applied, false)
				addStatsLog(t, s, jobID, stageIDByName(t, s, jobID, "Phone Screen"), statsNow.AddDate(0, 0, -6))
				addStatsLog(t, s, jobID, stageIDByName(t, s, jobID, "Technical Interview"), statsNow.AddDate(0, 0, -3))
			},
			check: func(t *testing.T, stats *Stats) {
				t.Helper()
				require.Equal(t, 1, funnelStage(t, stats, "Phone Screen").JobsReached)
				require.Equal(t, 1, funnelStage(t, stats, "Technical Interview").JobsReached)
				require.Zero(t, funnelStage(t, stats, "Offer").JobsReached)
			},
		},
		{
			name: "no_logs_counts_nowhere",
			seed: func(t *testing.T, s *Store) {
				t.Helper()
				seedStatsJob(t, s, StatusApplied, &applied, false)
			},
			check: func(t *testing.T, stats *Stats) {
				t.Helper()
				for _, stage := range stats.Funnel {
					require.Zero(t, stage.JobsReached, "stage %s", stage.Name)
				}
			},
		},
		{
			name: "custom_stage_goes_to_other",
			seed: func(t *testing.T, s *Store) {
				t.Helper()
				jobID := seedStatsJob(t, s, StatusInProgress, &applied, false)
				custom := &Stage{JobID: jobID, Name: "VP Chat", SortOrder: 99}
				require.NoError(t, s.CreateStage(custom))
				addStatsLog(t, s, jobID, custom.ID, statsNow.AddDate(0, 0, -2))
			},
			check: func(t *testing.T, stats *Stats) {
				t.Helper()
				require.Len(t, stats.Funnel, 6)
				other := stats.Funnel[5]
				require.Equal(t, "Other", other.Name)
				require.Equal(t, 1, other.JobsReached)
				require.Equal(t, 6, other.SortOrder, "after the last default stage")
			},
		},
		{
			name: "revisited_stage_still_counts_once",
			seed: func(t *testing.T, s *Store) {
				t.Helper()
				jobID := seedStatsJob(t, s, StatusInProgress, &applied, false)
				phoneScreen := stageIDByName(t, s, jobID, "Phone Screen")
				addStatsLog(t, s, jobID, phoneScreen, statsNow.AddDate(0, 0, -9))
				addStatsLog(t, s, jobID, stageIDByName(t, s, jobID, "Technical Interview"), statsNow.AddDate(0, 0, -6))
				addStatsLog(t, s, jobID, phoneScreen, statsNow.AddDate(0, 0, -3))
			},
			check: func(t *testing.T, stats *Stats) {
				t.Helper()
				require.Equal(t, 1, funnelStage(t, stats, "Phone Screen").JobsReached)
			},
		},
		{
			name: "null_stage_log_adds_no_entry",
			seed: func(t *testing.T, s *Store) {
				t.Helper()
				jobID := seedStatsJob(t, s, StatusInProgress, &applied, false)
				addStatsLog(t, s, jobID, 0, statsNow.AddDate(0, 0, -2))
			},
			check: func(t *testing.T, stats *Stats) {
				t.Helper()
				require.Len(t, stats.Funnel, 5, "no Other row")
				for _, stage := range stats.Funnel {
					require.Zero(t, stage.JobsReached, "stage %s", stage.Name)
				}
			},
		},
		{
			name: "other_row_absent_when_unused",
			seed: func(t *testing.T, s *Store) {
				t.Helper()
				jobID := seedStatsJob(t, s, StatusInProgress, &applied, false)
				addStatsLog(t, s, jobID, stageIDByName(t, s, jobID, "Phone Screen"), statsNow.AddDate(0, 0, -2))
			},
			check: func(t *testing.T, stats *Stats) {
				t.Helper()
				require.Len(t, stats.Funnel, 5)
				for _, stage := range stats.Funnel {
					require.NotEqual(t, "Other", stage.Name)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			s := newStatsStore(t)
			test.seed(t, s)

			stats, err := s.Stats(statsNow)
			require.NoError(t, err)
			test.check(t, stats)
		})
	}
}

func TestStatsAvgTimePerStage(t *testing.T) {
	t.Parallel()

	t.Run("closed_stints_average_across_jobs", func(t *testing.T) {
		t.Parallel()

		s := newStatsStore(t)
		applied := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
		start := statsNow.AddDate(0, 0, -10)

		first := seedStatsJob(t, s, StatusInProgress, &applied, false)
		addStatsLog(t, s, first, stageIDByName(t, s, first, "Phone Screen"), start)
		addStatsLog(t, s, first, stageIDByName(t, s, first, "Technical Interview"), start.AddDate(0, 0, 3))

		second := seedStatsJob(t, s, StatusInProgress, &applied, false)
		addStatsLog(t, s, second, stageIDByName(t, s, second, "Phone Screen"), start)
		addStatsLog(t, s, second, stageIDByName(t, s, second, "Technical Interview"), start.AddDate(0, 0, 5))

		stats, err := s.Stats(statsNow)
		require.NoError(t, err)

		phoneScreen := funnelStage(t, stats, "Phone Screen")
		require.NotNil(t, phoneScreen.AvgDays)
		require.InDelta(t, 4.0, *phoneScreen.AvgDays, 1e-9, "3-day and 5-day stints average to 4")
	})

	t.Run("current_stage_counts_to_now", func(t *testing.T) {
		t.Parallel()

		s := newStatsStore(t)
		applied := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)

		jobID := seedStatsJob(t, s, StatusInProgress, &applied, false)
		addStatsLog(t, s, jobID, stageIDByName(t, s, jobID, "Phone Screen"), statsNow.Add(-48*time.Hour))

		stats, err := s.Stats(statsNow)
		require.NoError(t, err)

		phoneScreen := funnelStage(t, stats, "Phone Screen")
		require.NotNil(t, phoneScreen.AvgDays)
		require.InDelta(t, 2.0, *phoneScreen.AvgDays, 1e-9, "open stint runs to now")
	})
}

func TestStatsFirstResponseAndNullAppliedAt(t *testing.T) {
	t.Parallel()

	s := newStatsStore(t)

	// applied_at is a wall date: stored at local midnight in its own offset,
	// diffed against the log's calendar day, never re-projected.
	appliedWest := time.Date(2026, 1, 1, 0, 0, 0, 0, time.FixedZone("CST", -6*3600))
	first := seedStatsJob(t, s, StatusInProgress, &appliedWest, false)
	addStatsLog(t, s, first, stageIDByName(t, s, first, "Phone Screen"), time.Date(2026, 1, 3, 12, 0, 0, 0, time.UTC)) // 2 days

	appliedUTC := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	second := seedStatsJob(t, s, StatusInProgress, &appliedUTC, false)
	addStatsLog(t, s, second, stageIDByName(t, s, second, "Phone Screen"), time.Date(2026, 1, 6, 8, 0, 0, 0, time.UTC)) // 4 days

	seedStatsJob(t, s, StatusRejected, &appliedUTC, false) // zero logs: not a response

	noDate := seedStatsJob(t, s, StatusInProgress, nil, false)
	addStatsLog(t, s, noDate, stageIDByName(t, s, noDate, "Technical Interview"), time.Date(2026, 1, 4, 0, 0, 0, 0, time.UTC))
	addStatsLog(t, s, noDate, stageIDByName(t, s, noDate, "Final Round"), time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC))

	stats, err := s.Stats(statsNow)
	require.NoError(t, err)

	require.NotNil(t, stats.AvgDaysToFirstResponse)
	require.InDelta(t, 3.0, *stats.AvgDaysToFirstResponse, 1e-9, "mean of 2 and 4; no-log and nil-applied jobs excluded")

	techInterview := funnelStage(t, stats, "Technical Interview")
	require.Equal(t, 1, techInterview.JobsReached, "nil-applied job still counts in the funnel")
	require.Nil(t, techInterview.AvgDays, "nil-applied job contributes no time samples")
	require.Equal(t, 1, stats.StatusBreakdown[StatusRejected])
	require.Equal(t, 3, stats.StatusBreakdown[StatusInProgress], "nil-applied job still counts in the breakdown")
}

func newStatsStore(t *testing.T) *Store {
	t.Helper()

	s, err := Open(filepath.Join(t.TempDir(), "jobs.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func seedStatsJob(t *testing.T, s *Store, status ApplicationStatus, appliedAt *time.Time, archived bool) uint {
	t.Helper()

	job := &Job{
		Company:   "ACME",
		Position:  "Engineer",
		Status:    status,
		AppliedAt: appliedAt,
	}
	if archived {
		archivedAt := statsNow.AddDate(0, 0, -1)
		job.ArchivedAt = &archivedAt
	}
	require.NoError(t, s.Create(job))
	return job.ID
}

// stageIDByName resolves a job's cloned per-job stage by template name.
func stageIDByName(t *testing.T, s *Store, jobID uint, name string) uint {
	t.Helper()

	stages, err := s.ListStages(jobID)
	require.NoError(t, err)
	for _, stage := range stages {
		if stage.Name == name {
			return stage.ID
		}
	}
	t.Fatalf("stage %q not found for job %d", name, jobID)
	return 0
}

// addStatsLog adds a stage log with a fixed CreatedAt (GORM keeps non-zero
// timestamps). stageID 0 means a null stage_id log (stage cleared).
func addStatsLog(t *testing.T, s *Store, jobID, stageID uint, createdAt time.Time) {
	t.Helper()

	log := &StageLog{CreatedAt: createdAt}
	if stageID != 0 {
		log.StageID = &stageID
	}
	require.NoError(t, s.AddStageLog(jobID, log))
}

func funnelStage(t *testing.T, stats *Stats, name string) StageStat {
	t.Helper()

	for _, stage := range stats.Funnel {
		if stage.Name == name {
			return stage
		}
	}
	t.Fatalf("funnel stage %q not found", name)
	return StageStat{}
}
