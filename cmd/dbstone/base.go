package dbstone

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var mysql *Mysqls
var DB *gorm.DB

type Mysqls struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	DBName   string `yaml:"name"`
}

func init() {
	dbConnection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local&timeout=30s",
		"root",
		"root",
		"127.0.0.1",
		"3306",
		"users")
	DB, err := gorm.Open("mysql", dbConnection)
	if err != nil {
		panic(err)
	}
	fmt.Println("connection succeeded")
	DB.DB().SetMaxIdleConns(10)
	DB.DB().SetMaxOpenConns(100)
	DB.SingularTable(true)
	fmt.Println(DB)
}
