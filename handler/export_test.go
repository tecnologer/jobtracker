package handler

import (
	"encoding/csv"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tecnologer/jobtracker/store"
)

func TestExportCSV(t *testing.T) {
	t.Parallel()

	st, err := store.Open(filepath.Join(t.TempDir(), "jobs.db"))
	if err != nil {
		t.Fatal("Open:", err)
	}

	applied := time.Date(2026, 7, 1, 0, 0, 0, 0, time.FixedZone("", -6*3600))
	job := &store.Job{
		Company:   "Acme, Inc",  // comma must be quoted
		Position:  `Dev "lead"`, // quotes must be escaped
		Status:    store.StatusApplied,
		AppliedAt: &applied,
		URL:       "http://x",
		Notes:     "line1\nline2", // newline must be quoted
	}
	if err := st.Create(job); err != nil {
		t.Fatal("Create:", err)
	}

	h := New(st)
	rec := httptest.NewRecorder()
	h.ExportCSV(rec, httptest.NewRequest(http.MethodGet, "/api/jobs/export", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/csv") {
		t.Errorf("Content-Type = %q, want text/csv", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "attachment") {
		t.Errorf("Content-Disposition = %q, want attachment", cd)
	}

	rows, err := csv.NewReader(strings.NewReader(rec.Body.String())).ReadAll()
	if err != nil {
		t.Fatal("parse CSV:", err)
	}
	if len(rows) != 2 {
		t.Fatalf("got %d rows (incl. header), want 2", len(rows))
	}

	header := []string{"ID", "Company", "Position", "Status", "Stage", "Applied At", "Archived", "URL", "Notes", "Created At"}
	if strings.Join(rows[0], ",") != strings.Join(header, ",") {
		t.Errorf("header = %v, want %v", rows[0], header)
	}

	got := rows[1]
	if got[1] != "Acme, Inc" || got[2] != `Dev "lead"` || got[8] != "line1\nline2" {
		t.Errorf("special chars not round-tripped: %q", got)
	}
	if got[3] != string(store.StatusApplied) {
		t.Errorf("status = %q, want %q", got[3], store.StatusApplied)
	}
	if got[5] != "2026-07-01" { // wall date, unaffected by viewer timezone
		t.Errorf("applied_at = %q, want 2026-07-01", got[5])
	}
}
