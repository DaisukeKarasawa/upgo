package tracker

import (
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type StatusTracker struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewStatusTracker(db *sql.DB, logger *zap.Logger) *StatusTracker {
	return &StatusTracker{
		db:     db,
		logger: logger,
	}
}

func (t *StatusTracker) TrackPRState(prID int, newState string) (bool, error) {
	var currentState sql.NullString
	var previousState sql.NullString
	err := t.db.QueryRow(
		"SELECT state, previous_state FROM pull_requests WHERE id = ?",
		prID,
	).Scan(&currentState, &previousState)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("PR状態の取得に失敗しました: %w", err)
	}

	// Check if state has changed
	currentStateValue := ""
	if currentState.Valid {
		currentStateValue = currentState.String
	}

	if currentStateValue != newState {
		_, err = t.db.Exec(
			"UPDATE pull_requests SET state = ?, previous_state = ?, updated_at = ? WHERE id = ?",
			newState, currentStateValue, time.Now(), prID,
		)
		if err != nil {
			return false, fmt.Errorf("PR状態の更新に失敗しました: %w", err)
		}

		t.logger.Info("PR状態が変更されました",
			zap.Int("pr_id", prID),
			zap.String("old_state", currentStateValue),
			zap.String("new_state", newState),
		)

		return true, nil
	}

	return false, nil
}

func (t *StatusTracker) UpdatePRLastSynced(prID int) error {
	_, err := t.db.Exec(
		"UPDATE pull_requests SET last_synced_at = ? WHERE id = ?",
		time.Now(), prID,
	)
	return err
}
