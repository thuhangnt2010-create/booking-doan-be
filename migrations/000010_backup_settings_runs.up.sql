CREATE TABLE backup_settings (
    id                   TEXT PRIMARY KEY DEFAULT 'global' CHECK (id = 'global'),
    enabled              BOOLEAN NOT NULL DEFAULT false,
    cron_expr            VARCHAR(50) NOT NULL DEFAULT '0 2 * * *',
    retention_count      INT NOT NULL DEFAULT 14 CHECK (retention_count >= 1),
    scope_db             BOOLEAN NOT NULL DEFAULT true,
    scope_uploads        BOOLEAN NOT NULL DEFAULT false,
    scope_configs        BOOLEAN NOT NULL DEFAULT true,
    rclone_remote        VARCHAR(50) NOT NULL DEFAULT 'gdrive',
    remote_folder        VARCHAR(200) NOT NULL DEFAULT 'booking-doan-backups',
    gdrive_client_id     VARCHAR(200) NOT NULL DEFAULT '',
    gdrive_client_secret VARCHAR(200) NOT NULL DEFAULT '',
    gdrive_account_email VARCHAR(200) NOT NULL DEFAULT '',
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by           UUID
);
INSERT INTO backup_settings (id) VALUES ('global') ON CONFLICT DO NOTHING;

CREATE TABLE backup_runs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kind          VARCHAR(20) NOT NULL CHECK (kind IN ('manual', 'scheduled')),
    status        VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'running', 'success', 'failed')),
    scope_db      BOOLEAN NOT NULL DEFAULT false,
    scope_uploads BOOLEAN NOT NULL DEFAULT false,
    scope_configs BOOLEAN NOT NULL DEFAULT false,
    started_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at   TIMESTAMPTZ,
    size_bytes    BIGINT NOT NULL DEFAULT 0,
    remote_path   VARCHAR(500) NOT NULL DEFAULT '',
    error         TEXT NOT NULL DEFAULT '',
    log_tail      TEXT NOT NULL DEFAULT '',
    triggered_by  UUID
);
CREATE INDEX idx_backup_runs_started_at ON backup_runs(started_at DESC);
CREATE INDEX idx_backup_runs_status ON backup_runs(status) WHERE status IN ('pending', 'running');
