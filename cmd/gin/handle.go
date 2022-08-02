package gin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"xingxing_server/cmd/dbstone"
	"xingxing_server/cmd/gin/middleware"
)

type Options struct {
	MysqlDB *dbstone.MysqlDB
	Engine *gin.Engine
	Addr   string
}

func (o *Options) RegisterHttpRoute() {

	o.Engine.Use(middleware.Cors())
	o.Engine.Use(middleware.HandleToken())
	request := o.Engine.Group("/api/v1")
	request.POST("/login", o.Login)
	request.POST("/upload", o.Upload)
	request.GET("/users", o.GetUser)
	request.POST("/users", o.CreateUser)
	request.GET("/metrics", o.Metrics)


}

func (o *Options) Run() {
	o.Engine.Run(o.Addr)
}

func (o *Options) Login(c *gin.Context) {
	var loginUser dbstone.LoginUser
	var loginResp dbstone.LoginResp

	if err := c.ShouldBindJSON(&loginUser); err != nil {
		loginResp.Meta.Msg = fmt.Errorf("请求参数错误").Error()
		loginResp.Meta.Status = 400
		c.JSON(http.StatusOK, loginResp)
		return
	}

	result, err := o.MysqlDB.Login(loginUser.Username)
	if err != nil {
		loginResp.Meta.Msg = err.Error()
		loginResp.Meta.Status = 400
		c.JSON(http.StatusOK, loginResp)
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(result.MG_PWD), []byte(loginUser.Password)); err != nil {
		loginResp.Meta.Msg = err.Error()
		loginResp.Meta.Status = 400
		c.JSON(http.StatusOK, loginResp)
		return
	}
	token, err := middleware.SetToken()
	if err != nil {
		loginResp.Meta.Msg = fmt.Sprintln("获取Token失败")
		loginResp.Meta.Status = 400
		c.JSON(http.StatusOK, loginResp)
		return
	}
	loginResp.Data.ID = result.MG_ID
	loginResp.Data.RID = result.ROLE_ID
	loginResp.Data.Username = result.MG_NAME
	loginResp.Data.Mobile = result.MG_MOBILE
	loginResp.Data.Email = result.MG_EMAIL
	loginResp.Data.Token = token

	loginResp.Meta.Msg = fmt.Sprintln("登陆成功")
	loginResp.Meta.Status = 200

	c.JSON(http.StatusOK, loginResp)

}

func (o *Options) GetUsersList(c *gin.Context) {

}

func (o *Options) CreateUser(c *gin.Context) {
	var createUser dbstone.CreateUser
	var createUserResp dbstone.CreateUserResp
	if err := c.ShouldBindJSON(&createUser); err != nil {
		createUserResp.Meta.Msg = fmt.Errorf("请求参数错误").Error()
		createUserResp.Meta.Status = 400
		c.JSON(http.StatusBadRequest, createUserResp)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(createUser.Password), bcrypt.DefaultCost)
	if err != nil {
		createUserResp.Meta.Msg = fmt.Errorf("密码错误").Error()
		createUserResp.Meta.Status = 400
		c.JSON(http.StatusBadRequest, createUserResp)
		return
	}
	createUser.Password = string(hash)
	result, err := o.MysqlDB.CreateUser(&createUser)
	if err != nil {
		createUserResp.Meta.Msg = err.Error()
		createUserResp.Meta.Status = 400
		c.JSON(http.StatusBadRequest, createUserResp)
		return
	}
	createUserResp.Data.ID = result.MG_ID
	createUserResp.Data.RID = result.MG_ID
	createUserResp.Data.Username = result.MG_NAME
	createUserResp.Data.Mobile = result.MG_MOBILE
	createUserResp.Data.Email = result.MG_EMAIL
	createUserResp.Meta.Msg = fmt.Sprintln("创建用户成功")
	createUserResp.Meta.Status = 200
	c.JSON(http.StatusOK, createUserResp)
}

func (o *Options) GetUser(c *gin.Context) {

}

func (o *Options) Metrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "200"})
}

func (o *Options) Upload(c *gin.Context) {
	var uploadResp dbstone.UpLoadResp
	file, _ := c.FormFile("file")
	dst := fmt.Sprintf("/Users/jimingyu/Documents/stu/xingxing_server/%s",file.Filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		uploadResp.Meta.Msg = fmt.Sprintf("上传文件失败")
		uploadResp.Meta.Status = 400
		c.JSON(http.StatusBadRequest, uploadResp)
		return
	}
	uploadResp.Meta.Msg = fmt.Sprintf("上传文件成功")
	uploadResp.Meta.Status = 200
	c.JSON(http.StatusOK, uploadResp)
}

func (o *Options) Menus(c *gin.Context) {

}