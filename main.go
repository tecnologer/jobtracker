package main

import (
	"crypto/subtle"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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

	authEmail := os.Getenv("AUTH_EMAIL")
	authPassword := os.Getenv("AUTH_PASSWORD")
	if authEmail == "" || authPassword == "" {
		log.Fatal("AUTH_EMAIL and AUTH_PASSWORD must both be set; refusing to start without credentials")
	}

	mux := apiRoutes(handler.New(s))

	// /healthz is registered on the root mux, outside the auth middleware, so
	// Railway's healthcheck can reach it without credentials.
	root := http.NewServeMux()
	root.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	root.Handle("/", basicAuth(mux, authEmail, authPassword))

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           root,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Println("listening on :8080")
	log.Fatal(srv.ListenAndServe())
}

// apiRoutes registers every API handler plus the static SPA fallback.
func apiRoutes(h *handler.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/jobs", h.List)
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
	mux.Handle("/", cacheControl(http.FileServer(http.Dir("web/dist"))))
	return mux
}

// basicAuth gates every request behind HTTP Basic Auth, comparing the submitted
// credentials against the configured email/password. Both fields are compared
// with constant-time equality to avoid leaking their contents via timing, and a
// failure never reveals which of the two was wrong.
func basicAuth(next http.Handler, email, password string) http.Handler {
	expectedEmail := []byte(email)
	expectedPassword := []byte(password)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		emailMatch := subtle.ConstantTimeCompare([]byte(user), expectedEmail) == 1
		passwordMatch := subtle.ConstantTimeCompare([]byte(pass), expectedPassword) == 1
		if !ok || !emailMatch || !passwordMatch {
			w.Header().Set("WWW-Authenticate", `Basic realm="jobtracker"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
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
