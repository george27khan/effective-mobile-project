package entity

import (
	"github.com/google/uuid"
	"time"
)

type (
	IdSubs      int
	ServiceName string
	Price       *int
	UserId      uuid.UUID
)

type Subscription struct {
	Id          IdSubs      `json:"id"`
	ServiceName ServiceName `json:"service_name"`
	Price       Price       `json:"price"`
	UserId      UserId      `json:"user_id"`
	StartDate   time.Time   `json:"start_date"`
	EndDate     *time.Time  `json:"end_date,omitempty"`
	DeleteDate  *time.Time  `json:"delete_date,omitempty"`
}
