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
	Login    string `form:"login" json:"login" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

var identityKey = "id"

func (a AuthController) Init() *jwt.GinJWTMiddleware {
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "test zone",
		Key:         []byte("secret key"),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: identityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*models.Users); ok {
				return jwt.MapClaims{
					identityKey: v.Login,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &models.Users{
				Login: claims[identityKey].(string),
			}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVals loginFields
			if err := c.ShouldBind(&loginVals); err != nil {
				return "", jwt.ErrMissingLoginValues
			}
			login := loginVals.Login
			password := loginVals.Password
			var user models.Users
			if err := database.DB.Where("login = ?", login).First(&user).Error; err != nil {
				return nil, err
			}
			var storedPassword models.Passwords
			if err := database.DB.Where("id = ?", user.ID).First(&storedPassword).Error; err != nil {
				return nil, err
			}
			if password == storedPassword.Password {
				return &user, nil
			}
			return nil, jwt.ErrFailedAuthentication
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			if v, ok := data.(*models.Users); ok && v.Login == "admin" {
				return true
			}

			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"error": gin.H{
					"message": message,
				},
			})
		},

		TokenLookup: "header: Authorization, query: token, cookie: jwt",

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
