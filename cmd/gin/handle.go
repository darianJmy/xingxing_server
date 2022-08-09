package gin

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"strconv"
	"strings"
	"xingxing_server/cmd/dbstone"
	"xingxing_server/cmd/gin/middleware"
	"xingxing_server/cmd/k8s"
	"xingxing_server/cmd/types"
)

var sitClient = k8s.SitInitClientSet()
var sbxClient = k8s.SbxInitClientSet()

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
	request.GET("/sbxPodList/:namespace/:projectName", o.SbxPodList)
	request.GET("/sitPodList/:namespace/:projectName", o.SitPodList)
	request.GET("/getPodLogs/:namespace/:podName", o.GetPodLogs)
	request.GET("/getProjectDropVo/:envName", o.GetProjectDropVo)
	request.GET("/getProjectServices/:envName", o.GetProjectServices)

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

func (o *Options) SbxPodList(c *gin.Context) {
	var podListResp types.PodListResp
	var List types.PodList
	namespace := c.Param("namespace")
	projectName := c.Param("projectName")
	podList, err := sbxClient.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"data": nil, "meta": gin.H{"msg": "获取pod列表失败", "status": http.StatusBadRequest}})
		return
	}
	for _, v := range podList.Items {
		for _, i := range v.ObjectMeta.OwnerReferences {
			_, err := sbxClient.AppsV1().ReplicaSets(namespace).Get(context.Background(), i.Name, metav1.GetOptions{})
			if err != nil {
				continue
			}

			OwnerNameList := strings.Split(i.Name, "-")
			OwnerName := strings.TrimSuffix(i.Name, OwnerNameList[len(OwnerNameList)-1])
			OwnerName = strings.TrimSuffix(OwnerName, "-")
			projectNameList := strings.Split(projectName, "-")
			if projectNameList[0] == OwnerNameList[0] &&
				projectNameList[1] == OwnerNameList[1] {
				List.PodName = v.Name
				List.PodIP = v.Status.PodIP
				List.HostIP = v.Status.HostIP
				List.PodStatus = string(v.Status.Phase)
				List.Namespace = v.Namespace
				List.ProjectName = projectNameList[0]
				List.OwnerName = OwnerName
				podListResp.Data = append(podListResp.Data, List)
			}
		}
	}
	podListResp.Meta.Msg = fmt.Sprintf("获取Pod列表成功")
	podListResp.Meta.Status = http.StatusOK
	c.JSON(http.StatusOK, podListResp)
}

func (o *Options) SitPodList(c *gin.Context) {
	var podListResp types.PodListResp
	var List types.PodList
	namespace := c.Param("namespace")
	projectName := c.Param("projectName")
	podList, err := sitClient.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"data": nil, "meta": gin.H{"msg": "获取pod列表失败", "status": http.StatusBadRequest}})
		return
	}
	for _, v := range podList.Items {
		for _, i := range v.ObjectMeta.OwnerReferences {
			_, err := sitClient.AppsV1().ReplicaSets(namespace).Get(context.Background(), i.Name, metav1.GetOptions{})
			if err != nil {
				continue
			}

			OwnerNameList := strings.Split(i.Name, "-")
			OwnerName := strings.TrimSuffix(i.Name, OwnerNameList[len(OwnerNameList)-1])
			OwnerName = strings.TrimSuffix(OwnerName, "-")
			projectNameList := strings.Split(projectName, "-")
			if projectNameList[0] == OwnerNameList[0] &&
				projectNameList[1] == OwnerNameList[1] {
				List.PodName = v.Name
				List.PodIP = v.Status.PodIP
				List.HostIP = v.Status.HostIP
				List.PodStatus = string(v.Status.Phase)
				List.Namespace = v.Namespace
				List.ProjectName = projectNameList[0]
				List.OwnerName = OwnerName
				podListResp.Data = append(podListResp.Data, List)
			}
		}
	}
	podListResp.Meta.Msg = fmt.Sprintf("获取Pod列表成功")
	podListResp.Meta.Status = http.StatusOK
	c.JSON(http.StatusOK, podListResp)
}

