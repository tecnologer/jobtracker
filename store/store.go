package store

import (
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type ApplicationStatus string

const (
	StatusProspect    ApplicationStatus = "prospect"
	StatusApplied     ApplicationStatus = "applied"
	StatusInProgress  ApplicationStatus = "in_progress"
	StatusOnHold      ApplicationStatus = "on_hold"
	StatusNegotiating ApplicationStatus = "negotiating"
	StatusAccepted    ApplicationStatus = "accepted"
	StatusRejected    ApplicationStatus = "rejected"
	StatusCanceled    ApplicationStatus = "canceled"
)

type Stage struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	JobID     uint   `json:"job_id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

type StageLog struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	JobID       uint      `json:"job_id"`
	PrevStageID *uint     `json:"prev_stage_id"`
	PrevStage   *Stage    `json:"prev_stage,omitempty" gorm:"foreignKey:PrevStageID"`
	StageID     *uint     `json:"stage_id"`
	Stage       *Stage    `json:"stage,omitempty" gorm:"foreignKey:StageID"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
}

type Contact struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	JobID uint   `json:"job_id"`
	Name  string `json:"name"`
	Role  string `json:"role"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

// Meeting is a scheduled call/interview attached to a job. Unlike Job.AppliedAt,
// ScheduledAt is a real instant (RFC3339), not a wall date: it must render in the
// viewer's local timezone, never through the applied_at wall-date helpers.
type Meeting struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	JobID       uint      `gorm:"index:idx_meetings_job_sched" json:"job_id"`
	Job         *Job      `json:"job,omitempty" gorm:"foreignKey:JobID"` // preloaded only by ListUpcomingMeetings
	Title       string    `json:"title"`
	ScheduledAt time.Time `gorm:"index:idx_meetings_job_sched" json:"scheduled_at"`
	URL         string    `json:"url"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
}

type Job struct {
	ID         uint              `gorm:"primaryKey" json:"id"`
	Company    string            `json:"company"`
	Position   string            `json:"position"`
	Status     ApplicationStatus `json:"status"`
	AppliedAt  *time.Time        `json:"applied_at"` // timezone-aware (RFC3339); nil when unset
	Notes      string            `json:"notes"`
	StageID    *uint             `json:"stage_id"`
	Stage      *Stage            `json:"stage,omitempty" gorm:"foreignKey:StageID"`
	Stages     []Stage           `json:"stages,omitempty" gorm:"foreignKey:JobID"`
	URL        string            `json:"url"`
	ArchivedAt *time.Time        `json:"archived_at"`
	TopMatch   bool              `json:"top_match"` // manual flag; excluded from jobEditFields, see SetTopMatch
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	DeletedAt  gorm.DeletedAt    `gorm:"index" json:"deleted_at"`
}

type Store struct {
	db *gorm.DB
}

func Open(path string) (*Store, error) {
	db, err := gorm.Open(sqlite.Open(path+"?_pragma=journal_mode(WAL)"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1) // ponytail: SQLite only supports one writer; single conn serializes without busy-wait
	s := &Store{db: db}
	if err := db.AutoMigrate(&Stage{}, &Job{}, &StageLog{}, &Contact{}, &Meeting{}); err != nil {
		return nil, fmt.Errorf("migrating schema: %w", err)
	}
	// one-time cleanup: remove stages that were mistakenly seeded as stages, now they are statuses
	db.Where("job_id = 0 AND name IN ?", []string{"Negotiating", "On Hold", "Accepted"}).Delete(&Stage{})
	return s, s.seedDefaultStages()
}

func (s *Store) seedDefaultStages() error {
	var count int64
	if err := s.db.Model(&Stage{}).Where("job_id = 0").Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	// ponytail: job_id=0 has an FK to jobs(id); disable FK enforcement for this insert in case it's on
	var fkOn int
	s.db.Raw("PRAGMA foreign_keys").Scan(&fkOn)
	if fkOn == 1 {
		s.db.Exec("PRAGMA foreign_keys = OFF")
		defer s.db.Exec("PRAGMA foreign_keys = ON")
	}
	return s.db.Create(&[]Stage{
		{JobID: 0, Name: "Phone Screen", SortOrder: 1},
		{JobID: 0, Name: "Technical Interview", SortOrder: 2},
		{JobID: 0, Name: "Code Challenge", SortOrder: 3},
		{JobID: 0, Name: "Final Round", SortOrder: 4},
		{JobID: 0, Name: "Offer", SortOrder: 5},
	}).Error
}

// Close closes the underlying database connection, checkpointing the WAL.
func (s *Store) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return fmt.Errorf("getting underlying database: %w", err)
	}
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("closing database: %w", err)
	}
	return nil
}

