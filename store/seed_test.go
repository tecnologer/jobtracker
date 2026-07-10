package store

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSeedDefaultStages(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "jobs.db")

	s, err := Open(path)
	require.NoError(t, err)
	stages, err := s.ListDefaultStages()
	require.NoError(t, err)
	require.NotEmpty(t, stages, "expected default stages")
	t.Logf("got %d default stages", len(stages))
}

func TestSeedDefaultStagesWithFKEnabled(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "jobs.db")

	s, err := Open(path)
	require.NoError(t, err)

	// Enable FK enforcement, wipe default stages, reopen — seed must succeed despite FK constraint
	s.db.Exec("PRAGMA foreign_keys = ON")
	s.db.Where("job_id = 0").Delete(&Stage{})

	s2, err := Open(path)
	require.NoError(t, err, "reopen with FK on")
	stages, err := s2.ListDefaultStages()
	require.NoError(t, err)
	require.NotEmpty(t, stages, "expected default stages after reopen with FK enabled")
	t.Logf("got %d default stages with FK enforcement on", len(stages))
}
