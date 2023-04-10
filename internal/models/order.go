package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

type Order struct {
	ID         string          `json:"number"`
	UserID     int             `json:"user_id,omitempty"`
	StatusID   int             `json:"status_id,omitempty"`
	Status     string          `json:"status,omitempty"`
	Accrual    sql.NullFloat64 `json:"accrual,omitempty"`
	UploadedAt time.Time       `json:"uploaded_at,omitempty"`
}

func (o *Order) MarshalJSON() ([]byte, error) {
	type Alias Order
	return json.Marshal(&struct {
		*Alias
		Accrual    float64 `json:"accrual,omitempty"`
		UploadedAt string  `json:"uploaded_at,omitempty"`
	}{
		Alias:      (*Alias)(o),
		Accrual:    o.Accrual.Float64,
		UploadedAt: o.UploadedAt.Format(time.RFC3339),
	})
}
