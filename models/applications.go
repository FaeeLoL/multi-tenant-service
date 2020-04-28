package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

type Application struct {
	ID        *uuid.UUID `gorm:"type:uuid;primary_key;"`
	Name      string
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}
type BasicApplicationSchema struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Name    string `json:"name"`
}

func (a Application) ToBasicApplicationSchema() BasicApplicationSchema {
	return BasicApplicationSchema{
		ID:      a.ID.String(),
		Version: a.Version,
		Name:    a.Name,
	}
}

type ApplicationsBatch struct {
	Items []BasicApplicationSchema `json:"items"`
}

type ApplicationPut struct {
	Name    *string `json:"name"`
	Version int     `json:"version" binding:"required"`
}
type ApplicationPost struct {
	Name string `json:"name" binding:"required"`
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
