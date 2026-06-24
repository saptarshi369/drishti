-- Inventory: raw discovered items + materialized resolution. Forward-only;
-- the DB is a rebuildable cache, so IF NOT EXISTS keeps re-runs clean.
CREATE TABLE IF NOT EXISTS inventory_items (
  id INTEGER PRIMARY KEY,
  agent_id INTEGER NOT NULL REFERENCES agents(id),
  project_root TEXT NOT NULL DEFAULT '',
  category TEXT NOT NULL,
  name TEXT NOT NULL,
  scope TEXT NOT NULL,
  rel_path TEXT,
  enabled INTEGER NOT NULL DEFAULT 1,
  attrs TEXT,                       -- JSON object
  first_seen_ms INTEGER, last_seen_ms INTEGER,
  UNIQUE(agent_id, project_root, category, scope, name)
);
CREATE INDEX IF NOT EXISTS ix_inv_lookup ON inventory_items(agent_id, project_root, category);

CREATE TABLE IF NOT EXISTS inventory_resolved (
  id INTEGER PRIMARY KEY,
  agent_id INTEGER NOT NULL REFERENCES agents(id),
  project_root TEXT NOT NULL DEFAULT '',
  category TEXT NOT NULL,
  name TEXT NOT NULL,
  effective_status TEXT NOT NULL,
  winner_item_id INTEGER REFERENCES inventory_items(id),
  precedence_trail TEXT,            -- JSON [{step,scope,decision,reason}]
  est_context_tokens INTEGER DEFAULT 0,
  resolved_ms INTEGER,
  UNIQUE(agent_id, project_root, category, name)
);

-- UI contract view: resolved rows joined to codes + the winner's scope/path/attrs,
-- and whether the same name exists in user / project scope (for the ladder columns).
DROP VIEW IF EXISTS v_active_inventory;
CREATE VIEW v_active_inventory AS
SELECT r.id, a.code AS agent, r.project_root, r.category, r.name,
       r.effective_status, r.est_context_tokens, r.precedence_trail,
       w.scope AS winner_scope, w.rel_path AS winner_path, w.attrs AS winner_attrs,
       EXISTS(SELECT 1 FROM inventory_items u WHERE u.agent_id=r.agent_id AND u.project_root=r.project_root
              AND u.category=r.category AND u.name=r.name AND u.scope='user') AS in_user,
       EXISTS(SELECT 1 FROM inventory_items p WHERE p.agent_id=r.agent_id AND p.project_root=r.project_root
              AND p.category=r.category AND p.name=r.name AND p.scope='project') AS in_project
FROM inventory_resolved r
JOIN agents a ON a.id = r.agent_id
LEFT JOIN inventory_items w ON w.id = r.winner_item_id;
