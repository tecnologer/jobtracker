package store

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	companyAcme      = "Acme"
	positionEngineer = "Engineer"
	titleSoon        = "Soon"
	titleLater       = "Later"
)

func TestMeetingCRUD(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "jobs.db")

	s, err := Open(path)
	require.NoError(t, err)

	job := &Job{
		Company:  companyAcme,
		Position: positionEngineer,
	}
	require.NoError(t, s.Create(job))

	scheduled := time.Now().Add(48 * time.Hour)
	m := &Meeting{
		JobID:       job.ID,
		Title:       "Tech interview",
		ScheduledAt: scheduled,
		URL:         "https://meet.example.com/x",
		Notes:       "bring resume",
	}
	require.NoError(t, s.CreateMeeting(m))
	require.NotZero(t, m.ID, "meeting ID must be set after create")

	meetings, err := s.ListMeetings(job.ID)
	require.NoError(t, err)
	require.Len(t, meetings, 1)
	assert.Equal(t, "Tech interview", meetings[0].Title)

	updated := &Meeting{
		Title:       "Tech interview (rescheduled)",
		ScheduledAt: scheduled.Add(24 * time.Hour),
		URL:         "https://meet.example.com/y",
		Notes:       "bring laptop",
	}
	require.NoError(t, s.UpdateMeeting(m.ID, updated))

	meetings, err = s.ListMeetings(job.ID)
	require.NoError(t, err)
	require.Len(t, meetings, 1)
	assert.Equal(t, "Tech interview (rescheduled)", meetings[0].Title)
	assert.Equal(t, "https://meet.example.com/y", meetings[0].URL)

	require.NoError(t, s.DeleteMeeting(m.ID))
	meetings, err = s.ListMeetings(job.ID)
	require.NoError(t, err)
	assert.Empty(t, meetings, "meetings after delete")
}

func TestMeetingListOrder(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "jobs.db")

	s, err := Open(path)
	require.NoError(t, err)

	job := &Job{
		Company:  companyAcme,
		Position: positionEngineer,
	}
	require.NoError(t, s.Create(job))

	later := time.Now().Add(72 * time.Hour)
	sooner := time.Now().Add(24 * time.Hour)
	require.NoError(t, s.CreateMeeting(&Meeting{JobID: job.ID, Title: titleLater, ScheduledAt: later}))
	require.NoError(t, s.CreateMeeting(&Meeting{JobID: job.ID, Title: "Sooner", ScheduledAt: sooner}))

	meetings, err := s.ListMeetings(job.ID)
	require.NoError(t, err)
	require.Len(t, meetings, 2)
	assert.Equal(t, "Sooner", meetings[0].Title, "ascending order by scheduled_at")
	assert.Equal(t, titleLater, meetings[1].Title, "ascending order by scheduled_at")
}

func TestMeetingCascadeOnJobDelete(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "jobs.db")

	s, err := Open(path)
	require.NoError(t, err)

	job := &Job{
		Company:  companyAcme,
		Position: positionEngineer,
	}
	require.NoError(t, s.Create(job))
	require.NoError(t, s.CreateMeeting(&Meeting{JobID: job.ID, Title: "Interview", ScheduledAt: time.Now().Add(24 * time.Hour)}))

	require.NoError(t, s.Delete(job.ID))

	meetings, err := s.ListMeetings(job.ID)
	require.NoError(t, err)
	assert.Empty(t, meetings, "meetings after cascade delete")
}

