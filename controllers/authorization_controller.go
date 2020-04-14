package controllers

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/faeelol/multi-tenant-service/database"
	"github.com/faeelol/multi-tenant-service/models"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

type AuthController struct {
	ControllerBase
}

type loginFields struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

var identityKey = "id"

func (a AuthController) Init() *jwt.GinJWTMiddleware {
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "test zone",
		Key:         []byte("secret key"), //todo get from env
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: identityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*models.AuthUser); ok {
				return jwt.MapClaims{
					identityKey: v.ID,
					"tenant_id": v.TenantId,
					"role":      v.Role,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &models.AuthUser{
				ID: claims[identityKey].(string),
			}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVals loginFields
			if err := c.ShouldBind(&loginVals); err != nil {
				return "", jwt.ErrMissingLoginValues
			}
			login := loginVals.Username
			password := loginVals.Password
			var user models.User
			if err := database.DB.Where("login = ?", login).First(&user).Error; err != nil {
				return nil, err
			}
			var storedPassword models.Password
			if err := database.DB.Where("id = ?", user.ID).First(&storedPassword).Error; err != nil {
				return nil, err
			}
			if password == storedPassword.Password {
				return &models.AuthUser{
						ID:       user.ID.String(),
						TenantId: user.TenantId.String(),
						Role:     user.Role},
					nil
			}
			return nil, jwt.ErrFailedAuthentication
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			return true
		},
		LoginResponse: func(c *gin.Context, code int, token string, expire time.Time) {
			c.JSON(code, gin.H{
				"access_token" : token,
				"token_type": "bearer",
				"expires_on": expire.Unix(),
			})
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"error": gin.H{
					"message": message,
				},
			})
		},
		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		// - "param:<name>"
		TokenLookup: "header: Authorization, query: token, cookie: jwt",
		// TokenLookup: "query:token",
		// TokenLookup: "cookie:token",

		// TokenHeadName is a string in the header. Default value is "Bearer"
		TokenHeadName: "Bearer",

		// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
		TimeFunc: time.Now,
	})
	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}
	return authMiddleware
}

func GetAuthUserClaims(c *gin.Context) models.AuthUser {
	claims := jwt.ExtractClaims(c)
	return models.AuthUser{
		ID:       claims["id"].(string),
		TenantId: claims["tenant_id"].(string),
		Role:     claims["role"].(string),
	}
}
