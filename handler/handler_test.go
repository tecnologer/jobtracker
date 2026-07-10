package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tecnologer/jobtracker/store"
)

func TestJobsCRUD(t *testing.T) {
	t.Parallel()

	mux, _ := newMux(t)

	rec := do(t, mux, http.MethodPost, "/api/jobs", `{"company":"Acme","position":"Dev","status":"applied"}`)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var job store.Job
	decode(t, rec, &job)
	require.NotZero(t, job.ID, "created job has no ID")

	rec = do(t, mux, http.MethodGet, "/api/jobs", "")
	require.Equal(t, http.StatusOK, rec.Code)
	var jobs []store.Job
	decode(t, rec, &jobs)
	require.Len(t, jobs, 1)
	assert.Equal(t, "Acme", jobs[0].Company)

	rec = do(t, mux, http.MethodPut, "/api/jobs/1", `{"company":"Acme","position":"Senior Dev","status":"in_progress"}`)
	require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())

	rec = do(t, mux, http.MethodPut, "/api/jobs/1/top-match", `{"top_match":true}`)
	require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())

	rec = do(t, mux, http.MethodGet, "/api/jobs", "")
	decode(t, rec, &jobs)
	require.Len(t, jobs, 1)
	assert.Equal(t, "Senior Dev", jobs[0].Position)
	assert.True(t, jobs[0].TopMatch)

	rec = do(t, mux, http.MethodDelete, "/api/jobs/1", "")
	require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())

	rec = do(t, mux, http.MethodGet, "/api/jobs", "")
	decode(t, rec, &jobs)
	assert.Empty(t, jobs, "job still listed after delete")
}

func TestBadRequests(t *testing.T) {
	t.Parallel()

	mux, st := newMux(t)
	// a real job so bad-body cases fail on the body, not a missing route param
	require.NoError(t, st.Create(&store.Job{
		Company:  "Acme",
		Position: "Dev",
		Status:   store.StatusApplied,
	}))

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "create_job_invalid_json", method: http.MethodPost, path: "/api/jobs", body: "{"},
		{name: "update_job_invalid_id", method: http.MethodPut, path: "/api/jobs/abc", body: `{}`},
		{name: "update_job_invalid_json", method: http.MethodPut, path: "/api/jobs/1", body: "{"},
		{name: "top_match_invalid_id", method: http.MethodPut, path: "/api/jobs/abc/top-match", body: `{}`},
		{name: "top_match_invalid_json", method: http.MethodPut, path: "/api/jobs/1/top-match", body: "{"},
		{name: "delete_job_invalid_id", method: http.MethodDelete, path: "/api/jobs/abc", body: ""},
		{name: "list_logs_invalid_id", method: http.MethodGet, path: "/api/jobs/abc/logs", body: ""},
		{name: "add_log_invalid_id", method: http.MethodPost, path: "/api/jobs/abc/logs", body: `{}`},
		{name: "add_log_invalid_json", method: http.MethodPost, path: "/api/jobs/1/logs", body: "{"},
		{name: "list_contacts_invalid_id", method: http.MethodGet, path: "/api/jobs/abc/contacts", body: ""},
		{name: "create_contact_invalid_id", method: http.MethodPost, path: "/api/jobs/abc/contacts", body: `{}`},
		{name: "create_contact_invalid_json", method: http.MethodPost, path: "/api/jobs/1/contacts", body: "{"},
		{name: "delete_contact_invalid_id", method: http.MethodDelete, path: "/api/jobs/1/contacts/abc", body: ""},
		{name: "create_default_stage_invalid_json", method: http.MethodPost, path: "/api/stages", body: "{"},
		{name: "list_stages_invalid_id", method: http.MethodGet, path: "/api/jobs/abc/stages", body: ""},
		{name: "create_stage_invalid_id", method: http.MethodPost, path: "/api/jobs/abc/stages", body: `{}`},
		{name: "create_stage_invalid_json", method: http.MethodPost, path: "/api/jobs/1/stages", body: "{"},
		{name: "update_stage_invalid_id", method: http.MethodPut, path: "/api/stages/abc", body: `{}`},
		{name: "update_stage_invalid_json", method: http.MethodPut, path: "/api/stages/1", body: "{"},
		{name: "delete_stage_invalid_id", method: http.MethodDelete, path: "/api/stages/abc", body: ""},
		{name: "list_meetings_invalid_id", method: http.MethodGet, path: "/api/jobs/abc/meetings", body: ""},
		{name: "create_meeting_invalid_id", method: http.MethodPost, path: "/api/jobs/abc/meetings", body: `{}`},
		{name: "create_meeting_invalid_json", method: http.MethodPost, path: "/api/jobs/1/meetings", body: "{"},
		{name: "create_meeting_missing_fields", method: http.MethodPost, path: "/api/jobs/1/meetings", body: `{"title":""}`},
		{name: "update_meeting_invalid_id", method: http.MethodPut, path: "/api/jobs/1/meetings/abc", body: `{}`},
		{name: "update_meeting_invalid_json", method: http.MethodPut, path: "/api/jobs/1/meetings/1", body: "{"},
		{name: "update_meeting_missing_fields", method: http.MethodPut, path: "/api/jobs/1/meetings/1", body: `{"title":"x"}`},
		{name: "delete_meeting_invalid_id", method: http.MethodDelete, path: "/api/jobs/1/meetings/abc", body: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			rec := do(t, mux, test.method, test.path, test.body)
			assert.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
		})
	}
}

