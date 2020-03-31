package controllers

import (
	"fmt"
	"github.com/faeelol/multi-tenant-service/database"
	"github.com/faeelol/multi-tenant-service/models"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"net/http"
)

type TenantsController struct {
	ControllerBase
}

func (t TenantsController) CreateTenant(c *gin.Context) {
	var newTenant models.Tenant
	err := c.Bind(&newTenant)
	if err != nil {
		t.JsonFail(c, http.StatusBadRequest, "Bad request")
	}
	fmt.Printf("new tenant data: %+v\n", newTenant)

}

func isChildAvailable(parent uuid.UUID, child uuid.UUID) bool {
	if parent == child {
		return true
	}
	current := child
	for true {
		var tenant models.Tenant
		if err := database.DB.Where("id = ?", current).First(&tenant).Error; err != nil {
			return false
		}
		if !tenant.AncestralAccess {
			return false
		}
		if *tenant.ParentId == parent {
			return true
		}
		if *tenant.ParentId == *tenant.ID {
			return false
		}
	}
	return false
}
