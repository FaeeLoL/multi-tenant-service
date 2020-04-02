package controllers

import (
	"github.com/faeelol/multi-tenant-service/database"
	"github.com/faeelol/multi-tenant-service/models"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"strings"
)

type UsersController struct {
	ControllerBase
}

func (u UsersController) CreateUser(c *gin.Context) {

	authUser := GetAuthUserClaims(c)
	if authUser.Role != models.TAdmin {
		u.JsonFail(c, http.StatusForbidden, "Access is denied")
		return
	}
	var newUser models.UserPost
	if err := c.Bind(&newUser); err != nil {
		u.JsonFail(c, http.StatusBadRequest, err.Error())
		return
	}
	authTenantId, err := uuid.FromString(authUser.TenantId)
	if err != nil {
		u.JsonFail(c, http.StatusConflict, "invalid authorized tenant")
		return
	}
	tenantId, err := uuid.FromString(newUser.TenantId)
	if err != nil {
		u.JsonFail(c, http.StatusBadRequest, "invalid tenant_id format")
		return
	}
	if !isChildAvailable(authTenantId, tenantId) {
		u.JsonFail(c, http.StatusForbidden, "access is denied")
		return
	}
	if !isLoginFree(newUser.Login) {
		u.JsonFail(c, http.StatusBadRequest, "login is already taken")
		return
	}
	if newUser.Role != models.TAdmin {
		newUser.Role = models.TUser
	} else {
		newUser.Role = models.TAdmin
	}
	if newUser.Password == "" {
		u.JsonFail(c, http.StatusBadRequest, "password field is empty")
		return
	}
	newUUID, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	user := models.User{ID: &newUUID, Login: newUser.Login, TenantId: &tenantId, Role: newUser.Role, Version: 1}
	tx := database.DB.Begin()
	if err := tx.Error; err != nil {
		panic(err)
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		panic(err)
	}
	password := models.Password{ID: user.ID, Password: newUser.Password}
	if err := tx.Create(&password).Error; err != nil {
		tx.Rollback()
		panic(err)
	}
	if err := tx.Commit().Error; err != nil {
		panic(err)
	}
	u.JsonSuccess(c, http.StatusCreated, user.ToBasicUserSchema())
}

func isLoginFree(login string) bool {
	var user models.User
	return gorm.IsRecordNotFoundError(database.DB.Where("login = ?", login).First(&user).Error)
}

func (u UsersController) GetUsersBatch(c *gin.Context) {
	authUser := GetAuthUserClaims(c)
	authTenantId, err := uuid.FromString(authUser.TenantId)
	if err != nil {
		u.JsonFail(c, http.StatusConflict, "invalid authorized tenant")
		return
	}
	var uuids []uuid.UUID
	ids := c.Request.URL.Query().Get("uuids")
	if ids == "" {
		u.JsonSuccess(c, http.StatusOK, models.UsersBatch{})
		return
	}
	for _, id := range strings.Split(ids, ",") {
		if cur, err := uuid.FromString(id); err == nil {
			uuids = append(uuids, cur)
		}
	}
	var users []models.User
	if err := database.DB.Where("id IN (?)", uuids).Find(&users).Error; err != nil {
		panic(err)
	}
	var results models.UsersBatch
	for _, user := range users {
		if isChildAvailable(authTenantId, *user.TenantId) {
			results.Items = append(results.Items, user.ToBasicUserSchema())
		}
	}
	u.JsonSuccess(c, http.StatusOK, results)
}
