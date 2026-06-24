-- Module 2 (Live Activity): extend events to tool_use/skill/blocked/error and
-- add the cumulative skill_stats counters (survive event rotation, spec §10).

INSERT OR IGNORE INTO event_types (id, code) VALUES
  (4,'tool_use'),(5,'skill'),(6,'blocked'),(7,'error');

ALTER TABLE events ADD COLUMN tool_name TEXT;
ALTER TABLE events ADD COLUMN skill_name TEXT;
ALTER TABLE events ADD COLUMN status TEXT;

CREATE INDEX IF NOT EXISTS ix_events_type_ts ON events(agent_id, type_id, ts_ms);
CREATE INDEX IF NOT EXISTS ix_events_skill   ON events(skill_name) WHERE skill_name IS NOT NULL;

-- Cumulative per-skill trigger counts. project_root is '' in M2 (events carry no
-- project root yet). Folded inside ApplyIngest so counts outlive event rotation.
CREATE TABLE IF NOT EXISTS skill_stats (
  agent_id            INTEGER NOT NULL REFERENCES agents(id),
  project_root        TEXT NOT NULL DEFAULT '',
  skill_name          TEXT NOT NULL,
  trigger_count_total INTEGER NOT NULL DEFAULT 0,
  first_fired_ms      INTEGER,
  last_fired_ms       INTEGER,
  PRIMARY KEY(agent_id, project_root, skill_name)
);
