package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/models"
)

type BackupRepository struct {
	DB *pgxpool.Pool
}

func (r *BackupRepository) GetSettings(ctx context.Context) (*models.BackupSettings, error) {
	row := r.DB.QueryRow(ctx, `
		SELECT id, enabled, cron_expr, retention_count, scope_db, scope_uploads, scope_configs,
		       rclone_remote, remote_folder, gdrive_client_id, gdrive_client_secret, gdrive_account_email,
		       updated_at, updated_by
		FROM backup_settings WHERE id = 'global'
	`)
	return scanSettings(row)
}

func (r *BackupRepository) UpdateSettings(ctx context.Context, in models.BackupSettings, updatedBy *string) (*models.BackupSettings, error) {
	row := r.DB.QueryRow(ctx, `
		UPDATE backup_settings SET
			enabled = $1, cron_expr = $2, retention_count = $3,
			scope_db = $4, scope_uploads = $5, scope_configs = $6,
			rclone_remote = $7, remote_folder = $8,
			updated_at = NOW(), updated_by = $9
		WHERE id = 'global'
		RETURNING id, enabled, cron_expr, retention_count, scope_db, scope_uploads, scope_configs,
		          rclone_remote, remote_folder, gdrive_client_id, gdrive_client_secret, gdrive_account_email,
		          updated_at, updated_by
	`, in.Enabled, in.CronExpr, in.RetentionCount, in.ScopeDB, in.ScopeUploads, in.ScopeConfigs,
		in.RcloneRemote, in.RemoteFolder, updatedBy)
	return scanSettings(row)
}

func (r *BackupRepository) UpdateGdriveCreds(ctx context.Context, clientID, clientSecret string) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE backup_settings SET gdrive_client_id = $1, gdrive_client_secret = $2, updated_at = NOW()
		WHERE id = 'global'
	`, clientID, clientSecret)
	return err
}

func (r *BackupRepository) UpdateGdriveAccount(ctx context.Context, email string) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE backup_settings SET gdrive_account_email = $1, updated_at = NOW() WHERE id = 'global'
	`, email)
	return err
}

func (r *BackupRepository) ClearGdriveAccount(ctx context.Context) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE backup_settings SET gdrive_account_email = '', updated_at = NOW() WHERE id = 'global'
	`)
	return err
}

func scanSettings(row pgx.Row) (*models.BackupSettings, error) {
	var s models.BackupSettings
	if err := row.Scan(&s.ID, &s.Enabled, &s.CronExpr, &s.RetentionCount, &s.ScopeDB, &s.ScopeUploads, &s.ScopeConfigs,
		&s.RcloneRemote, &s.RemoteFolder, &s.GdriveClientID, &s.GdriveClientSecret, &s.GdriveAccountEmail,
		&s.UpdatedAt, &s.UpdatedBy); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *BackupRepository) CreateRun(ctx context.Context, kind string, scopeDB, scopeUploads, scopeConfigs bool, triggeredBy *string) (*models.BackupRun, error) {
	row := r.DB.QueryRow(ctx, `
		INSERT INTO backup_runs (kind, status, scope_db, scope_uploads, scope_configs, triggered_by)
		VALUES ($1, 'pending', $2, $3, $4, $5)
		RETURNING id, kind, status, scope_db, scope_uploads, scope_configs, started_at, finished_at,
		          size_bytes, remote_path, error, log_tail, triggered_by
	`, kind, scopeDB, scopeUploads, scopeConfigs, triggeredBy)
	return scanRun(row)
}

func (r *BackupRepository) UpdateRunStatus(ctx context.Context, id, status, errMsg, remotePath, logTail string, sizeBytes int64, finished bool) error {
	if finished {
		_, err := r.DB.Exec(ctx, `
			UPDATE backup_runs
			SET status = $1, error = $2, remote_path = $3, log_tail = $4, size_bytes = $5, finished_at = NOW()
			WHERE id = $6
		`, status, errMsg, remotePath, logTail, sizeBytes, id)
		return err
	}
	_, err := r.DB.Exec(ctx, `UPDATE backup_runs SET status = $1, log_tail = $2 WHERE id = $3`, status, logTail, id)
	return err
}

func (r *BackupRepository) ListRuns(ctx context.Context, limit int) ([]models.BackupRun, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT id, kind, status, scope_db, scope_uploads, scope_configs, started_at, finished_at,
		       size_bytes, remote_path, error, log_tail, triggered_by
		FROM backup_runs ORDER BY started_at DESC LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []models.BackupRun
	for rows.Next() {
		run, err := scanRunFromRows(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, *run)
	}
	return runs, rows.Err()
}

func (r *BackupRepository) GetRun(ctx context.Context, id string) (*models.BackupRun, error) {
	row := r.DB.QueryRow(ctx, `
		SELECT id, kind, status, scope_db, scope_uploads, scope_configs, started_at, finished_at,
		       size_bytes, remote_path, error, log_tail, triggered_by
		FROM backup_runs WHERE id = $1
	`, id)
	run, err := scanRun(row)
	if err != nil && err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	return run, err
}

func (r *BackupRepository) DeleteRun(ctx context.Context, id string) error {
	ct, err := r.DB.Exec(ctx, `DELETE FROM backup_runs WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *BackupRepository) HasRunningBackup(ctx context.Context) (bool, error) {
	var count int
	err := r.DB.QueryRow(ctx, `SELECT COUNT(*) FROM backup_runs WHERE status IN ('pending', 'running')`).Scan(&count)
	return count > 0, err
}

func scanRun(row pgx.Row) (*models.BackupRun, error) {
	var run models.BackupRun
	if err := row.Scan(&run.ID, &run.Kind, &run.Status, &run.ScopeDB, &run.ScopeUploads, &run.ScopeConfigs,
		&run.StartedAt, &run.FinishedAt, &run.SizeBytes, &run.RemotePath, &run.Error, &run.LogTail, &run.TriggeredBy); err != nil {
		return nil, err
	}
	return &run, nil
}

func scanRunFromRows(rows pgx.Rows) (*models.BackupRun, error) {
	var run models.BackupRun
	if err := rows.Scan(&run.ID, &run.Kind, &run.Status, &run.ScopeDB, &run.ScopeUploads, &run.ScopeConfigs,
		&run.StartedAt, &run.FinishedAt, &run.SizeBytes, &run.RemotePath, &run.Error, &run.LogTail, &run.TriggeredBy); err != nil {
		return nil, err
	}
	return &run, nil
}
