package service

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/thuhangnt2010-create/booking-doan-be/internal/repository"
)

type BackupScheduler struct {
	Repo   *repository.BackupRepository
	Backup *BackupService
}

// Start runs an in-process cron loop honoring backup_settings.enabled/cron_expr,
// polling for settings changes every 60s so a config save takes effect without
// a redeploy. Returns a stop function.
func (s *BackupScheduler) Start(ctx context.Context) func() {
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		loc = time.UTC
	}
	c := cron.New(cron.WithLocation(loc))

	var curExpr string
	var curEnabled bool
	var entryID cron.EntryID

	reload := func() {
		settings, err := s.Repo.GetSettings(ctx)
		if err != nil {
			log.Printf("backup scheduler: get settings failed: %v", err)
			return
		}
		if settings.Enabled == curEnabled && settings.CronExpr == curExpr {
			return
		}
		if entryID != 0 {
			c.Remove(entryID)
			entryID = 0
		}
		if settings.Enabled {
			id, err := c.AddFunc(settings.CronExpr, func() { s.triggerScheduled(ctx) })
			if err != nil {
				log.Printf("backup scheduler: invalid cron expr %q: %v", settings.CronExpr, err)
				return
			}
			entryID = id
		}
		curEnabled, curExpr = settings.Enabled, settings.CronExpr
	}

	reload()
	c.Start()

	stopPolling := make(chan struct{})
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				reload()
			case <-stopPolling:
				return
			}
		}
	}()

	return func() {
		close(stopPolling)
		c.Stop()
	}
}

func (s *BackupScheduler) triggerScheduled(ctx context.Context) {
	busy, err := s.Repo.HasRunningBackup(ctx)
	if err != nil || busy {
		return
	}
	settings, err := s.Repo.GetSettings(ctx)
	if err != nil {
		return
	}
	run, err := s.Repo.CreateRun(ctx, "scheduled", settings.ScopeDB, settings.ScopeUploads, settings.ScopeConfigs, nil)
	if err != nil {
		return
	}
	go s.Backup.RunBackup(ctx, run.ID, settings)
}
