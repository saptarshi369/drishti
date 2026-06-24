package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/saptarshi369/drishti/internal/model"
)

// ReplaceInventory atomically swaps all raw items for one (agent, projectRoot):
// delete the prior set, insert the fresh set. This keeps the table a faithful
// mirror of what currently exists on disk (vanished items disappear).
func (s *Store) ReplaceInventory(agent, projectRoot string, items []model.InventoryItem) error {
	s.wmu.Lock()
	defer s.wmu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	now := time.Now().UnixMilli()
	aid := agentID(agent)
	// Delete resolved rows first: they hold a foreign key to inventory_items,
	// so we must remove dependents before removing the referenced rows.
	// (The resolved table is a derived cache; replacing items invalidates it.)
	if _, err := tx.Exec(`DELETE FROM inventory_resolved WHERE agent_id=? AND project_root=?`, aid, projectRoot); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM inventory_items WHERE agent_id=? AND project_root=?`, aid, projectRoot); err != nil {
		return err
	}
	for _, it := range items {
		attrs, err := json.Marshal(it.Attrs)
		if err != nil {
			return fmt.Errorf("marshal attrs: %w", err)
		}
		if _, err := tx.Exec(`
			INSERT INTO inventory_items
			  (agent_id, project_root, category, name, scope, rel_path, enabled, attrs, first_seen_ms, last_seen_ms)
			VALUES (?,?,?,?,?,?,?,?,?,?)`,
			aid, projectRoot, string(it.Category), it.Name, string(it.Scope), it.RelPath,
			boolToInt(it.Enabled), string(attrs), now, now); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ReplaceResolved atomically swaps the materialized rows for one (agent,
// projectRoot). Each winner is mapped back to its inventory_items.id via the
// natural key; the trail is stored as JSON.
func (s *Store) ReplaceResolved(agent, projectRoot string, resolved []model.ResolvedItem) error {
	s.wmu.Lock()
	defer s.wmu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	now := time.Now().UnixMilli()
	aid := agentID(agent)
	if _, err := tx.Exec(`DELETE FROM inventory_resolved WHERE agent_id=? AND project_root=?`, aid, projectRoot); err != nil {
		return err
	}
	for _, r := range resolved {
		trail, err := json.Marshal(r.PrecedenceTrail)
		if err != nil {
			return fmt.Errorf("marshal trail: %w", err)
		}
		var winnerID sql.NullInt64
		if r.Winner != nil {
			var id int64
			lookupErr := tx.QueryRow(`SELECT id FROM inventory_items
				WHERE agent_id=? AND project_root=? AND category=? AND scope=? AND name=?`,
				aid, projectRoot, string(r.Winner.Category), string(r.Winner.Scope), r.Winner.Name).Scan(&id)
			if lookupErr != nil && !errors.Is(lookupErr, sql.ErrNoRows) {
				return fmt.Errorf("lookup winner id: %w", lookupErr)
			}
			if lookupErr == nil {
				winnerID = sql.NullInt64{Int64: id, Valid: true}
			}
		}
		if _, err := tx.Exec(`
			INSERT INTO inventory_resolved
			  (agent_id, project_root, category, name, effective_status, winner_item_id, precedence_trail, est_context_tokens, resolved_ms)
			VALUES (?,?,?,?,?,?,?,?,?)`,
			aid, projectRoot, string(r.Category), r.Name, string(r.EffectiveStatus),
			winnerID, string(trail), r.EstContextTokens, now); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ListResolved reads materialized rows from the v_active_inventory view. An
// empty category means all categories. disabled/shadowed rows are hidden unless
// showDisabled is set (the screen's "show disabled" toggle).
func (s *Store) ListResolved(category, projectRoot string, showDisabled bool) ([]model.ResolvedRow, error) {
	q := `SELECT id, category, name, effective_status,
	             COALESCE(winner_scope,''), COALESCE(winner_path,''),
	             in_user, in_project, est_context_tokens, COALESCE(winner_attrs,'{}')
	      FROM v_active_inventory WHERE project_root=?`
	args := []any{projectRoot}
	if category != "" {
		q += " AND category=?"
		args = append(args, category)
	}
	if !showDisabled {
		q += " AND effective_status NOT IN ('disabled','shadowed')"
	}
	q += " ORDER BY category, name"

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []model.ResolvedRow
	for rows.Next() {
		var r model.ResolvedRow
		var attrs string
		if err := rows.Scan(&r.ID, &r.Category, &r.Name, &r.EffectiveStatus,
			&r.WinnerScope, &r.WinnerPath, &r.InUser, &r.InProject, &r.EstContextTokens, &attrs); err != nil {
			return nil, err
		}
		if jsonErr := json.Unmarshal([]byte(attrs), &r.Attrs); jsonErr != nil {
			return nil, fmt.Errorf("decode winner_attrs for row %d: %w", r.ID, jsonErr)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ResolvedTrail returns the precedence trail for one resolved row (the "why?"
// drawer). An unknown id yields an empty slice, not an error.
func (s *Store) ResolvedTrail(id int64) ([]model.PrecedenceStep, error) {
	var raw string
	err := s.db.QueryRow(`SELECT COALESCE(precedence_trail,'[]') FROM inventory_resolved WHERE id=?`, id).Scan(&raw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	var steps []model.PrecedenceStep
	if jsonErr := json.Unmarshal([]byte(raw), &steps); jsonErr != nil {
		return nil, fmt.Errorf("decode precedence_trail for id %d: %w", id, jsonErr)
	}
	return steps, nil
}

// boolToInt maps a Go bool to SQLite's 0/1 integer convention.
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
