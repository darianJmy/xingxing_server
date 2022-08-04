package types

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
