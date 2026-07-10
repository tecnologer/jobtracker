package store

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetTopMatchPersists(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "jobs.db")

	s, err := Open(path)
	require.NoError(t, err)

	job := &Job{
		Company:  companyAcme,
		Position: positionEngineer,
	}
	require.NoError(t, s.Create(job))
	require.False(t, job.TopMatch, "top_match must default to false")

	require.NoError(t, s.SetTopMatch(job.ID, true))
	jobs, err := s.List()
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	assert.True(t, jobs[0].TopMatch, "top_match true after SetTopMatch")

	require.NoError(t, s.SetTopMatch(job.ID, false))
	jobs, err = s.List()
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	assert.False(t, jobs[0].TopMatch, "top_match false after unsetting")
}

func TestUpdateDoesNotClobberTopMatch(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "jobs.db")

	s, err := Open(path)
	require.NoError(t, err)

	job := &Job{
		Company:  companyAcme,
		Position: positionEngineer,
	}
	require.NoError(t, s.Create(job))
	require.NoError(t, s.SetTopMatch(job.ID, true))

	// A general edit payload that omits top_match (zero value) must not reset the flag.
	edit := &Job{
		Company:  "Acme Corp",
		Position: "Senior Engineer",
	}
	require.NoError(t, s.Update(job.ID, edit))

	jobs, err := s.List()
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	assert.Equal(t, "Acme Corp", jobs[0].Company)
	assert.True(t, jobs[0].TopMatch, "top_match must survive an unrelated Update")
}
