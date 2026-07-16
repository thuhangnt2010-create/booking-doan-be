package models

import "time"

type BackupSettings struct {
	ID                 string    `json:"id"`
	Enabled            bool      `json:"enabled"`
	CronExpr           string    `json:"cronExpr"`
	RetentionCount     int       `json:"retentionCount"`
	ScopeDB            bool      `json:"scopeDb"`
	ScopeUploads       bool      `json:"scopeUploads"`
	ScopeConfigs       bool      `json:"scopeConfigs"`
	RcloneRemote       string    `json:"rcloneRemote"`
	RemoteFolder       string    `json:"remoteFolder"`
	GdriveClientID     string    `json:"gdriveClientId"`
	GdriveClientSecret string    `json:"-"` // never serialize the secret back to the browser
	GdriveAccountEmail string    `json:"gdriveAccountEmail"`
	UpdatedAt          time.Time `json:"updatedAt"`
	UpdatedBy          *string   `json:"updatedBy,omitempty"`
}

type BackupRun struct {
	ID           string     `json:"id"`
	Kind         string     `json:"kind"`
	Status       string     `json:"status"`
	ScopeDB      bool       `json:"scopeDb"`
	ScopeUploads bool       `json:"scopeUploads"`
	ScopeConfigs bool       `json:"scopeConfigs"`
	StartedAt    time.Time  `json:"startedAt"`
	FinishedAt   *time.Time `json:"finishedAt,omitempty"`
	SizeBytes    int64      `json:"sizeBytes"`
	RemotePath   string     `json:"remotePath"`
	Error        string     `json:"error"`
	LogTail      string     `json:"logTail"`
	TriggeredBy  *string    `json:"triggeredBy,omitempty"`
}
