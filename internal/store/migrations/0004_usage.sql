-- Module 3 (Usage & Cost). Forward-only, idempotent.
-- quota_samples: live plan-quota forwarded by the statusline helper. EPHEMERAL.
CREATE TABLE IF NOT EXISTS quota_samples (
  id INTEGER PRIMARY KEY,
  agent_id INTEGER NOT NULL REFERENCES agents(id),
  ts_ms INTEGER NOT NULL,
  window TEXT NOT NULL,                 -- five_hour | seven_day
  used_percentage REAL,
  resets_at_ms INTEGER,
  plan TEXT, source TEXT                -- statusline | oauth_fallback
);
CREATE INDEX IF NOT EXISTS ix_quota_ts ON quota_samples(agent_id, window, ts_ms);

-- Latest sample per (agent, window). Powers the gauges + /api/quota.
DROP VIEW IF EXISTS v_latest_quota;
CREATE VIEW v_latest_quota AS
SELECT q.agent_id, a.code AS agent, q.window, q.used_percentage, q.resets_at_ms,
       q.plan, q.source, q.ts_ms
FROM quota_samples q
JOIN agents a ON a.id = q.agent_id
JOIN (SELECT agent_id, window, MAX(ts_ms) AS mx
      FROM quota_samples GROUP BY agent_id, window) m
  ON q.agent_id = m.agent_id AND q.window = m.window AND q.ts_ms = m.mx;

-- Per-day totals across projects/models, joined to the agent code.
-- Powers trend + heatmap + streak.
DROP VIEW IF EXISTS v_usage_daily;
CREATE VIEW v_usage_daily AS
SELECT a.code AS agent, u.day,
  SUM(u.input_tokens)  AS input_tokens,
  SUM(u.output_tokens) AS output_tokens,
  SUM(u.cache_tokens)  AS cache_tokens,
  SUM(u.total_tokens)  AS total_tokens,
  SUM(u.est_cost_usd)  AS est_cost_usd,
  SUM(u.prompt_count)  AS prompt_count
FROM usage_rollup u JOIN agents a ON a.id = u.agent_id
GROUP BY a.id, u.day;
