package models

import (
	"time"
)

type Review struct {
    ID          string    `json:"id"`           // Unique identifier
    AppID       string    `json:"app_id"`       // iTunes app ID
    Author      string    `json:"author"`
    Content     string    `json:"content"`
    Rating      int       `json:"rating"`       // Score (1-5)
    SubmittedAt time.Time `json:"submitted_at"`
    FetchedAt   time.Time `json:"fetched_at"`   // When we fetched it
}