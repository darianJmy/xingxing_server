package options

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)


type Options struct {
	DB  *gorm.DB
	Engine *gin.Engine
	Addr string
}
