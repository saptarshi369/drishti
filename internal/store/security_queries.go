package store

// NOTE: package store already has its doc comment in store.go — do not repeat it here.

import (
	"github.com/saptarshi369/drishti/internal/model"
)

// ReplaceSecurityFindings atomically swaps all findings for one (agent,
// projectRoot): delete the prior set, insert the fresh set. This mirrors the
// ReplaceInventory pattern so the table is a faithful snapshot of the latest
// scan (resolved issues disappear). Findings never contain secret values —
// the rule engine already stripped them before producing model.Finding.
func (s *Store) ReplaceSecurityFindings(agent, projectRoot string, findings []model.Finding) error {
	// Serialise all writes through the single write mutex. WAL lets concurrent
	// readers proceed unblocked while we hold this lock.
	s.wmu.Lock()
	defer s.wmu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	// defer Rollback is a Go idiom for automatic cleanup: if Commit succeeds,
	// SQLite treats the subsequent Rollback as a no-op; if we return early on
	// error the transaction is rolled back automatically.
	defer func() { _ = tx.Rollback() }()

	aid := agentID(agent) // resolves "claude" → 1; 0 for unknown agents

	// Full-replace: delete the entire prior set for this (agent, project_root)
	// pair, then insert the fresh findings as a single atomic batch.
	if _, err := tx.Exec(
		`DELETE FROM security_findings WHERE agent_id=? AND project_root=?`,
		aid, projectRoot); err != nil {
		return err
	}

	for _, f := range findings {
		if _, err := tx.Exec(`
			INSERT INTO security_findings
			  (agent_id, project_root, rule_id, severity, title, target_key, detail, remediation, scope)
			VALUES (?,?,?,?,?,?,?,?,?)`,
			aid, projectRoot, f.RuleID, f.Severity, f.Title, f.TargetKey, f.Detail, f.Remediation, f.Scope); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// ListFindings returns the stored findings for one (agent, projectRoot), ordered
// for stable display: severity high→medium→low (using a CASE expression because
// SQLite sorts TEXT lexicographically, not by severity rank), then rule_id, then
// target_key. Returns an empty slice (not nil) when no findings are stored.
func (s *Store) ListFindings(agent, projectRoot string) ([]model.Finding, error) {
	aid := agentID(agent)

	// CASE severity maps strings to integers so ORDER BY yields high→medium→low.
	// Without this, alphabetic order would give high→low→medium (h < l < m).
	rows, err := s.db.Query(`
		SELECT rule_id, severity, title, target_key, detail, remediation, scope
		FROM security_findings
		WHERE agent_id=? AND project_root=?
		ORDER BY CASE severity WHEN 'high' THEN 0 WHEN 'medium' THEN 1 ELSE 2 END,
		         rule_id, target_key`,
		aid, projectRoot)
	if err != nil {
		return nil, err
	}
	// Always close rows — this releases the read lock on the result set and
	// prevents connection leaks even when rows.Next() returns false immediately.
	defer func() { _ = rows.Close() }()

	var out []model.Finding
	for rows.Next() {
		var f model.Finding
		if err := rows.Scan(
			&f.RuleID, &f.Severity, &f.Title, &f.TargetKey,
			&f.Detail, &f.Remediation, &f.Scope); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	// rows.Err() surfaces any iteration error that rows.Next() swallowed.
	return out, rows.Err()
}
