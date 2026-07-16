package store

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindDuplicate(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "jobs.db")

	store, err := Open(path)
	require.NoError(t, err)

	recent := time.Now().AddDate(0, -1, 0)
	require.NoError(t, store.Create(&Job{Company: companyAcme, Position: positionEngineer, AppliedAt: &recent}))

	// case/whitespace-insensitive match within 6 months
	dup, err := store.FindDuplicate(&Job{Company: " ACME ", Position: "engineer"})
	require.NoError(t, err)
	require.NotNil(t, dup)
	assert.Equal(t, companyAcme, dup.Company)

	// different position is not a duplicate
	dup, err = store.FindDuplicate(&Job{Company: companyAcme, Position: "Designer"})
	require.NoError(t, err)
	assert.Nil(t, dup)

	// same company+position applied over 6 months ago is not a duplicate
	old := time.Now().AddDate(0, -7, 0)
	require.NoError(t, store.Create(&Job{Company: "Globex", Position: positionEngineer, AppliedAt: &old}))

	dup, err = store.FindDuplicate(&Job{Company: "Globex", Position: positionEngineer})
	require.NoError(t, err)
	assert.Nil(t, dup)
}
