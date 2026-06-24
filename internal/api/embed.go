// Package api serves the Drishti HTTP + SSE surface and the embedded web UI.
//
// It is the only package that talks HTTP. It reads from the store (and the
// services that assemble views) and never reaches "up" into ingest or app. The
// embedded UI assets are provided by the drishti/web package (the embed must
// live there, since //go:embed cannot reference parent directories).
package api
