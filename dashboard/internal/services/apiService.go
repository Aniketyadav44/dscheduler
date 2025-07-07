package services

import (
	"database/sql"
	"encoding/json"

	"github.com/Aniketyadav44/dscheduler/dashboard/internal/models"
)

type APIService struct {
	db *sql.DB
}

func NewAPIService(db *sql.DB) *APIService {
	return &APIService{
		db: db,
	}
}

// To insert a new job into jobs table
func (s *APIService) CreateNewJob(job *models.Job) error {
	updatedQuery := `
			INSERT INTO jobs(hour, minute, type, payload)
			VALUES($1, $2, $3, $4)
			RETURNING id
	`

	payloadJSON, _ := json.Marshal(job.Payload)
	var jobId int
	if err := s.db.QueryRow(updatedQuery, job.Hour, job.Minute, job.Type, payloadJSON).Scan(&jobId); err != nil {
		return err
	}
	job.Id = jobId
	return nil
}

// To delete a job and all of it's run entries from job_runs table
func (s *APIService) DeleteJob(id int) error {
	delQuery := `DELETE FROM jobs WHERE id = $1`

	runsDelQuery := `DELETE FROM job_runs WHERE job_id = $1`

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(delQuery, id); err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		return err
	}

	if _, err := tx.Exec(runsDelQuery, id); err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
