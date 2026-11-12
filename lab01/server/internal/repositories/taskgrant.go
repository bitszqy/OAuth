package repositories

import (
	"context"
	"database/sql"

	"github.com/cyllective/oauth-labs/lab01/server/internal/dto"
)

type TaskGrantRepository struct {
	db *sql.DB
}

func NewTaskGrantRepository(db *sql.DB) *TaskGrantRepository {
	return &TaskGrantRepository{db: db}
}

func (r *TaskGrantRepository) Upsert(ctx context.Context, grant *dto.TaskGrant) error {
	stmt := `
		INSERT INTO user_task_grants(user_id, client_id, task_id, revoked)
		VALUES(?, ?, ?, false)
		ON DUPLICATE KEY UPDATE revoked = false, updated_at = CURRENT_TIMESTAMP`
	_, err := r.db.ExecContext(ctx, stmt, grant.UserID, grant.ClientID, grant.TaskID)
	return err
}

func (r *TaskGrantRepository) Exists(ctx context.Context, grant *dto.TaskGrant) bool {
	stmt := `
		SELECT 1
		FROM user_task_grants
		WHERE user_id = ? AND client_id = ? AND task_id = ? AND revoked = false`
	var exists int
	return r.db.QueryRowContext(ctx, stmt, grant.UserID, grant.ClientID, grant.TaskID).Scan(&exists) == nil
}

func (r *TaskGrantRepository) RevokeAll(ctx context.Context, clientID, userID string) error {
	stmt := `
		UPDATE user_task_grants
		SET revoked = true, updated_at = CURRENT_TIMESTAMP
		WHERE client_id = ? AND user_id = ?`
	_, err := r.db.ExecContext(ctx, stmt, clientID, userID)
	return err
}