func TestStages(t *testing.T) {
	t.Parallel()

	mux, st := newMux(t)
	require.NoError(t, st.Create(&store.Job{
		Company:  "Acme",
		Position: "Dev",
		Status:   store.StatusApplied,
	}))

	rec := do(t, mux, http.MethodGet, "/api/stages", "")
	require.Equal(t, http.StatusOK, rec.Code)
	var defaults []store.Stage
	decode(t, rec, &defaults)
	require.Len(t, defaults, 5, "seeded default stages")

	rec = do(t, mux, http.MethodPost, "/api/stages", `{"name":"Culture Fit","sort_order":6}`)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var created store.Stage
	decode(t, rec, &created)
	assert.Zero(t, created.JobID, "default stage must have job_id 0")

	rec = do(t, mux, http.MethodGet, "/api/jobs/1/stages", "")
	var jobStages []store.Stage
	decode(t, rec, &jobStages)
	require.Len(t, jobStages, 5, "stages cloned from defaults on job create")

	rec = do(t, mux, http.MethodPost, "/api/jobs/1/stages", `{"name":"Extra Round","sort_order":9}`)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var jobStage store.Stage
	decode(t, rec, &jobStage)
	assert.Equal(t, uint(1), jobStage.JobID, "job_id must come from the path, not the body")

	rec = do(t, mux, http.MethodPut, "/api/stages/"+itoa(jobStage.ID), `{"name":"Renamed","sort_order":9}`)
	require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())

	rec = do(t, mux, http.MethodDelete, "/api/stages/"+itoa(jobStage.ID), "")
	require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())

	rec = do(t, mux, http.MethodGet, "/api/jobs/1/stages", "")
	decode(t, rec, &jobStages)
	for _, stage := range jobStages {
		assert.NotEqual(t, "Extra Round", stage.Name, "deleted stage still listed")
		if stage.Name == "Renamed" {
			assert.Equal(t, jobStage.ID, stage.ID, "rename hit the wrong stage")
		}
	}
}

