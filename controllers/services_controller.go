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

type ServicesController struct {
	ControllerBase
}

func (s ServicesController) CreateService(c *gin.Context) {
	var newService models.ServicePost
	if err := c.Bind(&newService); err != nil {
		s.JsonFail(c, http.StatusBadRequest, err.Error())
		return
	}
	applicationId, err := uuid.FromString(newService.ApplicationId)
	if err != nil {
		s.JsonFail(c, http.StatusBadRequest, "invalid app_id format")
		return
	}
	app := models.Application{}
	if err := database.DB.Where("id = ?", applicationId).First(&app).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			s.JsonFail(c, http.StatusNotFound, fmt.Sprintf("The application with app_ID %s not found.", newService.ApplicationId))
			return
		}
		panic(err)
		return
	}
	if !isServiceNameFree(newService.Name, applicationId) {
		s.JsonFail(c, http.StatusBadRequest, "name is already taken")
		return
	}
	newServiceId, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	createdService := models.Service{
		ID:            &newServiceId,
		Name:          newService.Name,
		Version:       1,
		ApplicationId: &applicationId,
	}
	if err := database.DB.Create(&createdService).Error; err != nil {
		panic(err)
	}
	s.JsonSuccess(c, http.StatusCreated, createdService.ToBasicServiceSchema())
}

func (s ServicesController) GetServicesBatch(c *gin.Context) {
	var results models.ServicesBatch
	results.Items = make([]models.BasicServiceSchema, 0)
	var uuids []uuid.UUID
	ids := c.Request.URL.Query().Get("uuids")
	if ids == "" {
		s.JsonSuccess(c, http.StatusOK, results)
		return
	}
	for _, id := range strings.Split(ids, ",") {
		if cur, err := uuid.FromString(id); err == nil {
			uuids = append(uuids, cur)
		}
	}
	var services []models.Service
	if err := database.DB.Where("id IN (?)", uuids).Find(&services).Error; err != nil {
		panic(err)
	}
	for _, service := range services {
		results.Items = append(results.Items, service.ToBasicServiceSchema())
	}
	s.JsonSuccess(c, http.StatusOK, results)
}

func (s ServicesController) GetService(c *gin.Context) {
	/*authUser := oauth2.GetAuthUserClaims(c)
	_, err := uuid.FromString(authUser.TenantId)
	if err != nil {
		t.JsonFail(c, http.StatusConflict, "invalid authorized tenant")
		return
	}*/
	serviceIdS, ok := c.Params.Get("service_id")
	if !ok {
		s.JsonFail(c, http.StatusBadRequest, "empty service_id field")
		return
	}
	serviceId, err := uuid.FromString(serviceIdS)
	if err != nil {
		s.JsonFail(c, http.StatusBadRequest, "invalid service_id format")
		return
	}
	//add transaction?
	var service models.Service
	if err := database.DB.Where("id = ?", serviceId).First(&service).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			s.JsonFail(c, http.StatusNotFound, fmt.Sprintf("The service with ID %s not found.", serviceIdS))
			return
		}
		panic(err)
	}
	s.JsonSuccess(c, http.StatusOK, service.ToBasicServiceSchema())
}

func (s ServicesController) UpdateService(c *gin.Context) {
	serviceIdS, ok := c.Params.Get("service_id")
	if !ok {
		s.JsonFail(c, http.StatusBadRequest, "empty service_id field")
		return
	}
	serviceId, err := uuid.FromString(serviceIdS)
	if err != nil {
		s.JsonFail(c, http.StatusBadRequest, "invalid service_id format")
		return
	}

	tx := database.DB.Begin()
	var service models.Service
	if err := tx.Where("ID = ?", serviceId).Find(&service).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			s.JsonFail(c, http.StatusNotFound, fmt.Sprintf("The service with ID %s not found.", serviceIdS))
			return
		}
		tx.Rollback()
		panic(err)
	}

	var serviceInfo models.ServicePut
	if err := c.Bind(&serviceInfo); err != nil {
		s.JsonFail(c, http.StatusBadRequest, err.Error())
		return
	}

	if serviceInfo.Version != service.Version {
		s.JsonFail(c, http.StatusConflict, "conflict in version")
		tx.Rollback()
		return
	}

	applicationId, err := uuid.FromString(serviceInfo.ApplicationId)
	if err != nil {
		s.JsonFail(c, http.StatusBadRequest, "invalid app_id format")
		return
	}
	app := models.Application{}
	if err := tx.Where("ID = ?", applicationId).Find(&app).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			s.JsonFail(c, http.StatusNotFound, fmt.Sprintf("The application with app_ID %s not found.", serviceInfo.ApplicationId))
			return
		}
		tx.Rollback()
		panic(err)
	}
	*service.ApplicationId = applicationId //a

	if serviceInfo.Name != nil {
		if !isServiceNameFree(service.Name, applicationId) {
			s.JsonFail(c, http.StatusBadRequest,
				fmt.Sprintf("Service name %s is already taken ", *serviceInfo.Name))
			return
		}
		service.Name = *serviceInfo.Name
	}

	service.Version += 1
	if err := tx.Save(&service).Error; err != nil {
		tx.Rollback()
		panic(err)
	}
	if err := tx.Commit().Error; err != nil {
		panic(err)
	}
	s.JsonSuccess(c, http.StatusOK, service.ToBasicServiceSchema())
}

func (s ServicesController) DeleteService(c *gin.Context) {
	serviceIdS, ok := c.Params.Get("service_id")
	if !ok {
		s.JsonFail(c, http.StatusBadRequest, "empty service_id field")
		return
	}
	serviceId, err := uuid.FromString(serviceIdS)
	if err != nil {
		s.JsonFail(c, http.StatusBadRequest, "invalid service_id format")
		return
	}
	versionS := c.Request.URL.Query().Get("version")
	if versionS == "" {
		s.JsonFail(c, http.StatusBadRequest, "specify `version` in query")
		return
	}
	version, err := strconv.Atoi(versionS)
	if err != nil {
		s.JsonFail(c, http.StatusBadRequest, "invalid `version` parameter")
		return
	}
	tx := database.DB.Begin()
	var service models.Service
	if err := tx.Where("id = ?", serviceId).Find(&service).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			s.JsonFail(c, http.StatusNotFound, fmt.Sprintf("The service with ID %s not found.", serviceIdS))
			tx.Rollback()
			return
		}
		tx.Rollback()
		panic(err)
	}
	if version != service.Version {
		s.JsonFail(c, http.StatusConflict, "conflict in version")
		tx.Rollback()
		return
	}
	if err := tx.Delete(&service).Error; err != nil {
		tx.Rollback()
		panic(err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		panic(err)
	}
	s.JsonSuccess(c, http.StatusNoContent, nil)
}

func isServiceNameFree(name string, appId uuid.UUID) bool {
	var app models.Application
	return gorm.IsRecordNotFoundError(
		database.DB.Where("name = ? AND app_id = ?", name, appId).First(&app).Error)
}