func (o *Options) GetProjectDropVo(c *gin.Context) {
	envName := c.Param("envName")
	if envName == "" {
		c.JSON(http.StatusOK, gin.H{"data": nil, "meta": gin.H{"msg": "没有此环境", "status": http.StatusBadRequest}})
		return
	}
	token, err := middleware.LoginUPMS(envName)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"data": nil, "meta": gin.H{"msg": err.Error(), "status": http.StatusBadRequest}})
		return
	}
	var req *http.Request
	if envName == "sbx" {
		req, err = http.NewRequest("GET", "http://sbx-newnoa.voneyun.com/govern/project/getProjectDropVo", nil)
	} else {
		req, err = http.NewRequest("GET", "http://sit-newnoa.voneyun.com/govern/project/getProjectDropVo", nil)
	}
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
	var projectDropVoResp types.ProjectDropVoResp
	json.Unmarshal(body, &projectDropVoResp)
	c.JSON(200, projectDropVoResp)
}

func (o *Options) GetProjectServices(c *gin.Context) {
	envName := c.Param("envName")
	if envName == "" {
		c.JSON(http.StatusOK, gin.H{"data": nil, "meta": gin.H{"msg": "没有此环境", "status": http.StatusBadRequest}})
		return
	}

	var url string
	var token *string
	var err error
	projectId := c.Query("projectId")
	if envName == "sbx" && projectId != "" {
		url = fmt.Sprintf("http://sbx-newnoa.voneyun.com/govern/services/page?projectId=%s", projectId)
		token, err = middleware.LoginUPMS(envName)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"data": nil, "meta": gin.H{"msg": "认证失败", "status": http.StatusBadRequest}})
			return
		}
	} else if envName == "sit" && projectId != "" {
		url = fmt.Sprintf("http://sit-newnoa.voneyun.com/govern/services/page?projectId=%s", projectId)
		token, err = middleware.LoginUPMS(envName)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"data": nil, "meta": gin.H{"msg": "认证失败", "status": http.StatusBadRequest}})
			return
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"data": nil, "meta": gin.H{"msg": "认证失败", "status": http.StatusBadRequest}})
		return
	}

	req, err := http.NewRequest("GET", url, nil)
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
	var projectServicesResp  types.ProjectServicesResp
	json.Unmarshal(body, &projectServicesResp)
	c.JSON(200, projectServicesResp)
}

func (o *Options) GetPodLogs(c *gin.Context) {
	namespace := c.Param("namespace")
	podName := c.Param("podName")
	//container := c.Query("container")
	tailLines, _ := strconv.ParseInt(c.DefaultQuery("tailLines", "500"), 10, 64)
	timeStamps, _ := strconv.ParseBool(c.DefaultQuery("timeStamps", "true"))
	previous, _ := strconv.ParseBool(c.DefaultQuery("previous", "false"))

	if namespace == "" || podName == "" {
		c.JSON(http.StatusOK, gin.H{"data": nil, "meta": gin.H{"msg": "获取日志失败", "status": http.StatusBadRequest}})
		return
	}

	kubeLogger, err := k8s.NewKubeLogger(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"data": nil, "meta": gin.H{"msg": "升级日志失败", "status": http.StatusBadRequest}})
		return
	}

	opts := corev1.PodLogOptions{
		Timestamps: timeStamps,
		Previous: previous,
		Follow: true,
		TailLines: &tailLines,
	}
	req := sbxClient.CoreV1().Pods(namespace).GetLogs(podName, &opts)
	stream, err := req.Stream(context.Background())
	if err != nil {
		kubeLogger.Write([]byte(err.Error()))
		return
	}
	defer stream.Close()

	buf := bufio.NewReader(stream)
	for {
		bytes, err := buf.ReadBytes('\n')
		if err != nil {
			kubeLogger.Write([]byte(err.Error()))
			return
		}
		if err := kubeLogger.Write(bytes); err != nil {
			kubeLogger.Write([]byte(err.Error()))
			return
		}
	}
}
