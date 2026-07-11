package store

import (
	"fmt"
	"time"
)

// otherBucket collects funnel entries whose stage name matches no current
// default stage (renamed/custom stages, or defaults deleted after logging).
const otherBucket = "Other"

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

// Stats computes all dashboard aggregates in one pass over the stage logs.
// Read-only. now closes the open stint of each job's current stage, so the
// number moves between calls; tests pass a fixed now for determinism.
// Funnel buckets group cross-job stages by exact name against the current
// default stages; everything else lands in the "Other" bucket. Jobs with a
// nil applied_at are excluded from all time math but still counted in the
// funnel and status breakdown.
func (s *Store) Stats(now time.Time) (*Stats, error) {
	var jobs []Job
	if err := s.db.Model(&Job{}).Select("id", "status", "applied_at", "archived_at").Find(&jobs).Error; err != nil {
		return nil, fmt.Errorf("loading jobs: %w", err)
	}
	defaults, err := s.ListDefaultStages()
	if err != nil {
		return nil, fmt.Errorf("loading default stages: %w", err)
	}

	var rows []statLogRow
	err = s.db.Model(&StageLog{}).
		Select("stage_logs.job_id, stages.name AS stage_name, stage_logs.created_at").
		Joins("JOIN jobs ON jobs.id = stage_logs.job_id AND jobs.deleted_at IS NULL").
		Joins("LEFT JOIN stages ON stages.id = stage_logs.stage_id").
		Order("stage_logs.job_id, stage_logs.created_at").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("loading stage logs: %w", err)
	}

	stats := &Stats{
		TotalJobs:       len(jobs),
		StatusBreakdown: make(map[ApplicationStatus]int, 8),
	}
	appliedAtByJob, offerJobs := summarizeJobs(jobs, stats)

	defaultNames := make(map[string]bool, len(defaults))
	for _, stage := range defaults {
		defaultNames[stage.Name] = true
	}
	bucket := func(name string) string {
		if defaultNames[name] {
			return name
		}
		return otherBucket
	}

	acc := accumulateFunnel(rows, appliedAtByJob, offerJobs, bucket, now)

	stats.Offers = len(offerJobs)
	if acc.responseCount > 0 {
		avg := acc.responseSum / float64(acc.responseCount)
		stats.AvgDaysToFirstResponse = &avg
	}

	stats.Funnel = make([]StageStat, 0, len(defaults)+1)
	lastOrder := 0
	for _, stage := range defaults {
		stats.Funnel = append(stats.Funnel, stageStat(stage.Name, stage.SortOrder, acc))
		lastOrder = stage.SortOrder
	}
	if len(acc.reached[otherBucket]) > 0 {
		stats.Funnel = append(stats.Funnel, stageStat(otherBucket, lastOrder+1, acc))
	}
	return stats, nil
}

// summarizeJobs fills the status breakdown and job-level KPIs on stats,
// returning each job's applied_at for the time math over the stage logs and
// the set of jobs with an accepted offer (accumulateFunnel later adds jobs
// that logged an Offer stage to the same set).
func summarizeJobs(jobs []Job, stats *Stats) (map[uint]*time.Time, map[uint]bool) {
	for _, status := range []ApplicationStatus{
		StatusProspect, StatusApplied, StatusInProgress, StatusOnHold,
		StatusNegotiating, StatusAccepted, StatusRejected, StatusCanceled,
	} {
		stats.StatusBreakdown[status] = 0
	}

	appliedAtByJob := make(map[uint]*time.Time, len(jobs))
	offerJobs := map[uint]bool{}
	rejected, nonProspect := 0, 0
	for _, job := range jobs {
		stats.StatusBreakdown[job.Status]++
		appliedAtByJob[job.ID] = job.AppliedAt
		if job.Status != StatusRejected && job.Status != StatusCanceled && job.ArchivedAt == nil {
			stats.ActiveJobs++
		}
		if job.Status == StatusRejected {
			rejected++
		}
		if job.Status != StatusProspect {
			nonProspect++
		}
		if job.Status == StatusAccepted {
			offerJobs[job.ID] = true
		}
	}
	if nonProspect > 0 {
		stats.RejectionRate = float64(rejected) / float64(nonProspect)
	}
	return appliedAtByJob, offerJobs
}

