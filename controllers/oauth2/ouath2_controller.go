package oauth2

import (
	"fmt"
	"net/http"
	"time"

	"github.com/faeelol/multi-tenant-service/database"
	dbModels "github.com/faeelol/multi-tenant-service/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/manage"
	"github.com/jinzhu/gorm"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/models"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"
)

var srv *server.Server

const defaultClientID = "000000"
const defaultClientSecret = "999999"
const defaultClientDomain = "http://localhost:8080"
const defaultTokenSecret = "token secret"
const tokenExpTime = 2 * time.Hour

func InitOauth2() {
	manager := manage.NewDefaultManager()
	clientStore := store.NewClientStore()
	clientStore.Set(defaultClientID, &models.Client{
		ID:     defaultClientID,
		Secret: defaultClientSecret,
		Domain: defaultClientDomain,
	})
	manager.MapClientStorage(clientStore)
	manager.MustTokenStorage(store.NewFileTokenStore("./data/tokens.db"))
	manager.MapAccessGenerate(newAccessGenerate([]byte(defaultTokenSecret), jwt.SigningMethodHS256))
	manager.SetPasswordTokenCfg(&manage.Config{
		AccessTokenExp:    tokenExpTime,
		RefreshTokenExp:   tokenExpTime,
		IsGenerateRefresh: true,
	})
	srv = server.NewDefaultServer(manager)
	srv.SetClientInfoHandler(clientInfoHandler)
	srv.SetPasswordAuthorizationHandler(passwordAuthHandler)
	srv.SetAllowedGrantType(oauth2.PasswordCredentials)
	//srv.SetExtensionFieldsHandler(func(ti oauth2.TokenInfo) (fieldsValue map[string]interface{}){
	//	jwt.Parse()
	//})

	// set expires_on field
	srv.SetExtensionFieldsHandler(func(ti oauth2.TokenInfo) (fieldsValue map[string]interface{}) {
		fieldsValue = map[string]interface{}{
			"expires_on": time.Now().Add(tokenExpTime).Unix(),
		}
		return
	})
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
	var cl claims
	tk, err := jwt.ParseWithClaims(ti.GetAccess(), &cl, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("parse error")
		}
		return []byte(defaultTokenSecret), nil
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token parse error"})
		c.Abort()
	}
	if ti.GetAccessCreateAt().Add(ti.GetAccessExpiresIn()).Unix() < time.Now().Unix() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token is expired"})
		c.Abort()
	}
	c.Set("oauth2_token", ti)
	c.Set("jwt_token", tk)
	c.Set("claims", cl)
	c.Next()
}

func GetAuthUserClaims(c *gin.Context) dbModels.AuthUser {
	var cl claims
	if v, ok := c.Get("claims"); ok {
		cl = v.(claims)
	} else {
		panic(fmt.Errorf("invalid claims"))
	}
	return dbModels.AuthUser{
		ID:       cl.UserId,
		TenantId: cl.Scope.TenantId,
		Role:     cl.Scope.Role,
	}
}
