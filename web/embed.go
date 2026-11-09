// Package web exposes the built front end to the Go server.
package web

import "embed"

// Dist holds the Vite build output. It only contains .gitkeep before the first build.
//
//go:embed all:dist
var Dist embed.FS
