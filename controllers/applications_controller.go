package controllers

import (
	"fmt"
	"github.com/faeelol/multi-tenant-service/database"
	"github.com/faeelol/multi-tenant-service/models"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"strconv"
	"strings"
)

type ApplicationsController struct {
	ControllerBase
}

func (a ApplicationsController) CreateApplication(c *gin.Context) {
	var newApp models.ApplicationPost
	if err := c.Bind(&newApp); err != nil {
		a.JsonFail(c, http.StatusBadRequest, err.Error())
		return
	}
	if !isApplicationNameFree(newApp.Name) {
		a.JsonFail(c, http.StatusBadRequest, "name is already taken")
		return
	}
	newAppId, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	createdApp := models.Application{
		ID:      &newAppId,
		Name:    newApp.Name,
		Version: 1,
	}
	if err := database.DB.Create(&createdApp).Error; err != nil {
		panic(err)
	}
	a.JsonSuccess(c, http.StatusCreated, createdApp.ToBasicApplicationSchema())
}

func (a ApplicationsController) GetApplicationsBatch(c *gin.Context) {
	var results models.ApplicationsBatch
	results.Items = make([]models.BasicApplicationSchema, 0)
	var uuids []uuid.UUID
	ids := c.Request.URL.Query().Get("uuids")
	if ids == "" {
		a.JsonSuccess(c, http.StatusOK, results)
		return
	}
	for _, id := range strings.Split(ids, ",") {
		if cur, err := uuid.FromString(id); err == nil {
			uuids = append(uuids, cur)
		}
	}
	var apps []models.Application
	if err := database.DB.Where("id IN (?)", uuids).Find(&apps).Error; err != nil {
		panic(err)
	}
	for _, app := range apps {
		results.Items = append(results.Items, app.ToBasicApplicationSchema())
	}
	a.JsonSuccess(c, http.StatusOK, results)
}

func (a ApplicationsController) GetApplication(c *gin.Context) {
	/*authUser := oauth2.GetAuthUserClaims(c)
	_, err := uuid.FromString(authUser.TenantId)
	if err != nil {
		t.JsonFail(c, http.StatusConflict, "invalid authorized tenant")
		return
	}*/
	appIdS, ok := c.Params.Get("app_id")
	if !ok {
		a.JsonFail(c, http.StatusBadRequest, "empty app_id field")
		return
	}
	appId, err := uuid.FromString(appIdS)
	if err != nil {
		a.JsonFail(c, http.StatusBadRequest, "invalid app_id format")
		return
	}
	var app models.Application
	if err := database.DB.Where("id = ?", appId).First(&app).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			a.JsonFail(c, http.StatusNotFound, fmt.Sprintf("The app with ID %s not found.", appIdS))
			return
		}
		panic(err)
	}
	a.JsonSuccess(c, http.StatusOK, app.ToBasicApplicationSchema())
}

func (a ApplicationsController) UpdateApplication(c *gin.Context) {
	appIdS, ok := c.Params.Get("app_id")
	if !ok {
		a.JsonFail(c, http.StatusBadRequest, "empty app_id field")
		return
	}
	appId, err := uuid.FromString(appIdS)
	if err != nil {
		a.JsonFail(c, http.StatusBadRequest, "invalid app_id format")
		return
	}

	tx := database.DB.Begin()
	var app models.Application
	if err := tx.Where("ID = ?", appId).Find(&app).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			a.JsonFail(c, http.StatusNotFound, fmt.Sprintf("The app with ID %s not found.", appIdS))
			return
		}
		tx.Rollback()
		panic(err)
	}

	var appInfo models.ApplicationPut
	if err := c.Bind(&appInfo); err != nil {
		a.JsonFail(c, http.StatusBadRequest, err.Error())
		return
	}

	if appInfo.Version != app.Version {
		a.JsonFail(c, http.StatusConflict, "conflict in version")
		tx.Rollback()
		return
	}

	if appInfo.Name != nil {
		if !isApplicationNameFree(app.Name) {
			a.JsonFail(c, http.StatusBadRequest,
				fmt.Sprintf("application name %s is already taken ", *appInfo.Name))
			return
		}
		app.Name = *appInfo.Name
	}
	app.Version += 1
	if err := tx.Save(&app).Error; err != nil {
		tx.Rollback()
		panic(err)
	}
	if err := tx.Commit().Error; err != nil {
		panic(err)
	}
	a.JsonSuccess(c, http.StatusOK, app.ToBasicApplicationSchema())
}

func (a ApplicationsController) DeleteApplication(c *gin.Context) {
	appIdS, ok := c.Params.Get("app_id")
	if !ok {
		a.JsonFail(c, http.StatusBadRequest, "empty app_id field")
		return
	}
	appId, err := uuid.FromString(appIdS)
	if err != nil {
		a.JsonFail(c, http.StatusBadRequest, "invalid tenant_id format")
		return
	}
	versionS := c.Request.URL.Query().Get("version")
	if versionS == "" {
		a.JsonFail(c, http.StatusBadRequest, "specify `version` in query")
		return
	}
	version, err := strconv.Atoi(versionS)
	if err != nil {
		a.JsonFail(c, http.StatusBadRequest, "invalid `version` parameter")
		return
	}
	tx := database.DB.Begin()
	var app models.Application
	if err := tx.Where("id = ?", appId).Find(&app).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			a.JsonFail(c, http.StatusNotFound, fmt.Sprintf("The app with ID %s not found.", appIdS))
			tx.Rollback()
			return
		}
		tx.Rollback()
		panic(err)
	}
	if version != app.Version {
		a.JsonFail(c, http.StatusConflict, "conflict in version")
		tx.Rollback()
		return
	}
	if err := tx.Delete(&app).Error; err != nil {
		tx.Rollback()
		panic(err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		panic(err)
	}
	a.JsonSuccess(c, http.StatusNoContent, nil)
}

func isApplicationNameFree(name string) bool {
	var app models.Application
	return gorm.IsRecordNotFoundError(
		database.DB.Where("name = ?", name).First(&app).Error)
}
