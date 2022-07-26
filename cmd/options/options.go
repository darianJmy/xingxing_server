package options

import (
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

}

func (o *Options) Run() {
	o.Engine.Run(o.Addr)
}

func (o *Options) Login(c *gin.Context) {

}
func (o *Options) GetUsersList(c *gin.Context) {

}

func (o *Options) CreateUser(c *gin.Context) {
	var user dbstone.User
	var data dbstone.Response
	if err := c.ShouldBindJSON(&user); err != nil {
		data.Msg = "创建用户失败"
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
	c.JSON(http.StatusOK, &data)
	//	hash, err := bcrypt.GenerateFromPassword([]byte(u0.Password), bcrypt.DefaultCost) //加密处理
	//if err != nil {
	//	fmt.Println(err)
	//}
	//encodePWD := string(hash) // 保存在数据库的密码，虽然每次生成都不同，只需保存一份即可
	//fmt.Println(encodePWD)
	//
	//fmt.Println("====模拟登录====")
	//u1 := User{}
	//u1.Password = encodePWD //模拟从数据库中读取到的 经过bcrypt.GenerateFromPassword处理的密码值
	//loginPwd := "pwd"       //用户登录时输入的密码
	//// 密码验证
	//err = bcrypt.CompareHashAndPassword([]byte(u1.Password), []byte(loginPwd)) //验证（对比）
	//if err != nil {
	//	fmt.Println("pwd wrong")
	//} else {
	//	fmt.Println("pwd ok")
	//}
}

func (o *Options) GetUser(c *gin.Context) {
	var user dbstone.User
	var data dbstone.Response
	name := c.Query("name")
	user.MG_NAME = name
	d, err := o.UserDB.GetUser(&user)
	if err != nil {
		data.Msg = err.Error()
		data.Status = 400
		c.JSON(http.StatusBadRequest, data)
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