func TestListUpcomingMeetings(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "jobs.db")

	s, err := Open(path)
	require.NoError(t, err)

	jobA := &Job{
		Company:  companyAcme,
		Position: positionEngineer,
	}
	require.NoError(t, s.Create(jobA))
	jobB := &Job{
		Company:  "Globex",
		Position: "Manager",
	}
	require.NoError(t, s.Create(jobB))
	archivedJob := &Job{
		Company:  "Initech",
		Position: "Analyst",
	}
	require.NoError(t, s.Create(archivedJob))
	archivedAt := time.Now()
	archivedJob.ArchivedAt = &archivedAt
	require.NoError(t, s.Update(archivedJob.ID, archivedJob), "archive job")
	deletedJob := &Job{
		Company:  "Umbrella",
		Position: "Researcher",
	}
	require.NoError(t, s.Create(deletedJob))

	past := time.Now().Add(-2 * time.Hour)
	soon := time.Now().Add(2 * time.Hour)
	later := time.Now().Add(48 * time.Hour)

	require.NoError(t, s.CreateMeeting(&Meeting{JobID: jobA.ID, Title: "Past", ScheduledAt: past}))
	require.NoError(t, s.CreateMeeting(&Meeting{JobID: jobA.ID, Title: titleSoon, ScheduledAt: soon}))
	require.NoError(t, s.CreateMeeting(&Meeting{JobID: jobB.ID, Title: titleLater, ScheduledAt: later}))
	require.NoError(t, s.CreateMeeting(&Meeting{JobID: archivedJob.ID, Title: "Archived job meeting", ScheduledAt: soon}))
	require.NoError(t, s.CreateMeeting(&Meeting{JobID: deletedJob.ID, Title: "Soft-deleted job meeting", ScheduledAt: soon}))
	// Soft-delete the job directly (bypassing store.Delete's meeting cascade) so this
	// test exercises the upcoming query's own deleted_at exclusion, not the cascade.
	require.NoError(t, s.db.Model(&Job{}).Where("id = ?", deletedJob.ID).Update("deleted_at", time.Now()).Error, "soft-delete job")

	upcoming, err := s.ListUpcomingMeetings(10)
	require.NoError(t, err)
	require.Len(t, upcoming, 2, "only future meetings of live, unarchived jobs")
	assert.Equal(t, titleSoon, upcoming[0].Title, "ascending order")
	assert.Equal(t, titleLater, upcoming[1].Title, "ascending order")
	require.NotNil(t, upcoming[0].Job, "job must be preloaded")
	assert.Equal(t, "Acme", upcoming[0].Job.Company)

	limited, err := s.ListUpcomingMeetings(1)
	require.NoError(t, err)
	require.Len(t, limited, 1, "limit=1")
	assert.Equal(t, titleSoon, limited[0].Title, "soonest meeting first")
}

func TestNextMeetingTimes(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "jobs.db")

	s, err := Open(path)
	require.NoError(t, err)

	jobA := &Job{
		Company:  companyAcme,
		Position: positionEngineer,
	}
	require.NoError(t, s.Create(jobA))
	jobB := &Job{
		Company:  "Globex",
		Position: "Manager",
	}
	require.NoError(t, s.Create(jobB))
	jobNoMeetings := &Job{
		Company:  "NoMeetings Inc",
		Position: "Nobody",
	}
	require.NoError(t, s.Create(jobNoMeetings))

	past := time.Now().Add(-2 * time.Hour)
	soonest := time.Now().Add(2 * time.Hour)
	later := time.Now().Add(48 * time.Hour)

	require.NoError(t, s.CreateMeeting(&Meeting{JobID: jobA.ID, Title: "Past", ScheduledAt: past}))
	require.NoError(t, s.CreateMeeting(&Meeting{JobID: jobA.ID, Title: titleLater, ScheduledAt: later}))
	require.NoError(t, s.CreateMeeting(&Meeting{JobID: jobA.ID, Title: "Soonest", ScheduledAt: soonest}))

	nextTimes, err := s.NextMeetingTimes()
	require.NoError(t, err)
	got, ok := nextTimes[jobA.ID]
	require.True(t, ok, "expected an entry for jobA, got %+v", nextTimes)
	assert.True(t, got.Equal(soonest), "jobA's next meeting must be the soonest upcoming one (%v), got %v", soonest, got)
	assert.NotContains(t, nextTimes, jobB.ID, "jobB has no meetings")
	assert.NotContains(t, nextTimes, jobNoMeetings.ID, "jobNoMeetings has no meetings")
}
