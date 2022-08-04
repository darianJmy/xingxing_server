package gin

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"strings"
	"xingxing_server/cmd/dbstone"
	"xingxing_server/cmd/gin/middleware"
	"xingxing_server/cmd/k8s"
	"xingxing_server/cmd/types"
)

var client = k8s.InitClientSet()

type Options struct {
	MysqlDB *dbstone.MysqlDB
	Engine  *gin.Engine
	Addr    string
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
	request.GET("/podList/:namespace/:projectName", o.PodList)
	request.GET("/projectList", o.ProjectList)

}

func (o *Options) Run() {
	o.Engine.Run(o.Addr)
}

func (o *Options) Login(c *gin.Context) {
	var loginUser types.LoginUser
	var loginResp types.LoginResp

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
	var createUser types.CreateUser
	var createUserResp types.CreateUserResp
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
	var uploadResp types.UpLoadResp
	file, _ := c.FormFile("file")
	dst := fmt.Sprintf("/Users/jimingyu/Documents/stu/xingxing_server/%s", file.Filename)
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

func (o *Options) PodList(c *gin.Context) {
	var podListResp types.PodListResp
	var List types.PodList
	var Children types.Children
	namespace := c.Param("namespace")
	projectName := c.Param("projectName")
	podList, err := client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"data": nil, "meta": gin.H{"msg": "获取pod列表失败", "status": http.StatusBadRequest}})
		return
	}
	for _, v := range podList.Items {
		for _, i := range v.ObjectMeta.OwnerReferences {
			_, err := client.AppsV1().ReplicaSets(namespace).Get(context.Background(), i.Name, metav1.GetOptions{})
			if err != nil {
				continue
			}

			OwnerNameList := strings.Split(i.Name, "-")
			OwnerName := strings.TrimSuffix(i.Name, OwnerNameList[len(OwnerNameList)-1])
			OwnerName = strings.TrimSuffix(OwnerName, "-")
			if projectName == OwnerName {
				List.ProjectName = projectName
				Children.PodName = v.Name
				Children.PodIP = v.Status.PodIP
				List.Children = append(List.Children, Children)
			}
		}
	}
	podListResp.Data = append(podListResp.Data, List)
	podListResp.Meta.Msg = fmt.Sprintf("获取Pod列表成功")
	podListResp.Meta.Status = http.StatusOK
	c.JSON(http.StatusOK, podListResp)
}

func (o *Options) ProjectList(c *gin.Context) {
	token, err := middleware.LoginUPMS()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"data": nil, "meta": gin.H{"msg": "认证失败", "status": http.StatusBadRequest}})
		return
	}
	req, err := http.NewRequest("GET", "http://sbx-newnoa.voneyun.com/govern/project/getProjectDropVo", nil)
	req.Header.Add("Authorization", *token)
	req.Header.Add("TENANT-ID", "1")
	var client = &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"data": nil, "meta": gin.H{"msg": "认证失败", "status": http.StatusBadRequest}})
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"data": nil, "meta": gin.H{"msg": "认证失败", "status": http.StatusBadRequest}})
		return
	}
	var serviceManagerResp types.ServiceManagerResp
	json.Unmarshal(body, &serviceManagerResp)
	c.JSON(200, serviceManagerResp)
}
