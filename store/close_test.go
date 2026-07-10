package store

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClose(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "jobs.db")

	s, err := Open(path)
	require.NoError(t, err)

	require.NoError(t, s.Create(&Job{
		Company:  companyAcme,
		Position: positionEngineer,
	}))

	require.NoError(t, s.Close())

	// The underlying connection is really closed: further use must error,
	// not silently succeed (which would mean Close was a no-op).
	_, err = s.List()
	assert.Error(t, err)
}
