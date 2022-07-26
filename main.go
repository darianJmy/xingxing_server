package main

import (
	"github.com/gin-gonic/gin"
	"xingxing_server/cmd/dbstone"
	"xingxing_server/cmd/options"
)

func NewHttpServer(addr string) *options.Options {
	gin.SetMode(gin.DebugMode)

	o := options.Options{
		Addr:   addr,
		Engine: gin.Default(),
		UserDB: dbstone.NewUserDB(),
	}
	o.RegisterHttpRoute()
	return &o
}

func main() {
	s := NewHttpServer(":8081")
	s.Run()
}
