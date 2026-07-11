package handler

import "net/http"

// buildVersion is stamped at build time via
// -ldflags "-X github.com/tecnologer/jobtracker/handler.buildVersion=v1.2.3"
// (see Makefile and .github/workflows/release.yml). It lives here rather than
// in a main package so the web and desktop binaries share one flag.
var buildVersion = "dev" //nolint:gochecknoglobals // ldflags injection target; must be a package-level var

func (h *Handler) Version(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"version": buildVersion})
}
