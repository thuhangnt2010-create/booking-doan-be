package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/thuhangnt2010-create/booking-doan-be/internal/models"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/repository"
)

const backupStagingRoot = "/tmp/backups"

// BackupService orchestrates a single backup run: dump Postgres, mirror MinIO
// uploads (if any buckets are configured), tar the infra repo snapshot, then
// push everything to the configured rclone remote (Google Drive).
type BackupService struct {
	Repo         *repository.BackupRepository
	DatabaseURL  string
	MinioAddr    string
	MinioUser    string
	MinioPass    string
	MinioBuckets []string
}

func (s *BackupService) RunBackup(ctx context.Context, runID string, settings *models.BackupSettings) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Minute)
	defer cancel()

	var logBuf bytes.Buffer
	logf := func(format string, args ...any) { fmt.Fprintf(&logBuf, format+"\n", args...) }

	s.Repo.UpdateRunStatus(ctx, runID, "running", "", "", tail(&logBuf), 0, false)

	timestamp := time.Now().Format("20060102-150405")
	stageDir := filepath.Join(backupStagingRoot, timestamp)
	if err := os.MkdirAll(stageDir, 0o755); err != nil {
		s.fail(ctx, runID, &logBuf, err)
		return
	}
	defer os.RemoveAll(stageDir)

	if settings.ScopeDB {
		if err := s.dumpDatabase(ctx, stageDir, logf); err != nil {
			s.fail(ctx, runID, &logBuf, err)
			return
		}
	}
	if settings.ScopeUploads && len(s.MinioBuckets) > 0 {
		if err := s.mirrorUploads(ctx, stageDir, logf); err != nil {
			s.fail(ctx, runID, &logBuf, err)
			return
		}
	}
	if settings.ScopeConfigs {
		if err := s.archiveConfigs(ctx, stageDir, logf); err != nil {
			s.fail(ctx, runID, &logBuf, err)
			return
		}
	}

	remotePath := fmt.Sprintf("%s:%s/%s", settings.RcloneRemote, settings.RemoteFolder, timestamp)
	if err := s.rcloneCopy(ctx, stageDir, remotePath, logf); err != nil {
		s.fail(ctx, runID, &logBuf, err)
		return
	}

	size, err := s.rcloneSize(ctx, remotePath)
	if err != nil {
		logf("rclone size (non-fatal): %v", err)
	}

	if settings.RetentionCount > 0 {
		if err := s.applyRetention(ctx, settings, logf); err != nil {
			logf("retention warning: %v", err)
		}
	}

	s.Repo.UpdateRunStatus(ctx, runID, "success", "", remotePath, tail(&logBuf), size, true)
}

func (s *BackupService) fail(ctx context.Context, runID string, logBuf *bytes.Buffer, err error) {
	fmt.Fprintf(logBuf, "ERROR: %v\n", err)
	s.Repo.UpdateRunStatus(ctx, runID, "failed", err.Error(), "", tail(logBuf), 0, true)
}

func tail(buf *bytes.Buffer) string {
	s := buf.String()
	if len(s) > 4096 {
		return s[len(s)-4096:]
	}
	return s
}

func (s *BackupService) dumpDatabase(ctx context.Context, stageDir string, logf func(string, ...any)) error {
	out := filepath.Join(stageDir, "db.sql.gz")
	cmdStr := fmt.Sprintf("pg_dump --clean --if-exists '%s' | gzip > '%s'", s.DatabaseURL, out)
	cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)
	var combined bytes.Buffer
	cmd.Stdout, cmd.Stderr = &combined, &combined
	if err := cmd.Run(); err != nil {
		logf("pg_dump failed: %s", combined.String())
		return fmt.Errorf("pg_dump failed: %w", err)
	}
	logf("pg_dump ok")
	return nil
}

