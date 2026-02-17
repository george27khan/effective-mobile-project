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
	IsDelete    bool

	ServiceNameNil *string
	UserIdNil      *uuid.UUID
)

type Subscription struct {
	Id          IdSubs      `json:"id"`
	ServiceName ServiceName `json:"service_name"`
	Price       Price       `json:"price"`
	UserId      UserId      `json:"user_id"`
	StartDate   time.Time   `json:"start_date"`
	EndDate     *time.Time  `json:"end_date,omitempty"`
	IsDelete    IsDelete    `json:"is_delete,omitempty"`
}
