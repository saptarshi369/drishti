-- Module 5 (Security & Audit). Forward-only, idempotent.
-- security_findings: computed at each inventory refresh by the rule engine and
-- full-replaced per (agent, project_root). Holds NO secret values — detail names
-- the offending key only.
CREATE TABLE IF NOT EXISTS security_findings (
  id           INTEGER PRIMARY KEY,
  agent_id     INTEGER NOT NULL REFERENCES agents(id),
  project_root TEXT NOT NULL,
  rule_id      TEXT NOT NULL,
  severity     TEXT NOT NULL,            -- high | medium | low
  title        TEXT NOT NULL,
  target_key   TEXT NOT NULL,            -- scope+relpath, or scope/"global"
  detail       TEXT NOT NULL,            -- names the offending key — never a value
  remediation  TEXT NOT NULL,
  scope        TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS ix_security_findings_lookup
  ON security_findings(agent_id, project_root);
