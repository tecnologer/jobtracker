package handler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/tecnologer/jobtracker/store"
)

const yes = "yes"

// Import caps (NFR-01): both are enforced before any row is processed.
const (
	maxImportBytes = 1 << 20 // 1 MB
	maxImportRows  = 5000
)

// csvHeader is the CSV header contract shared by ExportCSV (writes it) and
// ImportCSV (rejects any file whose header doesn't match it exactly).
func csvHeader() []string {
	return []string{
		"ID", "Company", "Position", "Status", "Stage", "Applied At",
		"Archived", "Top Match", "URL", "Notes", "Created At", "Next Meeting",
	}
}

type Handler struct {
	store *store.Store
}

func New(s *store.Store) *Handler {
	return &Handler{store: s}
}

func (h *Handler) List(w http.ResponseWriter, _ *http.Request) {
	jobs, err := h.store.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	writeJSON(w, http.StatusOK, jobs)
}

// writeJSON sets the JSON content type, writes the status, and encodes v.
// The status line is already sent by the time encoding can fail, so the
// error can only be logged, not turned into an HTTP error response.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encoding JSON response: %v", err)
	}
}

func (h *Handler) Stats(w http.ResponseWriter, _ *http.Request) {
	stats, err := h.store.Stats(time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// ExportCSV streams jobs as a CSV attachment. An optional ids query param
// (comma-separated job IDs) restricts the export to those jobs — the frontend
// sends the currently visible (filtered) grid, whose filter logic lives
// client-side only. A present-but-empty ids param means an empty grid and
// yields a header-only export; an absent param exports everything. applied_at
// is rendered as a wall date (YYYY-MM-DD) in its own stored offset so the day
// never shifts for the viewer's timezone; created_at is a real instant,
// rendered in full RFC3339. Next Meeting is each job's soonest upcoming
// meeting (scheduled_at >= now), computed with a single grouped query rather
// than one lookup per row.
func (h *Handler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	var keep map[uint]bool // nil: no ids param, export everything
	if rawIDs, ok := r.URL.Query()["ids"]; ok {
		var err error
		keep, err = parseIDSet(rawIDs[0])
		if err != nil {
			http.Error(w, "invalid ids parameter", http.StatusBadRequest)

			return
		}
	}

	jobs, err := h.store.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// ponytail: filters in memory after List(); push into the store query if job counts ever get big
	if keep != nil {
		jobs = filterJobs(jobs, keep)
	}

	nextMeetings, err := h.store.NextMeetingTimes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="jobs-`+time.Now().Format("2006-01-02")+`.csv"`)

	cw := csv.NewWriter(w)
	defer cw.Flush()

	if err := cw.Write(csvHeader()); err != nil {
		return
	}

	for _, job := range jobs {
		if err := cw.Write(csvRow(job, nextMeetings)); err != nil {
			return
		}
	}
}

// filterJobs keeps only the jobs whose ID is in keep. Filters in place.
func filterJobs(jobs []store.Job, keep map[uint]bool) []store.Job {
	kept := jobs[:0]
	for _, job := range jobs {
		if keep[job.ID] {
			kept = append(kept, job)
		}
	}
	return kept
}

// parseIDSet parses a comma-separated ID list (whitespace tolerated, empty
// segments skipped) into a membership set. An empty string returns an empty
// set that matches nothing — the frontend sends the visible grid, which can
// be empty.
func parseIDSet(rawIDs string) (map[uint]bool, error) {
	keep := map[uint]bool{}
	for part := range strings.SplitSeq(rawIDs, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.ParseUint(part, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("parsing job id %q: %w", part, err)
		}
		keep[uint(id)] = true
	}
	return keep, nil
}

// csvRow renders one job as a CSV record; see ExportCSV for column semantics.
func csvRow(job store.Job, nextMeetings map[uint]time.Time) []string {
	appliedAt := ""
	if job.AppliedAt != nil {
		appliedAt = job.AppliedAt.Format("2006-01-02")
	}
	stageName := ""
	if job.Stage != nil {
		stageName = job.Stage.Name
	}
	archived := ""
	if job.ArchivedAt != nil {
		archived = yes
	}
	topMatch := ""
	if job.TopMatch {
		topMatch = yes
	}
	nextMeeting := ""
	if scheduledAt, ok := nextMeetings[job.ID]; ok {
		nextMeeting = scheduledAt.Format(time.RFC3339)
	}
	return []string{
		strconv.FormatUint(uint64(job.ID), 10),
		job.Company,
		job.Position,
		string(job.Status),
		stageName,
		appliedAt,
		archived,
		topMatch,
		job.URL,
		job.Notes,
		job.CreatedAt.Format(time.RFC3339),
		nextMeeting,
	}
}

// importJobView is the wire shape of a parsed-but-not-created duplicate row:
// the CSV Stage column is reported by name (raw, untrimmed match target) since
// the client resolves it against the target job's own stage list.
type importJobView struct {
	Company   string     `json:"company"`
	Position  string     `json:"position"`
	Status    string     `json:"status"`
	AppliedAt *time.Time `json:"applied_at"`
	Stage     string     `json:"stage"`
	Archived  bool       `json:"archived"`
	TopMatch  bool       `json:"top_match"`
	URL       string     `json:"url"`
	Notes     string     `json:"notes"`
}

type importDuplicate struct {
	Row      int           `json:"row"`
	Job      importJobView `json:"job"`
	Existing store.Job     `json:"existing"`
}

type importIssue struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

type importResult struct {
	Created       int               `json:"created"`
	StagesCreated int               `json:"stages_created"`
	Duplicates    []importDuplicate `json:"duplicates"`
	Errors        []importIssue     `json:"errors"`
	Warnings      []importIssue     `json:"warnings"`
}

// ImportCSV accepts a multipart file (field "file") produced by ExportCSV and
// creates jobs from it, best-effort per row (FR-05): rows that fail to parse
// don't block valid rows, and are reported by 1-based row number (header = row
// 1). Rows that FindDuplicate matches — including rows created earlier in the
// same file, since store.Create commits immediately — are not created; they
// are returned for the client to resolve via the existing job endpoints (see
// REQUIREMENTS.md phase 2). A header or size/row-count mismatch rejects the
// whole file with 400 before any row is processed (FR-01, NFR-01).
func (h *Handler) ImportCSV(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxImportBytes)
	//nolint:gosec // G120 false positive: r.Body is already capped by MaxBytesReader above (NFR-01)
	if err := r.ParseMultipartForm(maxImportBytes); err != nil {
		http.Error(w, "file too large (max 1 MB) or invalid upload", http.StatusBadRequest)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	cr := csv.NewReader(file)
	cr.FieldsPerRecord = -1 // rows with the wrong column count are row errors, not a rejected file

	header, err := cr.Read()
	if err != nil {
		http.Error(w, "invalid CSV file", http.StatusBadRequest)
		return
	}
	if !slices.Equal(header, csvHeader()) {
		http.Error(w, "CSV header does not match the expected export format", http.StatusBadRequest)
		return
	}

	rows, err := cr.ReadAll()
	if err != nil {
		http.Error(w, "invalid CSV file", http.StatusBadRequest)
		return
	}
	if len(rows) > maxImportRows {
		http.Error(w, fmt.Sprintf("file has %d data rows, exceeding the %d row limit", len(rows), maxImportRows), http.StatusBadRequest)
		return
	}

	result := importResult{
		Duplicates: []importDuplicate{},
		Errors:     []importIssue{},
		Warnings:   []importIssue{},
	}
	now := time.Now()
	for i, row := range rows {
		rowNum := i + 2 // header is row 1, first data row is 2
		h.importRow(&result, row, rowNum, now)
	}

	writeJSON(w, http.StatusOK, result)
}

// importRow parses and processes a single CSV data row, appending to result.
func (h *Handler) importRow(result *importResult, row []string, rowNum int, now time.Time) {
	job, stageName, err := parseImportRow(row, now)
	if err != nil {
		result.Errors = append(result.Errors, importIssue{Row: rowNum, Message: err.Error()})
		return
	}

	dup, err := h.store.FindDuplicate(&job)
	if err != nil {
		result.Errors = append(result.Errors, importIssue{Row: rowNum, Message: err.Error()})
		return
	}
	if dup != nil {
		result.Duplicates = append(result.Duplicates, importDuplicate{
			Row:      rowNum,
			Job:      toImportJobView(job, stageName),
			Existing: *dup,
		})
		return
	}

	if err := h.store.Create(&job); err != nil {
		result.Errors = append(result.Errors, importIssue{Row: rowNum, Message: err.Error()})
		return
	}
	result.Created++

	if stageName == "" {
		return
	}
	stages, err := h.store.ListStages(job.ID)
	if err != nil {
		result.Errors = append(result.Errors, importIssue{Row: rowNum, Message: err.Error()})
		return
	}
	matched := matchStage(stages, stageName)
	if matched == nil {
		created := store.Stage{JobID: job.ID, Name: stageName, SortOrder: len(stages)}
		if err := h.store.CreateStage(&created); err != nil {
			result.Errors = append(result.Errors, importIssue{Row: rowNum, Message: err.Error()})
			return
		}
		matched = &created
		result.StagesCreated++
	}
	// AddStageLog is the atomic helper (StageLog row + jobs.stage_id) shared with
	// POST /api/jobs/{id}/logs, preserving the stage-log invariant on import too.
	if err := h.store.AddStageLog(job.ID, &store.StageLog{StageID: &matched.ID}); err != nil {
		result.Errors = append(result.Errors, importIssue{Row: rowNum, Message: err.Error()})
	}
}

// parseImportRow validates and converts one CSV data row into a Job plus its
// raw (trimmed) Stage column, applying the column rules in REQUIREMENTS.md
// section 4. ID, Created At, and Next Meeting are ignored.
func parseImportRow(row []string, now time.Time) (store.Job, string, error) {
	wantFields := len(csvHeader())
	if len(row) != wantFields {
		return store.Job{}, "", fmt.Errorf("row has %d fields, want %d", len(row), wantFields)
	}

	company := strings.TrimSpace(row[1])
	if company == "" {
		return store.Job{}, "", fmt.Errorf("company is required")
	}
	position := strings.TrimSpace(row[2])
	if position == "" {
		return store.Job{}, "", fmt.Errorf("position is required")
	}

	status := strings.TrimSpace(row[3])
	if status == "" {
		status = string(store.StatusProspect)
	} else if !validStatus(status) {
		return store.Job{}, "", fmt.Errorf("invalid status %q", status)
	}

	var appliedAt *time.Time
	if raw := strings.TrimSpace(row[5]); raw != "" {
		// wall date: parsed with no zone info, time.Parse defaults to UTC, i.e.
		// stored at 00:00:00Z (see the applied_at invariant in CLAUDE.md).
		parsed, err := time.Parse(time.DateOnly, raw)
		if err != nil {
			return store.Job{}, "", fmt.Errorf("invalid applied at date %q, want YYYY-MM-DD", raw)
		}
		appliedAt = &parsed
	}

	var archivedAt *time.Time
	if strings.EqualFold(strings.TrimSpace(row[6]), yes) {
		archivedAt = &now
	}

	job := store.Job{
		Company:    company,
		Position:   position,
		Status:     store.ApplicationStatus(status),
		AppliedAt:  appliedAt,
		ArchivedAt: archivedAt,
		TopMatch:   strings.EqualFold(strings.TrimSpace(row[7]), yes),
		URL:        row[8],
		Notes:      row[9],
	}
	return job, strings.TrimSpace(row[4]), nil
}

func validStatus(s string) bool {
	switch store.ApplicationStatus(s) {
	case store.StatusProspect, store.StatusApplied, store.StatusInProgress, store.StatusOnHold,
		store.StatusNegotiating, store.StatusAccepted, store.StatusRejected, store.StatusCanceled:
		return true
	default:
		return false
	}
}

// matchStage finds a job's cloned stage by name, case/trim-insensitively.
// No match returns nil; the import caller then creates the stage for the job.
func matchStage(stages []store.Stage, name string) *store.Stage {
	name = strings.ToLower(strings.TrimSpace(name))
	for i := range stages {
		if strings.ToLower(strings.TrimSpace(stages[i].Name)) == name {
			return &stages[i]
		}
	}
	return nil
}

func toImportJobView(j store.Job, stageName string) importJobView {
	return importJobView{
		Company:   j.Company,
		Position:  j.Position,
		Status:    string(j.Status),
		AppliedAt: j.AppliedAt,
		Stage:     stageName,
		Archived:  j.ArchivedAt != nil,
		TopMatch:  j.TopMatch,
		URL:       j.URL,
		Notes:     j.Notes,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var j store.Job
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if r.URL.Query().Get("allow_duplicate") != "1" {
		dup, err := h.store.FindDuplicate(&j)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if dup != nil {
			writeJSON(w, http.StatusConflict, map[string]any{"error": "duplicate", "duplicate": dup})
			return
		}
	}
	if err := h.store.Create(&j); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, j)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var j store.Job
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.Update(uint(id), &j); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) SetTopMatch(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var body struct {
		TopMatch bool `json:"top_match"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.SetTopMatch(uint(id), body.TopMatch); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.store.Delete(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListDefaultStages(w http.ResponseWriter, _ *http.Request) {
	stages, err := h.store.ListDefaultStages()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, stages)
}

func (h *Handler) CreateDefaultStage(w http.ResponseWriter, r *http.Request) {
	var stage store.Stage
	if err := json.NewDecoder(r.Body).Decode(&stage); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.CreateDefaultStage(&stage); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, stage)
}

func (h *Handler) ListStages(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	stages, err := h.store.ListStages(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, stages)
}

func (h *Handler) CreateStage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var stage store.Stage
	if err := json.NewDecoder(r.Body).Decode(&stage); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	stage.JobID = uint(id)
	if err := h.store.CreateStage(&stage); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, stage)
}