func (s *Store) List() ([]Job, error) {
	jobs := []Job{}
	return jobs, s.db.Preload("Stage").Preload("Stages", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order")
	}).Find(&jobs).Error
}

func (s *Store) Create(j *Job) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(j).Error; err != nil {
			return err
		}
		var defaults []Stage
		if err := tx.Where("job_id = 0").Order("sort_order").Find(&defaults).Error; err != nil {
			return err
		}
		for i := range defaults {
			defaults[i].ID = 0
			defaults[i].JobID = j.ID
		}
		if len(defaults) > 0 {
			return tx.Create(&defaults).Error
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("creating job: %w", err)
	}
	return nil
}

func (s *Store) ListDefaultStages() ([]Stage, error) {
	stages := []Stage{}
	return stages, s.db.Where("job_id = 0").Order("sort_order").Find(&stages).Error
}

func (s *Store) CreateDefaultStage(stage *Stage) error {
	stage.JobID = 0
	return s.db.Create(stage).Error
}

func jobEditFields() []string {
	return []string{"company", "position", "status", "applied_at", "notes", "url", "archived_at", "stage_id"}
}

func jobBaseFields() []string { // jobEditFields without stage_id
	fields := jobEditFields()
	return fields[:len(fields)-1]
}

func (s *Store) Update(id uint, j *Job) error {
	if j.StageID == nil || *j.StageID == 0 {
		// ponytail: omit stage_id so a nil payload doesn't overwrite a stage set by the log endpoint
		return s.db.Model(&Job{}).Where("id = ?", id).Select(jobBaseFields()).Updates(j).Error
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var current Job
		if err := tx.Select("stage_id").First(&current, id).Error; err != nil {
			return err
		}
		if err := tx.Model(&Job{}).Where("id = ?", id).Select(jobEditFields()).Updates(j).Error; err != nil {
			return err
		}
		if current.StageID != nil && *current.StageID == *j.StageID {
			return nil
		}
		return tx.Create(&StageLog{JobID: id, PrevStageID: current.StageID, StageID: j.StageID}).Error
	})
	if err != nil {
		return fmt.Errorf("updating job: %w", err)
	}
	return nil
}

// SetTopMatch updates only the top_match column for a job, deliberately
// bypassing jobEditFields/jobBaseFields so it can never be reached through
// the general Update whitelist (see the stage_id exclusion above for the
// same reasoning).
func (s *Store) SetTopMatch(id uint, v bool) error {
	return s.db.Model(&Job{}).Where("id = ?", id).Update("top_match", v).Error
}

func (s *Store) Delete(id uint) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("job_id = ?", id).Delete(&Stage{}).Error; err != nil {
			return err
		}
		if err := tx.Where("job_id = ?", id).Delete(&Meeting{}).Error; err != nil {
			return err
		}
		return tx.Delete(&Job{}, id).Error
	})
	if err != nil {
		return fmt.Errorf("deleting job: %w", err)
	}
	return nil
}

func (s *Store) ListStages(jobID uint) ([]Stage, error) {
	stages := []Stage{}
	return stages, s.db.Where("job_id = ?", jobID).Order("sort_order").Find(&stages).Error
}

func (s *Store) CreateStage(stage *Stage) error {
	return s.db.Create(stage).Error
}

func (s *Store) UpdateStage(id uint, stage *Stage) error {
	return s.db.Model(&Stage{}).Where("id = ?", id).Updates(map[string]any{
		"name":       stage.Name,
		"sort_order": stage.SortOrder,
	}).Error
}

func (s *Store) DeleteStage(id uint) error {
	return s.db.Delete(&Stage{}, id).Error
}

