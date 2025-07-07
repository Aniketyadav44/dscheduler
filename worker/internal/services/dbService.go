package services

import (
	"database/sql"

	"github.com/Aniketyadav44/dscheduler/worker/internal/models"
)

type DBService struct {
	db *sql.DB
}

func NewDBService(db *sql.DB) *DBService {
	return &DBService{
		db: db,
	}
}

// to get job's retry count
func (ds *DBService) getJobRetryCount(id, hour int) (int, error) {
	query := `
		SELECT retries FROM jobs
		WHERE hour = $1 AND id = $2
	`

	var retries int
	if err := ds.db.QueryRow(query, hour, id).Scan(&retries); err != nil {
		return 0, err
	}
	return retries, nil
}

// make a job run entry
func (ds *DBService) registerJobEntry(job *models.Job, jobEntry *models.JobRun) error {
	query := `
		INSERT INTO job_runs(job_id, status, output, error, scheduled_at)
		VALUES($1, $2, $3, $4, $5)
	`
	if _, err := ds.db.Exec(query, jobEntry.JobId, jobEntry.Status, jobEntry.Output, jobEntry.Error, jobEntry.ScheduledAt); err != nil {
		return err
	}

	// if a failed job, increment retries count
	if jobEntry.Status == JOB_FAILED {
		updateRetryCountQuery := `
			UPDATE jobs SET retries = retries + 1 WHERE hour = $1 AND id = $2
		`
		if _, err := ds.db.Exec(updateRetryCountQuery, job.Hour, job.Id); err != nil {
			return err
		}
	}

	// if a permanently failed or completed job, resetting retries count to 0
	if jobEntry.Status == JOB_PERMANENTLY_FAILED || jobEntry.Status == JOB_COMPLETED {
		resetRetriesQuery := `
			UPDATE jobs SET retries = 0 WHERE hour = $1 AND id = $2
		`
		if _, err := ds.db.Exec(resetRetriesQuery, job.Hour, job.Id); err != nil {
			return err
		}
	}

	return nil
}