type statLogRow struct {
	JobID     uint
	StageName *string // nil: null stage_id or hard-deleted stage row
	CreatedAt time.Time
}

// funnelAcc holds the running aggregates from the single pass over stage logs.
type funnelAcc struct {
	reached       map[string]map[uint]bool // bucket -> distinct jobs
	stintSum      map[string]float64
	stintCount    map[string]int
	responseSum   float64
	responseCount int
}

// accumulateFunnel makes a single pass over the logs, which arrive grouped by
// job and ordered by time, collecting per-bucket reach and stint durations
// plus first-response times. Jobs that logged an "Offer" stage are marked in
// offerJobs.
func accumulateFunnel(
	rows []statLogRow, appliedAtByJob map[uint]*time.Time,
	offerJobs map[uint]bool, bucket func(string) string, now time.Time,
) *funnelAcc {
	acc := &funnelAcc{
		reached:    map[string]map[uint]bool{},
		stintSum:   map[string]float64{},
		stintCount: map[string]int{},
	}
	var lastJobID uint
	for i, row := range rows {
		appliedAt, known := appliedAtByJob[row.JobID]
		if !known {
			continue
		}
		if row.JobID != lastJobID { // first log of this job = its response
			lastJobID = row.JobID
			acc.addResponse(appliedAt, row.CreatedAt)
		}
		if row.StageName == nil {
			continue // ends the previous stint (via the next-log lookup) but adds no entry
		}
		if *row.StageName == "Offer" {
			offerJobs[row.JobID] = true
		}
		name := bucket(*row.StageName)
		if acc.reached[name] == nil {
			acc.reached[name] = map[uint]bool{}
		}
		acc.reached[name][row.JobID] = true
		if appliedAt == nil {
			continue // FR-08: no time math without an applied date
		}
		acc.stintSum[name] += stintEnd(rows, i, now).Sub(row.CreatedAt).Hours() / 24
		acc.stintCount[name]++
	}
	return acc
}

// addResponse records the days from applied_at (a wall date) to the job's
// first log; jobs without an applied date are excluded from the average.
func (acc *funnelAcc) addResponse(appliedAt *time.Time, firstLogAt time.Time) {
	if appliedAt == nil {
		return
	}
	acc.responseSum += wallDateDiffDays(*appliedAt, firstLogAt)
	acc.responseCount++
}

// stintEnd returns when the stage logged at rows[i] ended: at the job's next
// log, or now if it is the job's current (still open) stage.
func stintEnd(rows []statLogRow, i int, now time.Time) time.Time {
	if i+1 < len(rows) && rows[i+1].JobID == rows[i].JobID {
		return rows[i+1].CreatedAt
	}
	return now
}

func stageStat(name string, order int, acc *funnelAcc) StageStat {
	stat := StageStat{Name: name, SortOrder: order, JobsReached: len(acc.reached[name])}
	if count := acc.stintCount[name]; count > 0 {
		avg := acc.stintSum[name] / float64(count)
		stat.AvgDays = &avg
	}
	return stat
}

// wallDateDiffDays returns whole calendar days between two timestamps, each
// reduced to its YYYY-MM-DD in its own stored offset — applied_at is a wall
// date, so it must never be re-projected into another timezone (CLAUDE.md).
func wallDateDiffDays(from, to time.Time) float64 {
	fromDay, _ := time.Parse("2006-01-02", from.Format("2006-01-02"))
	toDay, _ := time.Parse("2006-01-02", to.Format("2006-01-02"))
	return toDay.Sub(fromDay).Hours() / 24
}
