package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/tecnologer/jobtracker/handler"
	"github.com/tecnologer/jobtracker/store"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "jobs.db"
	}
	s, err := store.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	h := handler.New(s)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/jobs", h.List)
	mux.HandleFunc("POST /api/jobs", h.Create)
	mux.HandleFunc("PUT /api/jobs/{id}", h.Update)
	mux.HandleFunc("DELETE /api/jobs/{id}", h.Delete)
	mux.HandleFunc("GET /api/jobs/{id}/logs", h.ListStageLogs)
	mux.HandleFunc("POST /api/jobs/{id}/logs", h.AddStageLog)
	mux.HandleFunc("GET /api/jobs/{id}/contacts", h.ListContacts)
	mux.HandleFunc("POST /api/jobs/{id}/contacts", h.CreateContact)
	mux.HandleFunc("DELETE /api/jobs/{id}/contacts/{cid}", h.DeleteContact)
	mux.HandleFunc("GET /api/stages", h.ListDefaultStages)
	mux.HandleFunc("POST /api/stages", h.CreateDefaultStage)
	mux.HandleFunc("GET /api/jobs/{id}/stages", h.ListStages)
	mux.HandleFunc("POST /api/jobs/{id}/stages", h.CreateStage)
	mux.HandleFunc("PUT /api/stages/{id}", h.UpdateStage)
	mux.HandleFunc("DELETE /api/stages/{id}", h.DeleteStage)
	mux.Handle("/", cacheControl(http.FileServer(http.Dir("web/dist"))))

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
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
