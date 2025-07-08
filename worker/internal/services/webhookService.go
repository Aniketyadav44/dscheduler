package services

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Aniketyadav44/dscheduler/worker/internal/models"
)

func processWebhookJob(job *models.Job) (string, error) {
	url, ok := job.Payload["url"].(string)
	if !ok {
		return "", fmt.Errorf("invalid webhook url: %s", job.Payload["url"])
	}

	body, ok := job.Payload["body"].(string)
	if !ok {
		return "", fmt.Errorf("invalid body: %s", job.Payload["body"])
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Post(url, "application/json", bytes.NewBuffer([]byte(body)))
	if err != nil {
		return "", fmt.Errorf("error calling POST: %s", err.Error())
	}
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("non 200/201 status code in post api: %d", res.StatusCode)
	} else {
		defer res.Body.Close()
		resBody, _ := io.ReadAll(res.Body)
		return "API called successfully: " + string(resBody), nil
	}
}