func (h *Handler) UpdateStage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var stage store.Stage
	if err := json.NewDecoder(r.Body).Decode(&stage); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.UpdateStage(uint(id), &stage); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteStage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.store.DeleteStage(uint(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListContacts(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	contacts, err := h.store.ListContacts(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, contacts)
}

func (h *Handler) CreateContact(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var c store.Contact
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	c.JobID = uint(id)
	if err := h.store.CreateContact(&c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (h *Handler) DeleteContact(w http.ResponseWriter, r *http.Request) {
	cid, err := strconv.ParseUint(r.PathValue("cid"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.store.DeleteContact(uint(cid)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListStageLogs(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	logs, err := h.store.ListStageLogs(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, logs)
}

func (h *Handler) AddStageLog(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var stageLog store.StageLog
	if err := json.NewDecoder(r.Body).Decode(&stageLog); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.store.AddStageLog(uint(id), &stageLog); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) ListMeetings(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	meetings, err := h.store.ListMeetings(uint(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, meetings)
}

func (h *Handler) CreateMeeting(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var m store.Meeting
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if m.Title == "" || m.ScheduledAt.IsZero() {
		http.Error(w, "title and scheduled_at are required", http.StatusBadRequest)
		return
	}
	m.JobID = uint(id)
	if err := h.store.CreateMeeting(&m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

func (h *Handler) UpdateMeeting(w http.ResponseWriter, r *http.Request) {
	mid, err := strconv.ParseUint(r.PathValue("mid"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var m store.Meeting
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if m.Title == "" || m.ScheduledAt.IsZero() {
		http.Error(w, "title and scheduled_at are required", http.StatusBadRequest)
		return
	}
	if err := h.store.UpdateMeeting(uint(mid), &m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteMeeting(w http.ResponseWriter, r *http.Request) {
	mid, err := strconv.ParseUint(r.PathValue("mid"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.store.DeleteMeeting(uint(mid)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListUpcomingMeetings(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	meetings, err := h.store.ListUpcomingMeetings(limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, meetings)
}
