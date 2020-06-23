package misc

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	//"golang.org/x/oauth2/jwt"
	"net/http"
	"time"
)

const (
	TokenAccecpt = iota
	TokenTimeout
	TokenValidationFailed
)

type Claims struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	jwt.StandardClaims
}

var jwtSecret = []byte(Config.JwtSecret)

func init() {
	fmt.Println("Jwt loading...")
}

func NewJwtClaims(userID, userName string) *Claims {
	nowTime := time.Now()
	expireTime := nowTime.Add(28 * 24 * time.Hour)
	return &Claims{
		userID,
		userName,
		jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			Issuer:    "visitw",
		},
	}
}

func (claims *Claims) GenerateToken() (string, error) {
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtSecret)
	return token, err
}

func ParseJwtToken(token string) (int, *Claims, error) {

	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			if time.Now().Unix() > claims.ExpiresAt {
				return TokenTimeout, nil, nil
			}
			return TokenAccecpt, claims, nil
		}
	}
	return TokenValidationFailed, nil, err
}

// Jwt example
func VerifyJWT(c *gin.Context) {

	var state string
	type resp struct {
		FullToken        string      `json:"full_token"`
		DecryptedPayload interface{} `json:"decrypted_payload"`
	}
	var data resp

	token, err := c.Cookie("jwtAccess")
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(token)

	if token == "" {
		state = "notLogin"
		c.JSONP(http.StatusOK, gin.H{
			"state": state,
		})
		return
	} else {
		tokenState, claims, err := ParseJwtToken(token)
		if err != nil {
			fmt.Println(err)
		}
		switch tokenState {
		case TokenAccecpt:
			state = "hadLogin"
			data.FullToken = token
			data.DecryptedPayload = claims
		case TokenTimeout:
			state = "timeout"
		case TokenValidationFailed:
			state = "tokenFail"
		}
	}

	c.JSONP(http.StatusOK, gin.H{
		"state": state,
		"data":  data,
	})

}
