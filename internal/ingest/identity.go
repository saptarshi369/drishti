package ingest

import "github.com/saptarshi369/drishti/internal/platform"

// fileIdentity is the default identity function (overridable in tests).
func fileIdentity(path string) (string, error) { return platform.FileIdentity(path) }
