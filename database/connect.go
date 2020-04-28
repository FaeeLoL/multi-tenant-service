package database

import (
	"github.com/faeelol/multi-tenant-service/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	uuid "github.com/satori/go.uuid"
)

var DB *gorm.DB

func InitDB() (*gorm.DB, error) {
	db, err := gorm.Open("sqlite3", "./data/service.db")
	if err != nil {
		return nil, err
	}
	db.DB().SetMaxIdleConns(100)
	DB = db
	db.AutoMigrate(&models.Tenant{}, &models.User{}, &models.Password{}, &models.Application{})
	initRoots()
	return db, err
}

func initRoots() {
	rootTenant := models.Tenant{
		ID:              &uuid.Nil,
		Name:            "root",
		ParentId:        &uuid.Nil,
		OwnerId:         &uuid.Nil,
		AncestralAccess: false,
		Version:         0,
	}
	DB.FirstOrCreate(&rootTenant)
	rootUser := models.User{
		ID:       &uuid.Nil,
		Login:    "root",
		TenantId: &uuid.Nil,
		Role:     "tenant_admin",
		Version:  0}
	DB.FirstOrCreate(&rootUser)
	DB.FirstOrCreate(&models.Password{ID: &uuid.Nil, Password: "root"})
}
