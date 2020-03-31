package models

import (
	"github.com/satori/go.uuid"
	"time"
)

type Users struct {
	ID        *uuid.UUID `gorm:"type:uuid;primary_key;"`
	Login     string
	TenantId  *uuid.UUID `gorm:"type:uuid"`
	Role      string
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

type Passwords struct {
	ID        *uuid.UUID `gorm:"type:uuid;primary_key"`
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}
