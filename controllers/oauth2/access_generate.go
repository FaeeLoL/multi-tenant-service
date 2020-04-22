package oauth2

import (
	"encoding/base64"
	"github.com/faeelol/multi-tenant-service/database"
	"github.com/faeelol/multi-tenant-service/models"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-oauth2/oauth2/generates"

	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/utils/uuid"
)

type accessGenerate struct {
	generates.JWTAccessGenerate
}

type userScope struct {
	TenantId string `json:"tenant_id,omitempty"`
	Role     string `json:"role,omitempty"`
}

type claims struct {
	jwt.StandardClaims
	Scope  userScope `json:"scope,omitempty"`
	UserId string    `json:"uid,omitempty"`
}

func newAccessGenerate(key []byte, method jwt.SigningMethod) *accessGenerate {
	return &accessGenerate{
		JWTAccessGenerate: generates.JWTAccessGenerate{
			SignedKey:    key,
			SignedMethod: method,
		},
	}
}
func (a *accessGenerate) Token(data *oauth2.GenerateBasic, isGenRefresh bool) (access, refresh string, err error) {
	claims := &claims{
		StandardClaims: jwt.StandardClaims{
			Audience:  data.Client.GetID(),
			Subject:   data.UserID,
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: data.TokenInfo.GetAccessCreateAt().Add(data.TokenInfo.GetAccessExpiresIn()).Unix(),
			Issuer:    "stub-testing-service",
		},
		Scope:  GetUserScope(data.UserID),
		UserId: data.UserID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	access, err = token.SignedString(a.SignedKey) //bug here in invalid key
	if err != nil {
		return
	}
	if isGenRefresh {
		refresh = base64.URLEncoding.EncodeToString(uuid.NewSHA1(uuid.Must(uuid.NewRandom()), []byte(access)).Bytes())
		refresh = strings.ToUpper(strings.TrimRight(refresh, "="))
	}
	return
}

func GetUserScope(userId string) userScope {
	var user models.User
	database.DB.Where("id = ?", userId).First(&user)
	return userScope{
		TenantId: user.TenantId.String(),
		Role:     user.Role,
	}
}
