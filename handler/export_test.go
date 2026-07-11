package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tecnologer/jobtracker/store"
)

func TestExportCSV(t *testing.T) {
	t.Parallel()

	st, err := store.Open(filepath.Join(t.TempDir(), "jobs.db"))
	require.NoError(t, err)

	applied := time.Date(2026, 7, 1, 0, 0, 0, 0, time.FixedZone("", -6*3600))
	job := &store.Job{
		Company:   "Acme, Inc",  // comma must be quoted
		Position:  `Dev "lead"`, // quotes must be escaped
		Status:    store.StatusApplied,
		AppliedAt: &applied,
		URL:       "http://x",
		Notes:     "line1\nline2", // newline must be quoted
		TopMatch:  true,
	}
	require.NoError(t, st.Create(job))

	h := New(st)
	rec := httptest.NewRecorder()
	h.ExportCSV(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/jobs/export", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, strings.HasPrefix(rec.Header().Get("Content-Type"), "text/csv"), "Content-Type = %q, want text/csv", rec.Header().Get("Content-Type"))
	assert.Contains(t, rec.Header().Get("Content-Disposition"), "attachment")

	rows, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	require.NoError(t, err, "parse CSV")
	require.Len(t, rows, 2, "rows incl. header")

	header := []string{
		"ID", "Company", "Position", "Status", "Stage", "Applied At",
		"Archived", "Top Match", "URL", "Notes", "Created At", "Next Meeting",
	}
	assert.Equal(t, header, rows[0])

	got := rows[1]
	assert.Equal(t, "Acme, Inc", got[1])
	assert.Equal(t, `Dev "lead"`, got[2])
	assert.Equal(t, "line1\nline2", got[9])
	assert.Equal(t, string(store.StatusApplied), got[3])
	assert.Equal(t, "2026-07-01", got[5], "applied_at is a wall date, unaffected by viewer timezone")
	assert.Equal(t, yes, got[7], "TopMatch: true")
	assert.Empty(t, got[11], "no meetings scheduled for this job")
}

func TestExportCSVIDs(t *testing.T) {
	t.Parallel()

	st, err := store.Open(filepath.Join(t.TempDir(), "jobs.db"))
	require.NoError(t, err)

	first := &store.Job{
		Company:  "Acme",
		Position: "Engineer",
		Status:   store.StatusApplied,
	}
	second := &store.Job{
		Company:  "Globex",
		Position: "Manager",
		Status:   store.StatusApplied,
	}
	require.NoError(t, st.Create(first))
	require.NoError(t, st.Create(second))

	h := New(st)

	tests := []struct {
		name          string
		query         string
		wantCode      int
		wantCompanies []string
	}{
		{
			name:          "subset_of_ids_exports_only_those_jobs",
			query:         "?ids=" + strconv.FormatUint(uint64(second.ID), 10),
			wantCode:      http.StatusOK,
			wantCompanies: []string{"Globex"},
		},
		{
			name: "multiple_ids_with_whitespace_and_trailing_comma",
			query: "?ids=" + url.QueryEscape(fmt.Sprintf(" %d , %d ,",
				first.ID, second.ID)),
			wantCode:      http.StatusOK,
			wantCompanies: []string{"Acme", "Globex"},
		},
		{
			name:          "empty_ids_param_exports_no_jobs",
			query:         "?ids=",
			wantCode:      http.StatusOK,
			wantCompanies: []string{},
		},
		{
			name:          "unknown_id_yields_empty_export",
			query:         "?ids=999999",
			wantCode:      http.StatusOK,
			wantCompanies: []string{},
		},
		{
			name:     "non_numeric_ids_rejected",
			query:    "?ids=abc",
			wantCode: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			h.ExportCSV(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/jobs/export"+test.query, nil))

			require.Equal(t, test.wantCode, rec.Code)
			if test.wantCode != http.StatusOK {
				return
			}

			rows, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
			require.NoError(t, err, "parse CSV")
			companies := []string{}
			for _, row := range rows[1:] {
				companies = append(companies, row[1])
			}
			assert.Equal(t, test.wantCompanies, companies)
		})
	}
}

func TestExportCSVNextMeeting(t *testing.T) {
	t.Parallel()

	st, err := store.Open(filepath.Join(t.TempDir(), "jobs.db"))
	require.NoError(t, err)

	job := &store.Job{
		Company:  "Acme",
		Position: "Engineer",
		Status:   store.StatusApplied,
	}
	require.NoError(t, st.Create(job))

	past := time.Now().Add(-time.Hour)
	soonest := time.Now().Add(2 * time.Hour)
	later := time.Now().Add(48 * time.Hour)
	for _, meeting := range []*store.Meeting{
		{JobID: job.ID, Title: "Past", ScheduledAt: past},
		{JobID: job.ID, Title: "Later", ScheduledAt: later},
		{JobID: job.ID, Title: "Soonest", ScheduledAt: soonest},
	} {
		require.NoError(t, st.CreateMeeting(meeting))
	}

	h := New(st)
	rec := httptest.NewRecorder()
	h.ExportCSV(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/jobs/export", nil))

	rows, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	require.NoError(t, err, "parse CSV")
	require.Len(t, rows, 2, "rows incl. header")

	assert.Equal(t, soonest.Format(time.RFC3339), rows[1][11], "next meeting must be the soonest upcoming one")
}
