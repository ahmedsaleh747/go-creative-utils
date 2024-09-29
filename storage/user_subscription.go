package storage

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Subscription struct {
	ID        uint      `json:"id" gorm:"primaryKey" extras:"hidden"`
	Endpoint  string    `json:"endpoint"`
	Auth      string    `json:"auth"`
	P256dh    string    `json:"p256dh"`
	UserId    *uint     `json:"user_id,string,omitempty" extras:"hidden"`
	User      User      `gorm:"foreignKey:user_id"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime" extras:"hidden"`
}

func (*Subscription) TableName() string {
	return "subscriptions"
}

func (*Subscription) GetTitle() string {
	return "Subscriptions Management"
}

func (*Subscription) GetApiUrl() string {
	return "/api/subscription"
}

func GetSubscriptionList(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"error": "Operation not supported!"})
}

func GetSubscription(c *gin.Context) {
	GetRecord(c, &Subscription{})
}

func CreateSubscription(c *gin.Context) {
	CreateRecord(c, &Subscription{})
}

func UpdateSubscription(c *gin.Context) {
	UpdateRecord(c, &Subscription{})
}

func DeleteSubscription(c *gin.Context) {
	DeleteRecord(c, &Subscription{})
}
