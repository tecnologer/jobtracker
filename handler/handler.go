package handler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/tecnologer/jobtracker/store"
)

const yes = "yes"

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

	if err := cw.Write([]string{
		"ID", "Company", "Position", "Status", "Stage", "Applied At",
		"Archived", "Top Match", "URL", "Notes", "Created At", "Next Meeting",
	}); err != nil {
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

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var j store.Job
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
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
