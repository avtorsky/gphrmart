package models

import (
	"encoding/json"
	"time"
)

type Withdrawal struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

func (w *Withdrawal) MarshalJSON() ([]byte, error) {
	type Alias Withdrawal
	return json.Marshal(&struct {
		*Alias
		ProcessedAt string `json:"processed_at"`
	}{
		Alias:       (*Alias)(w),
		ProcessedAt: w.ProcessedAt.Format(time.RFC3339),
	})
}
