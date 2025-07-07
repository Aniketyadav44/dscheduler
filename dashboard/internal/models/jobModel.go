package models

import "time"

type Job struct {
	Id        int            `json:"id"`
	Hour      int            `json:"hour"`
	Minute    int            `json:"minute"`
	Type      string         `json:"type"` //ping, email, slack, webhook
	Payload   map[string]any `json:"payload"`
	Retries   int            `json:"retries"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}
