package models

import (
	"database/sql"
	"time"
)

type JobRun struct {
	Id          int            `json:"id"`
	JobId       int            `json:"job_id"`
	Status      string         `json:"status"` //running, completed, failed, permanently failed
	Output      sql.NullString `json:"output"`
	Error       sql.NullString `json:"error"`
	ScheduledAt time.Time      `json:"scheduled_at"`
	CompletedAt time.Time      `json:"completed_at"`
}
