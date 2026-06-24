-- Lookup / dimension tables (seeded below). Integer FKs on hot tables.
CREATE TABLE IF NOT EXISTS agents (
  id INTEGER PRIMARY KEY, code TEXT NOT NULL UNIQUE, display_name TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS event_types (
  id INTEGER PRIMARY KEY, code TEXT NOT NULL UNIQUE
);
CREATE TABLE IF NOT EXISTS sources (
  id INTEGER PRIMARY KEY, code TEXT NOT NULL UNIQUE
);

-- Ingestion ledger (never purged).
CREATE TABLE IF NOT EXISTS source_files (
  id INTEGER PRIMARY KEY,
  agent_id INTEGER REFERENCES agents(id),
  kind TEXT NOT NULL,
  abs_path TEXT NOT NULL UNIQUE,
  file_id TEXT, size INTEGER, mtime_ms INTEGER,
  head_hash TEXT,
  last_offset INTEGER NOT NULL DEFAULT 0,
  last_line INTEGER NOT NULL DEFAULT 0,
  state TEXT NOT NULL DEFAULT 'active',
  first_seen_ms INTEGER, last_read_ms INTEGER,
  error_count INTEGER NOT NULL DEFAULT 0, last_error TEXT
);
CREATE INDEX IF NOT EXISTS ix_source_state ON source_files(state);

CREATE TABLE IF NOT EXISTS sessions (
  id INTEGER PRIMARY KEY,
  agent_id INTEGER NOT NULL REFERENCES agents(id),
  session_id TEXT NOT NULL,
  model TEXT, started_ms INTEGER, ended_ms INTEGER,
  prompt_count INTEGER DEFAULT 0,
  input_tokens INTEGER DEFAULT 0, output_tokens INTEGER DEFAULT 0,
  cache_tokens INTEGER DEFAULT 0, est_cost_usd REAL DEFAULT 0,
  UNIQUE(agent_id, session_id)
);

CREATE TABLE IF NOT EXISTS events (
  id INTEGER PRIMARY KEY,
  agent_id INTEGER NOT NULL REFERENCES agents(id),
  type_id INTEGER NOT NULL REFERENCES event_types(id),
  source_id INTEGER NOT NULL REFERENCES sources(id),
  session_id TEXT,
  ts_ms INTEGER NOT NULL,
  dedupe_key TEXT NOT NULL UNIQUE
);
CREATE INDEX IF NOT EXISTS ix_events_ts ON events(ts_ms);

CREATE TABLE IF NOT EXISTS usage_rollup (
  id INTEGER PRIMARY KEY,
  agent_id INTEGER NOT NULL REFERENCES agents(id),
  day INTEGER NOT NULL, project_root TEXT DEFAULT '', model TEXT DEFAULT '',
  input_tokens INTEGER DEFAULT 0, output_tokens INTEGER DEFAULT 0,
  cache_tokens INTEGER DEFAULT 0, total_tokens INTEGER DEFAULT 0,
  est_cost_usd REAL DEFAULT 0, session_count INTEGER DEFAULT 0, prompt_count INTEGER DEFAULT 0,
  UNIQUE(agent_id, day, project_root, model)
);

CREATE TABLE IF NOT EXISTS app_meta (key TEXT PRIMARY KEY, value TEXT);

-- Seeds (idempotent).
INSERT OR IGNORE INTO agents (id, code, display_name) VALUES (1,'claude','Claude Code');
INSERT OR IGNORE INTO event_types (id, code) VALUES (1,'prompt'),(2,'session_start'),(3,'session_end');
INSERT OR IGNORE INTO sources (id, code) VALUES (1,'transcript');

-- Overview KPI view: today's per-agent rollup, joined to codes (no raw ids leak).
DROP VIEW IF EXISTS v_overview_kpis;
CREATE VIEW v_overview_kpis AS
SELECT a.code AS agent,
  COALESCE(SUM(u.prompt_count),0)  AS prompts_today,
  COALESCE(SUM(u.est_cost_usd),0)  AS spend_today_usd,
  COALESCE(SUM(u.input_tokens),0)  AS input_tokens,
  COALESCE(SUM(u.output_tokens),0) AS output_tokens,
  COALESCE(SUM(u.cache_tokens),0)  AS cache_tokens
FROM agents a
LEFT JOIN usage_rollup u
  ON u.agent_id = a.id
  AND u.day = CAST(strftime('%Y%m%d','now','localtime') AS INTEGER)
GROUP BY a.id;
