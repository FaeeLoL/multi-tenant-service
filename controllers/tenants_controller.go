package controllers

import (
	"github.com/faeelol/multi-tenant-service/database"
	"github.com/faeelol/multi-tenant-service/models"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"net/http"
)

type TenantsController struct {
	ControllerBase
}

func (t TenantsController) CreateTenant(c *gin.Context) {
	authUser := GetAuthUserClaims(c)
	if authUser.Role != models.TAdmin {
		t.JsonFail(c, http.StatusForbidden, "Access is denied")
		return
	}

	var newTenant models.TenantPost
	if err := c.Bind(&newTenant); err != nil {
		t.JsonFail(c, http.StatusBadRequest, err.Error())
		return
	}

	authTenantId, err := uuid.FromString(authUser.TenantId)
	if err != nil {
		t.JsonFail(c, http.StatusConflict, "invalid authorized tenant")
		return
	}
	parentId, err := uuid.FromString(newTenant.ParentId)
	if err != nil {
		t.JsonFail(c, http.StatusBadRequest, "invalid parent_id format")
		return
	}
	if !isChildAvailable(authTenantId, parentId) {
		t.JsonFail(c, http.StatusForbidden, "access is denied")
		return
	}
	if !isTenantNameFree(newTenant.Name, parentId) {
		t.JsonFail(c, http.StatusBadRequest, "name is already taken")
		return
	}

	newTenantId, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	createTenant := models.Tenant{
		ID: &newTenantId,
		Name: newTenant.Name,
		ParentId: &parentId,
		OwnerId: &authTenantId,
		AncestralAccess: true,
		Version: 1,
	}
	if err := database.DB.Create(&createTenant).Error; err != nil {
		panic(err)
	}
	t.JsonSuccess(c, http.StatusCreated, createTenant.ToBasicTenantSchema())
}

func isTenantNameFree(name string, parentId uuid.UUID) bool {
	var user models.Tenant
	return gorm.IsRecordNotFoundError(
		database.DB.Where("name = ? AND parent_id = ?", name, parentId).First(&user).Error)
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