func TestContacts(t *testing.T) {
	t.Parallel()

	mux, st := newMux(t)
	require.NoError(t, st.Create(&store.Job{
		Company:  "Acme",
		Position: "Dev",
		Status:   store.StatusApplied,
	}))

	rec := do(t, mux, http.MethodPost, "/api/jobs/1/contacts", `{"name":"Jane","role":"Recruiter","email":"jane@acme.test"}`)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var contact store.Contact
	decode(t, rec, &contact)
	assert.Equal(t, uint(1), contact.JobID, "job_id must come from the path")

	rec = do(t, mux, http.MethodGet, "/api/jobs/1/contacts", "")
	require.Equal(t, http.StatusOK, rec.Code)
	var contacts []store.Contact
	decode(t, rec, &contacts)
	require.Len(t, contacts, 1)
	assert.Equal(t, "Jane", contacts[0].Name)

	rec = do(t, mux, http.MethodDelete, "/api/jobs/1/contacts/"+itoa(contact.ID), "")
	require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())

	rec = do(t, mux, http.MethodGet, "/api/jobs/1/contacts", "")
	decode(t, rec, &contacts)
	assert.Empty(t, contacts, "contact still listed after delete")
}

func TestStageLogs(t *testing.T) {
	t.Parallel()

	mux, st := newMux(t)
	job := &store.Job{
		Company:  "Acme",
		Position: "Dev",
		Status:   store.StatusApplied,
	}
	require.NoError(t, st.Create(job))
	stages, err := st.ListStages(job.ID)
	require.NoError(t, err)
	require.NotEmpty(t, stages)
	target := stages[0]

	rec := do(t, mux, http.MethodPost, "/api/jobs/1/logs", `{"stage_id":`+itoa(target.ID)+`,"notes":"moved to phone screen"}`)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	// invariant: adding a log also updates the job's current stage
	rec = do(t, mux, http.MethodGet, "/api/jobs", "")
	var jobs []store.Job
	decode(t, rec, &jobs)
	require.Len(t, jobs, 1)
	require.NotNil(t, jobs[0].StageID)
	assert.Equal(t, target.ID, *jobs[0].StageID, "job stage_id not updated by log")

	rec = do(t, mux, http.MethodGet, "/api/jobs/1/logs", "")
	require.Equal(t, http.StatusOK, rec.Code)
	var logs []store.StageLog
	decode(t, rec, &logs)
	require.Len(t, logs, 1)
	assert.Equal(t, "moved to phone screen", logs[0].Notes)
}

func TestMeetings(t *testing.T) {
	t.Parallel()

	mux, st := newMux(t)
	require.NoError(t, st.Create(&store.Job{
		Company:  "Acme",
		Position: "Dev",
		Status:   store.StatusApplied,
	}))

	scheduled := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	rec := do(t, mux, http.MethodPost, "/api/jobs/1/meetings", `{"title":"Phone Screen","scheduled_at":"`+scheduled+`"}`)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var meeting store.Meeting
	decode(t, rec, &meeting)
	assert.Equal(t, uint(1), meeting.JobID, "job_id must come from the path")

	rec = do(t, mux, http.MethodGet, "/api/jobs/1/meetings", "")
	require.Equal(t, http.StatusOK, rec.Code)
	var meetings []store.Meeting
	decode(t, rec, &meetings)
	require.Len(t, meetings, 1)
	assert.Equal(t, "Phone Screen", meetings[0].Title)

	rec = do(t, mux, http.MethodPut, "/api/jobs/1/meetings/"+itoa(meeting.ID), `{"title":"Final Round","scheduled_at":"`+scheduled+`"}`)
	require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())

	rec = do(t, mux, http.MethodGet, "/api/meetings/upcoming?limit=1", "")
	require.Equal(t, http.StatusOK, rec.Code)
	decode(t, rec, &meetings)
	require.Len(t, meetings, 1)
	assert.Equal(t, "Final Round", meetings[0].Title)

	rec = do(t, mux, http.MethodDelete, "/api/jobs/1/meetings/"+itoa(meeting.ID), "")
	require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())

	rec = do(t, mux, http.MethodGet, "/api/jobs/1/meetings", "")
	decode(t, rec, &meetings)
	assert.Empty(t, meetings, "meeting still listed after delete")
}

