package controllers

import (
	"fmt"
	"github.com/faeelol/multi-tenant-service/database"
	"github.com/faeelol/multi-tenant-service/models"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"net/http"
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
		u.JsonFail(c, http.StatusBadRequest, fmt.Sprintf("bind fail: %+v", err.Error()))
		return
	}
	parentTenant, err := uuid.FromString(authUser.TenantId)
	if err != nil {
		u.JsonFail(c, http.StatusConflict, "invalid authorized tenant")
		return
	}
	tenantId, err := uuid.FromString(newUser.TenantId)
	if err != nil {
		u.JsonFail(c, http.StatusBadRequest, "invalid tenant_id format")
		return
	}
	if !isChildAvailable(parentTenant, tenantId) {
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
