package dbstone

type LoginUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResp struct {
	Data LoginData
	Meta Meta
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
	Data CreateUserData
	Meta Meta
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
	Data UpLoad
	Meta Meta
}

type Meta struct {
	Msg    string `json:"msg"`
	Status int    `json:"status"`
}
