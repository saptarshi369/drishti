-- Module: per-folder Overview filtering.
-- Events gain a project key so the live-activity feed/sparkline/counters can be
-- scoped to one project, the same way usage_rollup.project_root already scopes
-- spend. The value is Claude's encoded project dir (e.g. -Users-me-dev-app),
-- written at ingest from the transcript's parent directory. Existing rows default
-- to '' (no key) and therefore only appear under the "All" scope — acceptable
-- because the activity windows are rolling/recent and re-key themselves as new
-- events arrive. '' on the read side means "no filter / All".
ALTER TABLE events ADD COLUMN project_root TEXT NOT NULL DEFAULT '';

-- Composite index for the scoped, newest-first reads (WHERE project_root=? ORDER BY ts).
CREATE INDEX IF NOT EXISTS ix_events_project ON events(project_root, ts_ms);
