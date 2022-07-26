package options

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
	"xingxing_server/cmd/dbstone"
	"xingxing_server/cmd/middleware"
)

type Options struct {
	UserDB *dbstone.UserDB
	Engine *gin.Engine
	Addr   string
}

func (o *Options) RegisterHttpRoute() {

	o.Engine.Use(middleware.Cors())
	request := o.Engine.Group("/apis/v1")
	request.POST("/login", o.Login)
	request.GET("/users", o.GetUser)
	request.POST("/users", o.CreateUser)
	request.GET("/metrics", o.Metrics)
	request.POST("/upload", o.Upload)

}

func (o *Options) Run() {
	o.Engine.Run(o.Addr)
}

func (o *Options) Login(c *gin.Context) {
	var user dbstone.User
	var data dbstone.Response
	if err := c.ShouldBindJSON(&user); err != nil {
		data.Msg = err.Error()
		data.Status = 400
		c.JSON(http.StatusBadRequest, data)
		return
	}
	d, err := o.UserDB.GetUser(&user)
	if err != nil {
		data.Msg = err.Error()
		data.Status = 400
		c.JSON(http.StatusBadRequest, data)
		return
	}
	if user.MG_NAME == d.MG_NAME {
		if err = bcrypt.CompareHashAndPassword([]byte(d.MG_PWD), []byte(user.MG_PWD)); err != nil  {
			data.Msg = err.Error()
			data.Status = 400
			c.JSON(http.StatusBadRequest, data)
			return
		} else {
			data.ID = d.MG_ID
			data.RoleID = d.ROLE_ID
			data.Username = d.MG_NAME
			data.Mobile = d.MG_MOBILE
			data.Email = d.MG_EMAIL
			data.Msg = "登陆成功"
			data.Status = 200
			c.JSON(http.StatusOK, data)
		}
	}
}

func (o *Options) GetUsersList(c *gin.Context) {

}

func (o *Options) CreateUser(c *gin.Context) {
	var user dbstone.User
	var data dbstone.Response
	if err := c.ShouldBindJSON(&user); err != nil {
		data.Msg = err.Error()
		data.Status = 400
		c.JSON(http.StatusBadRequest, data)
		return
	}
	if &user.MG_NAME != nil &&
		&user.MG_PWD != nil &&
		&user.MG_EMAIL != nil &&
		&user.MG_MOBILE != nil &&
		user.MG_NAME != "" &&
		user.MG_PWD != "" &&
		user.MG_EMAIL != "" &&
		user.MG_MOBILE != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(user.MG_PWD), bcrypt.DefaultCost)
		if err != nil {
			data.Status = 400
			c.JSON(http.StatusBadRequest, data)
			return
		}
		encodePWD := string(hash)
		user.MG_PWD = encodePWD
	}
	user.MG_TIME = time.Now().Unix()
	user.MG_STATE = 1
	d, err := o.UserDB.CreateUser(&user)
	if err != nil {
		data.Msg = err.Error()
		data.Status = 400
		c.JSON(http.StatusBadRequest, data)
		return
	}
	data.ID = d.MG_ID
	data.Username = d.MG_NAME
	data.Mobile = d.MG_MOBILE
	data.Type = 1
	data.Email = d.MG_EMAIL
	data.OpenID = ""
	data.Create_Time = d.MG_TIME
	data.IS_Delete = false
	data.IS_Active = false
	data.Msg = "创建用户成功"
	data.Status = 200
	c.JSON(http.StatusOK, data)
}

func (o *Options) GetUser(c *gin.Context) {
	var user dbstone.User
	var data dbstone.Response
	username := c.Query("username")
	user.MG_NAME = username
	d, err := o.UserDB.GetUser(&user)
	if err != nil {
		data.Msg = err.Error()
		data.Status = 400
		c.JSON(http.StatusBadRequest, data)
		return
	}
	data.ID = d.MG_ID
	data.Username = d.MG_NAME
	data.RoleID = d.ROLE_ID
	data.Mobile = d.MG_MOBILE
	data.Email = d.MG_EMAIL
	data.Msg = "查询用户成功"
	data.Status = 200
	c.JSON(http.StatusOK, data)
}

func (o *Options) Metrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "200"})
}

func (o *Options) Upload(c *gin.Context) {
	var data dbstone.Response
	form, _ := c.MultipartForm()
	files := form.File["upload[]"]
	for _, file := range files {
		fmt.Println(file.Filename)

		// Upload the file to specific dst.
		dst := fmt.Sprintf("/Users/jimingyu/Documents/stu/xingxing_server/%s",file.Filename)
		c.SaveUploadedFile(file, dst)
	}
	data.Msg = fmt.Sprintf("上传文件成功")
	data.Status = 200
	c.JSON(http.StatusOK, data)
}