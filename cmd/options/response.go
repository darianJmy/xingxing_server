package options

import (
	"github.com/gin-gonic/gin"
	"xingxing_server/cmd/middleware"
)

package response

import (
"github.com/gin-gonic/gin"
"xingxing_server/cmd/middleware"
)


func RegisterHttpRoute() *gin.Engine {
	router := gin.Default()
	router.Use(middleware.Cors())
	router.Group("/apis/v1")
	router.POST("/login", ReturnToken)
	router.GET("/users", GetUsersList)
	router.POST("/users", CreateUser)
	router.Run(":8080")

	return router
}


func ReturnToken(c *gin.Context) {

}
func GetUsersList(c *gin.Context) {

}

func CreateUser(c *gin.Context) {

}
func(o *Options) Run() {

}

