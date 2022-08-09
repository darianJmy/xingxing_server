package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"xingxing_server/cmd/types"
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
			c.Header("Access-Control-Allow-Origin", "*") // 可将将 * 替换为指定的域名
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
	claims := &Claims{
		UserId: 2,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(), //过期时间
			IssuedAt:  time.Now().Unix(),
			Issuer:    "127.0.0.1",
			Subject:   "user token",
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

func LoginUPMS(envName string) (*string, error) {
	var req *http.Request
	var err error
	user := make(map[string]string)
	if envName == "sbx" {
		user["email"] = "mingyu.ji@sincerecloud.com"
		user["password"] = "ckr2fB8UpG15qWTmhxe2aQ=="
	} else if envName == "sit" {
		user["email"] = "mingyu.ji@sincerecloud.com"
		user["password"] = "m6yJLj7ghLE5k3EupsLzAQ=="
	} else {
		return nil, fmt.Errorf("没有此环境")
	}

	bytes, _ := json.Marshal(user)
	reqBody := strings.NewReader(string(bytes))

	if envName == "sbx" {
		req, err = http.NewRequest("POST", "http://sbx-flora.voneyun.com/upms/user/login", reqBody)
		if err != nil {
			return nil, err
		}
	} else if envName == "sit" {
		req, err = http.NewRequest("POST", "http://sit-flora.voneyun.com/upms/user/login", reqBody)
		if err != nil {
			return nil, err
		}
	}
	req.Header.Add("Content-Type", "application/json;charset=UTF-8")
	var client = &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result types.UPMSResp
	if err = json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result.Data.Token, nil
}