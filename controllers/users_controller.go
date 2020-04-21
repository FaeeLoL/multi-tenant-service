package controllers

import (
	"fmt"
	"github.com/faeelol/multi-tenant-service/controllers/oauth2"
	"github.com/faeelol/multi-tenant-service/database"
	"github.com/faeelol/multi-tenant-service/models"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"strconv"
	"strings"
)

type UsersController struct {
	ControllerBase
}

func (u UsersController) CreateUser(c *gin.Context) {

	authUser := oauth2.GetAuthUserClaims(c)
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
	authUser := oauth2.GetAuthUserClaims(c)
	authTenantId, err := uuid.FromString(authUser.TenantId)
	if err != nil {
		u.JsonFail(c, http.StatusConflict, "invalid authorized tenant")
		return
	}
	var results models.UsersBatch
	results.Items = make([]models.BasicUserSchema, 0)
	var uuids []uuid.UUID
	ids := c.Request.URL.Query().Get("uuids")
	if ids == "" {
		u.JsonSuccess(c, http.StatusOK, results)
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
	for _, user := range users {
		if isChildAvailable(authTenantId, *user.TenantId) {
			results.Items = append(results.Items, user.ToBasicUserSchema())
		}
	}
	u.JsonSuccess(c, http.StatusOK, results)
}

func (u UsersController) GetSelfInfo(c *gin.Context) {
	authUser := oauth2.GetAuthUserClaims(c)
	authUserId, err := uuid.FromString(authUser.ID)
	if err != nil {
		u.JsonFail(c, http.StatusConflict, "invalid authorized tenant")
		return
	}
	var user models.User
	if err := database.DB.Where("id = ?", authUserId).Find(&user).Error; err != nil {
		panic(err)
	}
	u.JsonSuccess(c, http.StatusOK, user.ToBasicUserSchema())
}

func (u UsersController) GetUser(c *gin.Context) {
	authUser := GetAuthUserClaims(c)
	_, err := uuid.FromString(authUser.TenantId)
	if err != nil {
		u.JsonFail(c, http.StatusConflict, "invalid authorized tenant")
		return
	}
	userIdS, ok := c.Params.Get("user_id")
	if !ok {
		u.JsonFail(c, http.StatusBadRequest, "empty user_id field")
		return
	}
	userId, err := uuid.FromString(userIdS)
	if err != nil {
		u.JsonFail(c, http.StatusBadRequest, "invalid user_id format")
		return
	}
	var user models.User
	if err := database.DB.Where("id = ?", userId).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			u.JsonFail(c, http.StatusNotFound, fmt.Sprintf("The user with ID %s not found.", userIdS))
			return
		}
		panic(err)
	}
	//if !isChildAvailable(authTenantId, *user.TenantId) {
	//	u.JsonFail(c, http.StatusForbidden, fmt.Sprintf("access to user %s is forbidden", userIdS))
	//	return
	//}
	u.JsonSuccess(c, http.StatusOK, user.ToBasicUserSchema())
}

func (u UsersController) UpdateUser(c *gin.Context) {
	authUser := oauth2.GetAuthUserClaims(c)
	if authUser.Role != models.TAdmin {
		u.JsonFail(c, http.StatusForbidden, "Access is denied")
		return
	}
	authTenantId, err := uuid.FromString(authUser.TenantId)
	if err != nil {
		u.JsonFail(c, http.StatusConflict, "invalid authorized tenant")
		return
	}

	userIdS, ok := c.Params.Get("user_id")
	if !ok {
		u.JsonFail(c, http.StatusBadRequest, "empty user_id field")
		return
	}
	userId, err := uuid.FromString(userIdS)
	if err != nil {
		u.JsonFail(c, http.StatusBadRequest, "invalid user_id format")
		return
	}
	var user models.User
	tx := database.DB.Begin()
	if err := tx.Where("id = ?", userId).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			u.JsonFail(c, http.StatusNotFound, fmt.Sprintf("The user with ID %s not found.", userIdS))
			tx.Rollback()
			return
		}
		panic(err)
	}
	if !isChildAvailable(authTenantId, *user.TenantId) {
		u.JsonFail(c, http.StatusForbidden, fmt.Sprintf("access to user %s is forbidden", userIdS))
		tx.Rollback()
		return
	}

	var userInfo models.UserPut
	if err := c.Bind(&userInfo); err != nil {
		u.JsonFail(c, http.StatusBadRequest, err.Error())
		tx.Rollback()
		return
	}

	if userInfo.Version != user.Version {
		u.JsonFail(c, http.StatusConflict, "conflict in version")
		tx.Rollback()
		return
	}

	if userInfo.Role != nil {
		if *userInfo.Role == models.TAdmin || *userInfo.Role == models.TUser {
			user.Role = *userInfo.Role
		} else {
			u.JsonFail(c, http.StatusBadRequest, "Invalid user `role`")
			tx.Rollback()
			return
		}
	}
	if userInfo.Login != nil {
		if !isLoginFree(*userInfo.Login) {
			u.JsonFail(c, http.StatusBadRequest, fmt.Sprintf("Username %s is already taken", *userInfo.Login))
			tx.Rollback()
			return
		}
		user.Login = *userInfo.Login
	}

	user.Version += 1
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		panic(err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		panic(err)
	}

	u.JsonSuccess(c, http.StatusOK, user.ToBasicUserSchema())
}

func (u UsersController) DeleteUser(c *gin.Context) {
	authUser := oauth2.GetAuthUserClaims(c)
	if authUser.Role != models.TAdmin {
		u.JsonFail(c, http.StatusForbidden, "Access is denied")
		return
	}
	authTenantId, err := uuid.FromString(authUser.TenantId)
	if err != nil {
		u.JsonFail(c, http.StatusConflict, "invalid authorized tenant")
		return
	}

	userIdS, ok := c.Params.Get("user_id")
	if !ok {
		u.JsonFail(c, http.StatusBadRequest, "empty user_id field")
		return
	}
	userId, err := uuid.FromString(userIdS)
	if err != nil {
		u.JsonFail(c, http.StatusBadRequest, "invalid user_id format")
		return
	}
	versionS := c.Request.URL.Query().Get("version")
	if versionS == "" {
		u.JsonFail(c, http.StatusBadRequest, "specify `version` in query")
		return
	}
	version, err := strconv.Atoi(versionS)
	if err != nil {
		u.JsonFail(c, http.StatusBadRequest, "invalid `version` parameter")
		return
	}
	tx := database.DB.Begin()
	var user models.User
	if err := tx.Where("id = ?", userId).Find(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			u.JsonFail(c, http.StatusNotFound, fmt.Sprintf("The user with ID %s not found.", userIdS))
			tx.Rollback()
			return
		}
		tx.Rollback()
		panic(err)
	}
	if !isChildAvailable(authTenantId, *user.TenantId) {
		u.JsonFail(c, http.StatusForbidden, "access is denied")
		tx.Rollback()
		return
	}
	if version != user.Version {
		u.JsonFail(c, http.StatusConflict, "conflict in version")
		tx.Rollback()
		return
	}
	if err := tx.Delete(&user).Error; err != nil {
		tx.Rollback()
		panic(err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		panic(err)
	}
	u.JsonSuccess(c, http.StatusNoContent, nil)
}
