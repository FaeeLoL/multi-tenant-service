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

type TenantPut struct {
	Name            *string `json:"name"`
	ParentId        *string `json:"parent_id"`
	AncestralAccess *bool   `json:"ancestral_access"`
	Version         int     `json:"version" binding:"required"`
}

type BasicTenantSchema struct {
	ID              string `json:"id"`
	Version         int    `json:"version"`
	Name            string `json:"name"`
	OwnerId         string `json:"owner_id"`
	ParentId        string `json:"parent_id"`
	AncestralAccess bool   `json:"ancestral_access"`
}

type TenantsBatch struct {
	Items []BasicTenantSchema `json:"items"`
}

func (t Tenant) ToBasicTenantSchema() BasicTenantSchema {
	return BasicTenantSchema{
		ID:              t.ID.String(),
		Version:         t.Version,
		Name:            t.Name,
		OwnerId:         t.OwnerId.String(),
		ParentId:        t.ParentId.String(),
		AncestralAccess: t.AncestralAccess,
	}
}