// newMux builds a Handler over a fresh temp-file store and registers the same
// route patterns as main.go, so r.PathValue works in tests.
func TestStats(t *testing.T) {
	t.Parallel()

	mux, st := newMux(t)

	rec := do(t, mux, http.MethodPost, "/api/jobs", `{"company":"Acme","position":"Dev","status":"applied","applied_at":"2026-01-02T00:00:00-06:00"}`)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var job store.Job
	decode(t, rec, &job)

	stages, err := st.ListStages(job.ID)
	require.NoError(t, err)
	require.NotEmpty(t, stages)
	rec = do(t, mux, http.MethodPost, "/api/jobs/"+itoa(job.ID)+"/logs", `{"stage_id":`+itoa(stages[0].ID)+`}`)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	rec = do(t, mux, http.MethodGet, "/api/stats", "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	body := rec.Body.String() // capture before decode drains the buffer
	var stats store.Stats
	decode(t, rec, &stats)
	assert.Equal(t, 1, stats.TotalJobs)
	assert.Equal(t, 1, stats.ActiveJobs)
	assert.Equal(t, 1, stats.StatusBreakdown[store.StatusApplied])
	require.NotEmpty(t, stats.Funnel)
	assert.Equal(t, 1, stats.Funnel[0].JobsReached)

	// wire shape: snake_case keys the frontend contract relies on
	for _, key := range []string{"total_jobs", "active_jobs", "offers", "rejection_rate", "avg_days_to_first_response", "status_breakdown", "funnel", "jobs_reached", "avg_days", "sort_order"} {
		assert.Contains(t, body, `"`+key+`"`)
	}
}

func newMux(t *testing.T) (*http.ServeMux, *store.Store) {
	t.Helper()

	st, err := store.Open(filepath.Join(t.TempDir(), "jobs.db"))
	require.NoError(t, err)
	h := New(st)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/jobs", h.List)
	mux.HandleFunc("GET /api/stats", h.Stats)
	mux.HandleFunc("GET /api/jobs/export", h.ExportCSV)
	mux.HandleFunc("POST /api/jobs", h.Create)
	mux.HandleFunc("PUT /api/jobs/{id}", h.Update)
	mux.HandleFunc("PUT /api/jobs/{id}/top-match", h.SetTopMatch)
	mux.HandleFunc("DELETE /api/jobs/{id}", h.Delete)
	mux.HandleFunc("GET /api/jobs/{id}/logs", h.ListStageLogs)
	mux.HandleFunc("POST /api/jobs/{id}/logs", h.AddStageLog)
	mux.HandleFunc("GET /api/jobs/{id}/contacts", h.ListContacts)
	mux.HandleFunc("POST /api/jobs/{id}/contacts", h.CreateContact)
	mux.HandleFunc("DELETE /api/jobs/{id}/contacts/{cid}", h.DeleteContact)
	mux.HandleFunc("GET /api/jobs/{id}/meetings", h.ListMeetings)
	mux.HandleFunc("POST /api/jobs/{id}/meetings", h.CreateMeeting)
	mux.HandleFunc("PUT /api/jobs/{id}/meetings/{mid}", h.UpdateMeeting)
	mux.HandleFunc("DELETE /api/jobs/{id}/meetings/{mid}", h.DeleteMeeting)
	mux.HandleFunc("GET /api/meetings/upcoming", h.ListUpcomingMeetings)
	mux.HandleFunc("GET /api/stages", h.ListDefaultStages)
	mux.HandleFunc("POST /api/stages", h.CreateDefaultStage)
	mux.HandleFunc("GET /api/jobs/{id}/stages", h.ListStages)
	mux.HandleFunc("POST /api/jobs/{id}/stages", h.CreateStage)
	mux.HandleFunc("PUT /api/stages/{id}", h.UpdateStage)
	mux.HandleFunc("DELETE /api/stages/{id}", h.DeleteStage)
	return mux, st
}

func do(t *testing.T, mux *http.ServeMux, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func decode(t *testing.T, rec *httptest.ResponseRecorder, v any) {
	t.Helper()

	require.NoError(t, json.NewDecoder(rec.Body).Decode(v), "decoding response %q", rec.Body)
}

func itoa(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}
