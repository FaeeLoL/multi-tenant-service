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

type TenantPost struct {
	Name     string `json:"name" binding:"required"`
	ParentId string `json:"parent_id" binding:"required"`
}

type BasicTenantSchema struct {
	ID              string `json:"id"`
	Version         int `json:"version"`
	Name            string `json:"name"`
	ParentId        string `json:"parent_id"`
	AncestralAccess bool   `json:"ancestral_access"`
}

func (t Tenant) ToBasicTenantSchema() BasicTenantSchema {
	return BasicTenantSchema{
		ID:              t.ID.String(),
		Version:         t.Version,
		Name:            t.Name,
		ParentId:        t.ParentId.String(),
		AncestralAccess: t.AncestralAccess,
	}
}
