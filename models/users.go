package models

import (
	"github.com/satori/go.uuid"
	"time"
)

const TAdmin = "tenant_admin"
const TUser = "tenant_user"

type User struct {
	ID        *uuid.UUID `gorm:"type:uuid;primary_key;"`
	Login     string     `json:"login"`
	TenantId  *uuid.UUID `gorm:"type:uuid";json:"tenant_id"`
	Role      string
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

type UserPost struct {
	Login    string `json:"login"`
	TenantId string `json:"tenant_id"`
	Role     string `json:"role"`
	Password string `json:"password"`
}

type BasicUserSchema struct {
	ID        string `json:"id"`
	Login     string `json:"login"`
	TenantId  string `json:"tenant_id"`
	Role      string `json:"role"`
	Version   int    `json:"version"`
	CreatedAt string `json:"CreatedAt"`
}

func (u User) ToBasicUserSchema() BasicUserSchema {
	return BasicUserSchema{
		ID:        u.ID.String(),
		Login:     u.Login,
		TenantId:  u.TenantId.String(),
		Role:      u.Role,
		Version:   u.Version,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05"),
	}
}

type AuthUser struct {
	ID       string `json:"id"`
	TenantId string `json:"tenant_id"`
	Role     string `json:"role"`
}

type Password struct {
	ID        *uuid.UUID `gorm:"type:uuid;primary_key"`
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}
