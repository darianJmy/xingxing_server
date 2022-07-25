package main

import (
	"github.com/gin-gonic/gin"
	"xingxing_server/cmd/options"
	"xingxing_server/cmd/response"
)

func NewHttpServer(addr string) *options.Options {
	gin.SetMode(gin.DebugMode)
	response.RegisterHttpRoute()

	o := options.Options{
		Addr: addr,
	}
	return &o
}

func main()  {
	s := NewHttpServer(":8080")
	s.Run()
}