func (s *Store) AddStageLog(jobID uint, log *StageLog) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var current Job
		if err := tx.Select("stage_id").First(&current, jobID).Error; err != nil {
			return err
		}
		log.JobID = jobID
		log.PrevStageID = current.StageID
		if err := tx.Create(log).Error; err != nil {
			return err
		}
		var stageVal any
		if log.StageID != nil {
			stageVal = *log.StageID
		}
		return tx.Model(&Job{}).Where("id = ?", jobID).Update("stage_id", stageVal).Error
	})
	if err != nil {
		return fmt.Errorf("adding stage log: %w", err)
	}
	return nil
}

func (s *Store) ListContacts(jobID uint) ([]Contact, error) {
	var contacts []Contact
	return contacts, s.db.Where("job_id = ?", jobID).Find(&contacts).Error
}

func (s *Store) CreateContact(c *Contact) error {
	return s.db.Create(c).Error
}

func (s *Store) DeleteContact(id uint) error {
	return s.db.Delete(&Contact{}, id).Error
}

func (s *Store) ListStageLogs(jobID uint) ([]StageLog, error) {
	logs := []StageLog{}
	return logs, s.db.Where("job_id = ?", jobID).Preload("Stage").Preload("PrevStage").Order("created_at desc").Find(&logs).Error
}

func (s *Store) ListMeetings(jobID uint) ([]Meeting, error) {
	meetings := []Meeting{}
	return meetings, s.db.Where("job_id = ?", jobID).Order("scheduled_at asc").Find(&meetings).Error
}

func (s *Store) CreateMeeting(m *Meeting) error {
	return s.db.Create(m).Error
}

func (s *Store) UpdateMeeting(id uint, m *Meeting) error {
	return s.db.Model(&Meeting{}).Where("id = ?", id).Updates(map[string]any{
		"title":        m.Title,
		"scheduled_at": m.ScheduledAt,
		"url":          m.URL,
		"notes":        m.Notes,
	}).Error
}

func (s *Store) DeleteMeeting(id uint) error {
	return s.db.Delete(&Meeting{}, id).Error
}

// ListUpcomingMeetings returns the next `limit` meetings across all jobs
// (scheduled_at >= now, ascending), preloading each meeting's Job for display.
// Meetings belonging to soft-deleted or archived jobs are excluded via the join.
func (s *Store) ListUpcomingMeetings(limit int) ([]Meeting, error) {
	if limit <= 0 {
		limit = 10
	}
	meetings := []Meeting{}
	err := s.db.
		Preload("Job").
		Joins("JOIN jobs ON jobs.id = meetings.job_id").
		Where("meetings.scheduled_at >= ?", time.Now()).
		Where("jobs.deleted_at IS NULL AND jobs.archived_at IS NULL").
		Order("meetings.scheduled_at asc").
		Limit(limit).
		Find(&meetings).Error
	return meetings, err
}

// NextMeetingTimes returns, per job ID, the soonest upcoming (scheduled_at >= now)
// meeting time, computed with a single grouped MIN query rather than per-row lookups.
// Used by the CSV export's "Next Meeting" column.
func (s *Store) NextMeetingTimes() (map[uint]time.Time, error) {
	// Next is scanned as a string: the sqlite driver returns aggregate (MIN)
	// datetime columns as raw text rather than converting them to time.Time,
	// since the aggregate loses the source column's declared type.
	type nextMeetingRow struct {
		JobID uint
		Next  string
	}
	var rows []nextMeetingRow
	err := s.db.Model(&Meeting{}).
		Select("job_id, MIN(scheduled_at) AS next").
		Where("scheduled_at >= ?", time.Now()).
		Group("job_id").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uint]time.Time, len(rows))
	for _, row := range rows {
		// Parse using the same layout GORM/the sqlite driver stores time.Time in.
		parsed, err := time.Parse("2006-01-02 15:04:05.999999999-07:00", row.Next)
		if err != nil {
			return nil, fmt.Errorf("parsing next meeting time %q: %w", row.Next, err)
		}
		result[row.JobID] = parsed
	}
	return result, nil
}
