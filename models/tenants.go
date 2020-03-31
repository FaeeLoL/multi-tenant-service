package models

import (
	"github.com/satori/go.uuid"
	"time"
)

type Tenant struct {
	ID              *uuid.UUID `gorm:"type:uuid;primary_key;"`
	Name            string
	ParentId        *uuid.UUID `gorm:"type:uuid"`
	OwnerId         *uuid.UUID `gorm:"type:uuid"`
	AncestralAccess bool
	Version         int
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time `sql:"index"`
}
