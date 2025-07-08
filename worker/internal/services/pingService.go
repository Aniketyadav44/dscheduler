package services

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Aniketyadav44/dscheduler/worker/internal/models"
)

func processPingJob(job *models.Job) (string, error) {
	url, ok := job.Payload["url"].(string)
	if !ok {
		return "", fmt.Errorf("invalid url: %s", job.Payload["url"])
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("url ping resulted %d code", resp.StatusCode)
	}
	return "URL pinged successfully", nil
}
