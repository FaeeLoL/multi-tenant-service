package controllers

import (
	"fmt"
	"github.com/faeelol/multi-tenant-service/database"
	"github.com/faeelol/multi-tenant-service/models"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"strings"
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
		ID:              &newTenantId,
		Name:            newTenant.Name,
		ParentId:        &parentId,
		OwnerId:         &authTenantId,
		AncestralAccess: true,
		Version:         1,
	}
	if err := database.DB.Create(&createTenant).Error; err != nil {
		panic(err)
	}
	t.JsonSuccess(c, http.StatusCreated, createTenant.ToBasicTenantSchema())
}

func (t TenantsController) FetchTenantsBatch(c *gin.Context) {
	authUser := GetAuthUserClaims(c)
	authTenantId, err := uuid.FromString(authUser.TenantId)
	if err != nil {
		t.JsonFail(c, http.StatusConflict, "invalid authorized tenant")
		return
	}
	var tenants []models.Tenant
	var uuids []uuid.UUID
	ids := c.Request.URL.Query().Get("uuids")
	tenant := c.Request.URL.Query().Get("tenant_id")
	if ids != "" {
		for _, id := range strings.Split(ids, ",") {
			if cur, err := uuid.FromString(id); err == nil {
				uuids = append(uuids, cur)
			}
		}
		if err := database.DB.Where("id IN (?)", uuids).Find(&tenants).Error; err != nil {
			panic(err)
		}
	} else if tenant != "" {
		parentId, err := uuid.FromString(tenant)
		if err != nil {
			t.JsonFail(c, http.StatusBadRequest, "invalid parent_id format")
			return
		}
		if err := database.DB.Where("parent_id = ?", parentId).Find(&tenants).Error; err != nil {
			panic(err)
		}
	} else {
		t.JsonFail(c, http.StatusBadRequest, "specify `tenant_id` or `uuids` in query")
		return
	}
	var results models.TenantsBatch
	results.Items = make([]models.BasicTenantSchema, 0)
	for _, tenant := range tenants {
		if isChildAvailable(authTenantId, *tenant.ID) {
			results.Items = append(results.Items, tenant.ToBasicTenantSchema())
		}
	}
	t.JsonSuccess(c, http.StatusOK, results)
}

func (t TenantsController) GetTenant(c *gin.Context) {
	authUser := GetAuthUserClaims(c)
	authTenantId, err := uuid.FromString(authUser.TenantId)
	if err != nil {
		t.JsonFail(c, http.StatusConflict, "invalid authorized tenant")
		return
	}
	tenantIdS, ok := c.Params.Get("tenant_id")
	if !ok {
		t.JsonFail(c, http.StatusBadRequest, "empty tenant_id field")
		return
	}
	tenantId, err := uuid.FromString(tenantIdS)
	if err != nil {
		t.JsonFail(c, http.StatusBadRequest, "invalid tenant_id format")
		return
	}
	var tenant models.Tenant
	if err := database.DB.Where("id = ?", tenantId).First(&tenant).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			t.JsonFail(c, http.StatusNotFound, fmt.Sprintf("The tenant with ID %s not found.", tenantIdS))
			return
		}
		panic(err)
	}
	if !isChildAvailable(authTenantId, tenantId) {
		t.JsonFail(c, http.StatusForbidden, fmt.Sprintf("access to tenant %s is forbidden", tenantIdS))
		return
	}
	t.JsonSuccess(c, http.StatusOK, tenant.ToBasicTenantSchema())
}

func (t TenantsController) UpdateTenant(c *gin.Context) {
	authUser := GetAuthUserClaims(c)
	if authUser.Role != models.TAdmin {
		t.JsonFail(c, http.StatusForbidden, "Access is denied")
		return
	}
	authTenantId, err := uuid.FromString(authUser.TenantId)
	if err != nil {
		t.JsonFail(c, http.StatusConflict, "invalid authorized tenant")
		return
	}

	tenantIdS, ok := c.Params.Get("tenant_id")
	if !ok {
		t.JsonFail(c, http.StatusBadRequest, "empty tenant_id field")
		return
	}
	tenantId, err := uuid.FromString(tenantIdS)
	if err != nil {
		t.JsonFail(c, http.StatusBadRequest, "invalid tenant_id format")
		return
	}

	tx := database.DB.Begin()
	var tenant models.Tenant
	if err := tx.Where("ID = ?", tenantId).Find(&tenant).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			t.JsonFail(c, http.StatusNotFound, fmt.Sprintf("The tenant with ID %s not found.", tenantIdS))
			return
		}
		tx.Rollback()
		panic(err)
	}
	if !isChildAvailable(authTenantId, tenantId) {
		t.JsonFail(c, http.StatusForbidden, "access is denied")
		return
	}

	var tenantInfo models.TenantPut
	if err := c.Bind(&tenantInfo); err != nil {
		t.JsonFail(c, http.StatusBadRequest, err.Error())
		return
	}

	if tenantInfo.Version != tenant.Version {
		t.JsonFail(c, http.StatusConflict, "conflict in version")
		tx.Rollback()
		return
	}

	fmt.Printf("newInfo: %+v\n", tenantInfo)
	if tenantInfo.ParentId != nil {
		newParentId, err := uuid.FromString(*tenantInfo.ParentId)
		if err != nil {
			t.JsonFail(c, http.StatusBadRequest, "invalid parent_id format")
			return
		}
		if !isChildAvailable(authTenantId, newParentId) {
			t.JsonFail(c, http.StatusForbidden, fmt.Sprintf("access to %s is forbidden", newParentId))
			return
		}
		if tenantInfo.Name != nil {
			if !isTenantNameFree(tenant.Name, newParentId) {
				t.JsonFail(c, http.StatusBadRequest,
					fmt.Sprintf("tenant name %s is already taken under tenant %s",
						*tenantInfo.Name, newParentId))
				return
			}
		}
		tenant.ParentId = &newParentId
	}
	if tenantInfo.Name != nil {
		if !isTenantNameFree(*tenantInfo.Name, *tenant.ParentId) {
			t.JsonFail(c, http.StatusBadRequest, fmt.Sprintf("tenant name %s is already taken", *tenantInfo.Name))
			return
		}
		tenant.Name = *tenantInfo.Name
	}
	if tenantInfo.AncestralAccess != nil {
		tenant.AncestralAccess = *tenantInfo.AncestralAccess
	}
	tenant.Version += 1
	if err := tx.Save(&tenant).Error; err != nil {
		tx.Rollback()
		panic(err)
	}
	if err := tx.Commit().Error; err != nil {
		panic(err)
	}
	t.JsonSuccess(c, http.StatusOK, tenant.ToBasicTenantSchema())
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
		current = *tenant.ParentId
	}
	return false
}
