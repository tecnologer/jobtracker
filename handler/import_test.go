package handler

import (
	"bytes"
	"encoding/csv"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tecnologer/jobtracker/store"
)

func TestImportCSVHeaderMismatch(t *testing.T) {
	t.Parallel()

	mux, st := newMux(t)

	// "Applied Date" instead of "Applied At": renamed column (FR-01).
	body := "ID,Company,Position,Status,Stage,Applied Date,Archived,Top Match,URL,Notes,Created At,Next Meeting\n" +
		"1,Acme,Engineer,applied,,,,,,,,\n"

	rec := doImport(t, mux, body)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	jobs, err := st.List()
	require.NoError(t, err)
	assert.Empty(t, jobs, "no rows created when the header is rejected")
}

func TestImportCSVMissingFile(t *testing.T) {
	t.Parallel()

	mux, _ := newMux(t)

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	require.NoError(t, mw.WriteField("not_file", "x"))
	require.NoError(t, mw.Close())

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/jobs/import", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestImportCSVRows covers FR-02/03/04/05: valid rows are created (with a
// matched stage getting a StageLog, an unmatched stage auto-created for the
// job, and an empty stage staying unset with no warning), while invalid rows
// are reported by row number without blocking the valid ones.
func TestImportCSVRows(t *testing.T) {
	t.Parallel()

	mux, st := newMux(t)

	body := buildImportCSV(t,
		[]string{"", "Acme", "Engineer", "applied", "Phone Screen", "2026-01-15", "yes", "yes", "http://acme.test", "great fit", "", ""},
		[]string{"", "Globex", "Manager", "", "", "", "", "", "", "", "", ""},
		[]string{"", "Initech", "Lead", "applied", "Culture Fit", "", "", "", "", "", "", ""},
		[]string{"", "", "Dev", "applied", "", "", "", "", "", "", "", ""},
		[]string{"", "Bad Status Co", "Dev", "bogus", "", "", "", "", "", "", "", ""},
		[]string{"", "Bad Date Co", "Dev", "applied", "", "not-a-date", "", "", "", "", "", ""},
		[]string{"", "Short Row Co", "Dev"},
	)

	rec := doImport(t, mux, body)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var result importResult
	decode(t, rec, &result)

	assert.Equal(t, 3, result.Created, "3 valid rows out of 7")
	assert.Equal(t, 1, result.StagesCreated, "only Culture Fit was missing from the job's stages")
	assert.Empty(t, result.Warnings, "unmatched stage is auto-created, not warned")

	require.Len(t, result.Errors, 4)
	wantErrorRows := []int{5, 6, 7, 8}
	for i, err := range result.Errors {
		assert.Equal(t, wantErrorRows[i], err.Row)
	}

	jobs, err := st.List()
	require.NoError(t, err)
	require.Len(t, jobs, 3)

	byCompany := map[string]store.Job{}
	for _, job := range jobs {
		byCompany[job.Company] = job
	}

	acme := byCompany["Acme"]
	require.NotNil(t, acme.StageID, "stage name matched a cloned stage")
	assert.Equal(t, "Phone Screen", acme.Stage.Name)
	assert.True(t, acme.TopMatch)
	require.NotNil(t, acme.ArchivedAt)
	require.NotNil(t, acme.AppliedAt)
	assert.Equal(t, "2026-01-15", acme.AppliedAt.Format(time.DateOnly), "applied_at round-trips as a wall date")
	assert.Equal(t, 0, acme.AppliedAt.Hour(), "applied_at is stored at midnight")

	logs, err := st.ListStageLogs(acme.ID)
	require.NoError(t, err)
	require.Len(t, logs, 1, "exactly one StageLog for the matched stage")

	globex := byCompany["Globex"]
	assert.Equal(t, store.StatusProspect, globex.Status, "empty status defaults to prospect")
	assert.Nil(t, globex.StageID, "empty stage column stays unset with no warning")

	initech := byCompany["Initech"]
	require.NotNil(t, initech.StageID, "unmatched stage name is auto-created for the job")
	assert.Equal(t, "Culture Fit", initech.Stage.Name)
	assert.Equal(t, initech.ID, initech.Stage.JobID, "created stage belongs to the job, not the defaults")
}

// TestImportCSVDuplicates covers FR-06: duplicate rows (including ones that
// duplicate a row created earlier in the same file) are reported for
// resolution instead of being created.
func TestImportCSVDuplicates(t *testing.T) {
	t.Parallel()

	mux, st := newMux(t)

	body := buildImportCSV(t,
		[]string{"", "Acme", "Engineer", "applied", "", "", "", "", "", "", "", ""},
		[]string{"", "Acme", "Engineer", "applied", "", "", "", "", "", "", "", ""}, // in-file duplicate of the row above
	)

	rec := doImport(t, mux, body)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var result importResult
	decode(t, rec, &result)
	assert.Equal(t, 1, result.Created)
	require.Len(t, result.Duplicates, 1)
	assert.Equal(t, 3, result.Duplicates[0].Row, "second data row is row 3")
	assert.Equal(t, "Acme", result.Duplicates[0].Job.Company)

	jobs, err := st.List()
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	assert.Equal(t, jobs[0].ID, result.Duplicates[0].Existing.ID)

	// importing the same file again: the job created above now duplicates every row
	rec = doImport(t, mux, body)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var result2 importResult
	decode(t, rec, &result2)
	assert.Equal(t, 0, result2.Created)
	assert.Len(t, result2.Duplicates, 2)
}

// buildImportCSV writes the shared csvHeader() plus the given data rows as a
// CSV document, using encoding/csv so field escaping matches ImportCSV's own reader.
func buildImportCSV(t *testing.T, rows ...[]string) string {
	t.Helper()

	var buf bytes.Buffer
	cw := csv.NewWriter(&buf)
	require.NoError(t, cw.Write(csvHeader()))
	for _, row := range rows {
		require.NoError(t, cw.Write(row))
	}
	cw.Flush()
	require.NoError(t, cw.Error())
	return buf.String()
}

// doImport posts csvBody as a multipart/form-data upload under the "file"
// field, matching what ImportCSV expects.
func doImport(t *testing.T, mux *http.ServeMux, csvBody string) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("file", "jobs.csv")
	require.NoError(t, err)
	_, err = fw.Write([]byte(csvBody))
	require.NoError(t, err)
	require.NoError(t, mw.Close())

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/jobs/import", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}
