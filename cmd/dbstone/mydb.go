package dbstone

type User struct {
	MG_ID     int    `json:"mg_id"`
	MG_NAME   string `json:"mg_name"`
	MG_PWD    string `json:"mg_pwd"`
	MG_TIME   int64  `json:"mg_time"`
	ROLE_ID   int    `json:"role_id"`
	MG_STATE  int    `json:"mg_state"`
	MG_MOBILE string `json:"mg_mobile"`
	MG_EMAIL  string `json:"mg_email"`
}

type Response struct {
	Data
	Meta
}

type Data struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	Mobile      string `json:"mobile"`
	Type        int    `json:"type"`
	OpenID      string `json:"openid"`
	Email       string `json:"email"`
	RoleID      int    `json:"role_id"`
	Create_Time int64  `json:"create_time"`
	Modify_Time string `json:"modify_time"`
	IS_Delete   bool   `json:"is_delete"`
	IS_Active   bool   `json:"is_active"`
}

type Meta struct {
	Msg    string `json:"msg"`
	Status int    `json:"status"`
}