func (s *BackupService) mirrorUploads(ctx context.Context, stageDir string, logf func(string, ...any)) error {
	uploadsDir := filepath.Join(stageDir, "uploads")
	if err := os.MkdirAll(uploadsDir, 0o755); err != nil {
		return err
	}
	env := append(os.Environ(), fmt.Sprintf("MC_HOST_local=http://%s:%s@%s", s.MinioUser, s.MinioPass, s.MinioAddr))
	for _, bucket := range s.MinioBuckets {
		cmd := exec.CommandContext(ctx, "mc", "mirror", "local/"+bucket, filepath.Join(uploadsDir, bucket))
		cmd.Env = env
		var combined bytes.Buffer
		cmd.Stdout, cmd.Stderr = &combined, &combined
		if err := cmd.Run(); err != nil {
			logf("mc mirror %s failed: %s", bucket, combined.String())
			return fmt.Errorf("mc mirror %s failed: %w", bucket, err)
		}
	}
	tarPath := filepath.Join(stageDir, "uploads.tar.gz")
	cmd := exec.CommandContext(ctx, "tar", "-czf", tarPath, "-C", stageDir, "uploads")
	var combined bytes.Buffer
	cmd.Stdout, cmd.Stderr = &combined, &combined
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tar uploads failed: %w: %s", err, combined.String())
	}
	os.RemoveAll(uploadsDir)
	logf("uploads mirrored: %d bucket(s)", len(s.MinioBuckets))
	return nil
}

func (s *BackupService) archiveConfigs(ctx context.Context, stageDir string, logf func(string, ...any)) error {
	if _, err := os.Stat("/infra-snapshot"); err != nil {
		logf("configs scope: /infra-snapshot not mounted, skip")
		return nil
	}
	tarPath := filepath.Join(stageDir, "configs.tar.gz")
	cmd := exec.CommandContext(ctx, "tar", "-czf", tarPath, "--exclude=.git", "--exclude=node_modules", "-C", "/", "infra-snapshot")
	var combined bytes.Buffer
	cmd.Stdout, cmd.Stderr = &combined, &combined
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tar configs failed: %w: %s", err, combined.String())
	}
	logf("configs archived")
	return nil
}

func (s *BackupService) rcloneCopy(ctx context.Context, stageDir, remotePath string, logf func(string, ...any)) error {
	cmd := exec.CommandContext(ctx, "rclone", "copy", stageDir, remotePath)
	var combined bytes.Buffer
	cmd.Stdout, cmd.Stderr = &combined, &combined
	if err := cmd.Run(); err != nil {
		logf("rclone copy failed: %s", combined.String())
		return fmt.Errorf("rclone copy failed: %w", err)
	}
	logf("rclone copy ok -> %s", remotePath)
	return nil
}

func (s *BackupService) rcloneSize(ctx context.Context, remotePath string) (int64, error) {
	cmd := exec.CommandContext(ctx, "rclone", "size", "--json", remotePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return 0, err
	}
	var parsed struct {
		Bytes int64 `json:"bytes"`
	}
	if err := json.Unmarshal(out.Bytes(), &parsed); err != nil {
		return 0, err
	}
	return parsed.Bytes, nil
}

func (s *BackupService) applyRetention(ctx context.Context, settings *models.BackupSettings, logf func(string, ...any)) error {
	remoteRoot := fmt.Sprintf("%s:%s", settings.RcloneRemote, settings.RemoteFolder)
	cmd := exec.CommandContext(ctx, "rclone", "lsf", remoteRoot, "--dirs-only")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rclone lsf failed: %w", err)
	}

	var folders []string
	for _, line := range strings.Split(out.String(), "\n") {
		f := strings.TrimSuffix(strings.TrimSpace(line), "/")
		if f != "" {
			folders = append(folders, f)
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(folders))) // timestamp folder names sort lexicographically = time order
	if len(folders) <= settings.RetentionCount {
		return nil
	}

	for _, old := range folders[settings.RetentionCount:] {
		purgeCmd := exec.CommandContext(ctx, "rclone", "purge", fmt.Sprintf("%s/%s", remoteRoot, old))
		var combined bytes.Buffer
		purgeCmd.Stdout, purgeCmd.Stderr = &combined, &combined
		if err := purgeCmd.Run(); err != nil {
			logf("purge %s failed: %s", old, combined.String())
			continue
		}
		logf("purged old backup: %s", old)
	}
	return nil
}
