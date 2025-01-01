package storage

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Identity interface {
	GetId() uint
	GetName() string
	GetRole() string
}

type User struct {
	ID       uint   `json:"id" gorm:"primaryKey" extras:"hidden"`
	Name     string `json:"username" gorm:"unique"`
	Password string `json:"password" extras:"sensitive"`
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

func (u *User) GetId() uint {
	return u.ID
}

func (u *User) GetName() string {
	return u.Name
}

func (u *User) GetRole() string {
	return "Unknown" //unknown role, should be overridden by the child structs
}

func GetUserUsingNameAndPassword(c *gin.Context, user Identity) bool {
	var requestUser User
	if err := c.BindJSON(&requestUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return false
	}
    db, err := GetDb(c)
	if err != nil {
		return false
	}
	if err := db.Where("name ILIKE ? and password = ?", requestUser.Name, requestUser.Password).
		First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return false
	}
	return true
}

// PostLoad called by reflection
func (record *User) PostLoad() {
	record.Password = "****"
}
