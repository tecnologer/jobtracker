package store

import (
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
		return nil, err
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1) // ponytail: SQLite only supports one writer; single conn serializes without busy-wait
	s := &Store{db: db}
	if err := db.AutoMigrate(&Stage{}, &Job{}, &StageLog{}, &Contact{}); err != nil {
		return nil, err
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

func (s *Store) List() ([]Job, error) {
	jobs := []Job{}
	return jobs, s.db.Preload("Stage").Preload("Stages", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order")
	}).Find(&jobs).Error
}

func (s *Store) Create(j *Job) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
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
}

func (s *Store) ListDefaultStages() ([]Stage, error) {
	stages := []Stage{}
	return stages, s.db.Where("job_id = 0").Order("sort_order").Find(&stages).Error
}

func (s *Store) CreateDefaultStage(stage *Stage) error {
	stage.JobID = 0
	return s.db.Create(stage).Error
}

var jobEditFields = []string{"company", "position", "status", "applied_at", "notes", "url", "archived_at", "stage_id"}
var jobBaseFields = jobEditFields[:len(jobEditFields)-1] // without stage_id

func (s *Store) Update(id uint, j *Job) error {
	if j.StageID == nil || *j.StageID == 0 {
		// ponytail: omit stage_id so a nil payload doesn't overwrite a stage set by the log endpoint
		return s.db.Model(&Job{}).Where("id = ?", id).Select(jobBaseFields).Updates(j).Error
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		var current Job
		if err := tx.Select("stage_id").First(&current, id).Error; err != nil {
			return err
		}
		if err := tx.Model(&Job{}).Where("id = ?", id).Select(jobEditFields).Updates(j).Error; err != nil {
			return err
		}
		if current.StageID != nil && *current.StageID == *j.StageID {
			return nil
		}
		return tx.Create(&StageLog{JobID: id, PrevStageID: current.StageID, StageID: j.StageID}).Error
	})
}

func (s *Store) Delete(id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("job_id = ?", id).Delete(&Stage{}).Error; err != nil {
			return err
		}
		return tx.Delete(&Job{}, id).Error
	})
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
	return s.db.Transaction(func(tx *gorm.DB) error {
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
