package storage

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID       uint   `json:"id" gorm:"primaryKey" extras:"hidden"`
	Name     string `json:"username" gorm:"unique"`
	Password string `json:"password" extras:"sensitive"`
	Role     string `json:"role" extras:"enum:Admin|Scraper"`
}

func (user *User) MarshalJSON() ([]byte, error) {
	type Alias User
	return json.Marshal(&struct {
		Password string `json:"password"` //this needs to be the same work as the field, but all lower case, to match utils
		*Alias
	}{
		Password: "****",
		Alias:    (*Alias)(user),
	})
}

func (*User) TableName() string {
	return "users"
}

func (*User) GetTitle() string {
	return "Users Management"
}

func (*User) GetApiUrl() string {
	return "/api/user"
}

func GetUserList(c *gin.Context) {
	var records []User
	GetRecords(c, &records)
}

func GetUser(c *gin.Context) {
	GetRecord(c, &User{})
}

func CreateUser(c *gin.Context) {
	CreateRecord(c, &User{})
}

func UpdateUser(c *gin.Context) {
	UpdateRecord(c, &User{})
}

func DeleteUser(c *gin.Context) {
	DeleteRecord(c, &User{})
}

func GetUserUsingNameAndPassword(c *gin.Context) (user User) {
	var requestUser User
	if err := c.BindJSON(&requestUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	if err := GetDb().Where("name ILIKE ? and password = ?", requestUser.Name, requestUser.Password).
		First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}
	return
}
