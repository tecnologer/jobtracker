package main

import (
	"crypto/subtle"
	"log"
	"net/http"
	"os"
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

	mux := handler.Routes(handler.New(s), http.FileServer(http.Dir("web/dist")))

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
