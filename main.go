package main

import (
	"github.com/gin-gonic/gin"
	"xingxing_server/cmd/dbstone"
	gin2 "xingxing_server/cmd/gin"
)

func NewHttpServer(addr string) *gin2.Options {
	gin.SetMode(gin.DebugMode)

	o := gin2.Options{
		Addr:    addr,
		Engine:  gin.Default(),
		MysqlDB: dbstone.NewMysqlDB(),
	}
	o.RegisterHttpRoute()
	return &o
}

func main() {
	s := NewHttpServer(":8081")
	s.Run()
}
