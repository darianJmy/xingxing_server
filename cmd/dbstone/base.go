package dbstone

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var mysql *Mysqls
var DB *gorm.DB

type UserDB struct {
	dbstone *gorm.DB
}

type Mysqls struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	DBName   string `yaml:"name"`
}

func NewUserDB() *UserDB {
	return &UserDB{
		dbstone: DB,
	}
}

func (u *UserDB) CreateUser(user *User) (*User, error) {
	_, err := u.GetUser(user)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := u.dbstone.Create(&user).Error; err != nil {
			return nil, fmt.Errorf("创建用户失败")
		}
		result, err := u.GetUser(user)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	return nil, fmt.Errorf("用户已存在")
}
func (u *UserDB) GetUser(user *User) (*User, error) {
	var d User
	if err := u.dbstone.Where("mg_name = ?", user.MG_NAME).First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

//func (u *UserDB) List(ctx
//context.Context, name
//string) (*[]models.User, error) {
//var users []models.User
//if err := u.dbstone.Where("name = ?", name).Find(&users).Error; err != nil {
//return nil, err
//}
//return &users, nil
//}
//
//func (u *UserDB) Update(ctx
//context.Context, name
//string, age
//int) error{
//var user models.User
//return u.dbstone.Model(&user).Where("name = ?", name).Update("age", age).Error
//}
//
//func (u *UserDB) Delete(ctx
//context.Context, name
//string, age
//int) error{
//var user models.User
//return u.dbstone.Where("name = ? AND age = ?", name, age).Delete(&user).Error
//}

func init() {
	var err error
	var user User
	dbConnection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local&timeout=30s",
		"root",
		"root",
		"127.0.0.1",
		"3306",
		"gorm")
	DB, err = gorm.Open("mysql", dbConnection)
	if err != nil {
		panic(err)
	}
	fmt.Println("connection succeeded")
	DB.DB().SetMaxIdleConns(10)
	DB.DB().SetMaxOpenConns(100)
	DB.SingularTable(true)
	DB.AutoMigrate(&user)
	fmt.Println(DB)
}
