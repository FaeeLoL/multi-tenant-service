package oauth2

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/faeelol/multi-tenant-service/database"
	dbModels "github.com/faeelol/multi-tenant-service/models"
	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/generates"
	"github.com/go-oauth2/oauth2/manage"
	"github.com/jinzhu/gorm"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/models"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"
	"net/http"
	"time"
)

var srv *server.Server

var defaultClientID = "000000"
var defaultClientSecret = "999999"
var defaultClientDomain = "http://localhost:8080"
var defaultTokenSecret = "token secret"

func InitOauth2() {
	manager := manage.NewDefaultManager()
	clientStore := store.NewClientStore()
	clientStore.Set(defaultClientID, &models.Client{
		ID:     defaultClientID,
		Secret: defaultClientSecret,
		Domain: defaultClientDomain,
	})
	manager.MapClientStorage(clientStore)
	manager.MustTokenStorage(store.NewFileTokenStore("tokens.db"))
	manager.MapAccessGenerate(generates.NewJWTAccessGenerate([]byte(defaultTokenSecret), jwt.SigningMethodHS256))
	srv = server.NewDefaultServer(manager)
	srv.SetClientInfoHandler(clientInfoHandler)
	srv.SetPasswordAuthorizationHandler(passwordAuthHandler)
	srv.SetAllowedGrantType(oauth2.PasswordCredentials)
}

func HandleAuthorizeRequest(c *gin.Context) {
	err := srv.HandleAuthorizeRequest(c.Writer, c.Request)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.Abort()
}

func HandleTokenRequest(c *gin.Context) {
	err := srv.HandleTokenRequest(c.Writer, c.Request)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.Abort()
}

func clientInfoHandler(r *http.Request) (clientID, clientSecret string, err error) {
	clientID = defaultClientID
	clientSecret = defaultClientSecret
	return
}

func passwordAuthHandler(username string, password string) (userID string, err error) {
	var user dbModels.User
	if err := database.DB.Where("login = ?", username).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return "", nil
		}
		return "", err
	}
	var storedPassword dbModels.Password
	if err := database.DB.Where("id = ?", user.ID).First(&storedPassword).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return "", nil
		}
		return "", err
	}
	if password == storedPassword.Password {
		return user.ID.String(), nil
	}
	return "", nil
}

func VerifyToken(c *gin.Context) {
	ti, err := srv.ValidationBearerToken(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		c.Abort()
	}

	if ti.GetAccessCreateAt().Add(ti.GetAccessExpiresIn()).Unix() < time.Now().Unix() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token is expired"})
	}
	c.Set("secure token", ti)
	c.Next()
}

func GetAuthUserClaims(c *gin.Context) dbModels.AuthUser {
	var ti oauth2.TokenInfo
	if t, ok := c.Get("secure token"); ok {
		ti = t.(oauth2.TokenInfo)
	} else {
		panic(fmt.Errorf("lost token"))
	}
	userID := ti.GetUserID()
	var user dbModels.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		panic(err)
	}
	return dbModels.AuthUser{
		ID:       userID,
		TenantId: user.TenantId.String(),
		Role:     user.Role,
	}
}
