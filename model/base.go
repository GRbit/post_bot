package model

import (
	"time"

	"gorm.io/gorm"
)

type Timestamps struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

// Base : the common properties of all gorm models
type Base struct {
	ID int64
	Timestamps
}
