package types

type Meta struct {
	Msg    string `json:"msg"`
	Status int    `json:"status"`
}

type LoginUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResp struct {
	Data LoginData `json:"data"`
	Meta Meta      `json:"meta"`
}

type LoginData struct {
	ID       int    `json:"id"`
	RID      int    `json:"rid"`
	Username string `json:"username"`
	Mobile   string `json:"mobile"`
	Email    string `json:"email"`
	Token    string `json:"token"`
}

type CreateUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Mobile   string `json:"mobile"`
}

type CreateUserResp struct {
	Data CreateUserData `json:"data"`
	Meta Meta           `json:"meta"`
}

type CreateUserData struct {
	ID       int    `json:"id"`
	RID      int    `json:"rid"`
	Username string `json:"username"`
	Mobile   string `json:"mobile"`
	Email    string `json:"email"`
}

type UpLoad struct {
}

type UpLoadResp struct {
	Data UpLoad `json:"data"`
	Meta Meta   `json:"meta"`
}

type PodList struct {
	PodName string `json:"podName"`
	PodIP   string `json:"podIp"`
	PodPort string `json:"podPort"`
	PodStatus string `json:"podStatus"`
	HostIP string `json:"hostIp"`
	Namespace string `json:"namespaceName"`
	ProjectName string `json:"projectName"`
	OwnerName string `json:"ownerName"`
}

type PodListResp struct {
	Data []PodList `json:"data"`
	Meta Meta      `json:"meta"`
}

type Children struct {
	PodName string `json:"podName"`
	PodIP   string `json:"podIp"`
}

type UPMSResp struct {
	Data UPMSBodyData `json:"data"`
}

type UPMSBodyData struct {
	Token string `json:"token"`
}

type ProjectDropVoResp struct {
	Status      int                  `json:"status"`
	StatusCode  string               `json:"statusCode"`
	Msg         string               `json:"msg"`
	ResultType  int                  `json:"resultType"`
	Timestamp   string               `json:"timestamp"`
	ElapsedTime int                  `json:"elapsedTime"`
	Data        []ProjectDropVoData `json:"data"`
}

type ProjectDropVoData struct {
	ProjectId   string `json:"projectId"`
	ProjectName string `json:"projectName"`
}

type ProjectServicesResp struct {
	Status int `json:"status"`
	StatusCode string `json:"statusCode"`
	Msg string `json:"msg"`
	ResultType int `json:"resultType"`
	Timestamp string `json:"timestamp"`
	ElapsedTime int `json:"elapsedTime"`
	Data ProjectServicesData `json:"data"`

}

type ProjectServicesData struct {
	Total string `json:"total"`
	Records []ProjectServicesRecords `json:"records"`
}

type ProjectServicesRecords struct {
	ID int `json:"id"`
	ServiceId string `json:"serviceId"`
	ServiceName string `json:"serviceName"`
	FullServiceName string `json:"fullServiceName"`
	ServiceDescription string `json:"serviceDescription"`
	ProjectName string `json:"projectName"`
	NamespaceName string `json:"namespaceName"`
	CreateTime string `json:"createTime"`
	LastUpdateTime string `json:"lastUpdateTime"`
}