// Package web holds the built SvelteKit static assets, baked into the binary at
// compile time so end users never need Node.
//
// The embed directive lives here, not in internal/api, because //go:embed
// patterns may not use ".." to escape their package directory — the embedded
// files must live at or below this file. The real bundle lands in web/build via
// `make build-ui`; a placeholder lives there meanwhile. The "all:" prefix keeps
// files SvelteKit emits with leading underscores (e.g. _app/).
package web

import "embed"

// FS is the embedded web/build directory. Consumers use fs.Sub(web.FS, "build").
//
//go:embed all:build
var FS embed.FS
