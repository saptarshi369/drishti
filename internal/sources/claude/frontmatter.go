package claude

import (
	"bytes"
	"errors"

	"gopkg.in/yaml.v3"
)

// errNoFrontmatter signals a file with no leading --- block.
var errNoFrontmatter = errors.New("no frontmatter block")

// parseFrontmatter extracts the leading YAML block delimited by --- lines and
// unmarshals it into out. Content after the closing --- (the Markdown body) is
// ignored. Tolerant of CRLF. Returns errNoFrontmatter if the file does not open
// with a --- line, so callers can fall back to filename-derived defaults.
func parseFrontmatter(b []byte, out any) error {
	b = bytes.TrimPrefix(b, []byte("\xef\xbb\xbf"))       // strip a UTF-8 BOM if present (exactly once)
	b = bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")) // normalize CRLF so fence detection is line-ending agnostic
	if !bytes.HasPrefix(b, []byte("---\n")) {
		return errNoFrontmatter
	}
	// Drop the opening fence line, then find the closing fence.
	rest := b[4:] // skip the opening "---\n" fence (guaranteed present by the HasPrefix guard above)
	end := bytes.Index(rest, []byte("\n---"))
	if end < 0 {
		return errNoFrontmatter
	}
	return yaml.Unmarshal(rest[:end], out)
}
