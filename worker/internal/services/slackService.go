package services

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/Aniketyadav44/dscheduler/worker/internal/models"
)

func processSlackJob(job *models.Job) (string, error) {
	url, ok := job.Payload["url"].(string)
	if !ok {
		return "", fmt.Errorf("invalid url: %s", job.Payload["url"])
	}

	msg, ok := job.Payload["mg"].(string)
	if !ok {
		return "", fmt.Errorf("invalid msg: %s", job.Payload["msg"])
	}

	body := fmt.Sprintf("{\"text\": \"%s\"}", msg)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Post(url, "application/json", bytes.NewBuffer([]byte(body)))
	if err != nil {
		return "", fmt.Errorf("error sending message: %s", err.Error())
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("non 200/201 status code: %d", res.StatusCode)
	}
	return "Message sent successfully", nil
}
