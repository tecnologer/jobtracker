package handler

import (
	"net/http"
	"strings"
)

// Routes registers every API handler plus the given static handler as the
// SPA fallback. static serves the built frontend — a directory-backed
// http.FileServer for the web target, an embedded fs.FS-backed one for the
// desktop target.
func Routes(h *Handler, static http.Handler) *http.ServeMux {
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
	mux.Handle("/", cacheControl(static))
	return mux
}

// cacheControl sets caching headers appropriate for an SPA build:
// content-hashed assets under /assets/ are cached immutably, while the
// unhashed index.html entry point must be revalidated on every load so a
// stale copy never references deleted hashed filenames (blank page on reload).
func cacheControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			// index.html and any other SPA entry response.
			w.Header().Set("Cache-Control", "no-cache")
		}
		next.ServeHTTP(w, r)
	})
}
