package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

type Service struct {
	ID            *uuid.UUID `gorm:"type:uuid;primary_key;"`
	Name          string
	Version       int
	ApplicationId *uuid.UUID `gorm:"type:uuid" json:"app_id"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time `sql:"index"`
}
type BasicServiceSchema struct {
	ID            string `json:"id"`
	Version       int    `json:"version"`
	Name          string `json:"name"`
	ApplicationId string `json:"app_id"`
}

func (s Service) ToBasicServiceSchema() BasicServiceSchema {
	return BasicServiceSchema{
		ID:            s.ID.String(),
		Version:       s.Version,
		Name:          s.Name,
		ApplicationId: s.ApplicationId.String(),
	}
}

type ServicesBatch struct {
	Items []BasicServiceSchema `json:"items"`
}

type ServicePut struct {
	Name          *string `json:"name"`
	Version       int     `json:"version" binding:"required"`
	ApplicationId string  `json:"app_id"`
}
type ServicePost struct {
	Name          string `json:"name" binding:"required"`
	ApplicationId string `json:"app_id"  binding:"required"`
}

/*

}

type TenantPut struct {
	Name            *string `json:"name"`
	ParentId        *string `json:"parent_id"`
	AncestralAccess *bool   `json:"ancestral_access"`
	Version         int     `json:"version" binding:"required"`
}







*/
