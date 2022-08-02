package middleware

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

var jwtkey = []byte("MyNameIsJiXingXing")
var str string

type Claims struct {
	UserId uint
	jwt.StandardClaims
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", "*")  // 可将将 * 替换为指定的域名
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

func HandleToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()
		if path != "/api/v1/login" {
			tokenString := c.GetHeader("Authorization")
			//vcalidate token formate
			if tokenString == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"status": 401, "msg": "权限不足"})
				c.Abort()
				return
			}

			token, _, err := ParseToken(tokenString)
			if err != nil || !token.Valid {
				c.JSON(http.StatusUnauthorized, gin.H{"status": 401, "msg": "权限不足"})
				c.Abort()
				return
			}
			c.Next()
		}
	}
}

func SetToken() (string, error) {
	expireTime := time.Now().Add(7 * 24 * time.Hour)
	claims := &Claims {
		UserId: 2,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(), //过期时间
			IssuedAt: time.Now().Unix(),
			Issuer: "127.0.0.1",
			Subject: "user token",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtkey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
func ParseToken(tokenString string) (*jwt.Token, *Claims, error) {
	Claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, Claims, func(token *jwt.Token) (i interface{}, err error) {
		return jwtkey, nil
	})
	return token, Claims, err
}
