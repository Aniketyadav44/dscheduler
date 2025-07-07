package services

import (
	"database/sql"
	"encoding/json"

	"github.com/Aniketyadav44/dscheduler/dashboard/internal/models"
)

type WebService struct {
	db *sql.DB
}

func NewWebServcie(db *sql.DB) *WebService {
	return &WebService{
		db: db,
	}
}

// Returns stats of all jobs, completed jobs and failed jobs
func (s *WebService) GetStats() (int, int, int, error) {
	return 0, 0, 0, nil
}

// Returns all jobs. Requires page number and limit
func (s *WebService) GetAllJobs(page, limit int) ([]*models.Job, error) {
	query := `
		SELECT id, hour, minute, type, payload, retries, created_at, updated_at
		FROM jobs
		ORDER BY created_at DESC
		OFFSET $1
		LIMIT $2;
	`

	rows, err := s.db.Query(query, (page-1)*limit, limit)
	if err != nil {
		return nil, err
	}

	jobs := make([]*models.Job, 0)
	for rows.Next() {
		var job models.Job
		var jsonPayload string
		if err := rows.Scan(&job.Id, &job.Hour, &job.Minute, &job.Type, &jsonPayload, &job.Retries, &job.CreatedAt, &job.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(jsonPayload), &job.Payload)
		jobs = append(jobs, &job)
	}
	return jobs, nil
}

// Get all job runs of a job. Returns job id, page number and limit
func (s *WebService) GetJobRuns(jobId, page, limit int, status string) ([]*models.JobRun, error) {
	query := `
		SELECT id, job_id, status, output, error, scheduled_at, completed_at
		FROM job_runs
		WHERE job_id = $1
	`

	args := []any{jobId, (page - 1) * limit, limit}
	if status != "" {
		query += " AND status = $4"
		args = append(args, status)
	}

	query += `
		ORDER BY scheduled_at DESC
		OFFSET $2 LIMIT $3
	`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	jobRuns := make([]*models.JobRun, 0)
	for rows.Next() {
		var jobRun models.JobRun
		if err := rows.Scan(&jobRun.Id, &jobRun.JobId, &jobRun.Status, &jobRun.Output, &jobRun.Error, &jobRun.ScheduledAt, &jobRun.CompletedAt); err != nil {
			return nil, err
		}
		jobRuns = append(jobRuns, &jobRun)
	}
	return jobRuns, nil
}
