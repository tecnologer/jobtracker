// Package web embeds the built frontend for the desktop target. It lives at
// the module root of web/ because go:embed cannot reach outside its own
// package directory (i.e. a file under handler/ or cmd/ could not embed
// web/dist directly).
package web

import "embed"

//go:embed all:dist
var Dist embed.FS
