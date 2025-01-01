package storage

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB
var err error

func InitDatabaseModels(dsn string, models []interface{}) {
	log.Printf("Configuring db connection for %d models ...", len(models))
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic("failed to connect to database")
	}

	if err := db.AutoMigrate(models...); err != nil {
		log.Fatalf("failed to migrate database: %v\n", err)
		return
	}

	models = append(models, &User{})
	models = append(models, &Subscription{})
	for _, model := range models {
		AddConfig(model)
	}
}

func DBMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Add a scoped DB instance to the context
        c.Set("db", db.Session(&gorm.Session{}))
        c.Next()
    }
}

func TransactionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start a transaction
		tx := db.Begin()
		if tx.Error != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			c.Abort()
			return
		}

		// Add transaction to context
		c.Set("tx", tx)

		// Process request
		c.Next()

		// Commit or rollback based on errors
		if len(c.Errors) > 0 {
			tx.Rollback()
		} else {
			if err := tx.Commit().Error; err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			}
		}
	}
}

func GetDbSpecial() (*gorm.DB) {
    return db
}

// GetTx retrieves the scoped *gorm.DB instance from the Gin context.
func GetDb(c *gin.Context) (*gorm.DB, error) {
	db, exists := c.Get("db")
	if !exists {
	    errorStr := "Database connection not found in context"
        c.JSON(http.StatusInternalServerError, gin.H{"error": errorStr})
		return nil, errors.New(errorStr)
	}

	// Assert that the value is a *gorm.DB
	gormDb, ok := db.(*gorm.DB)
	if !ok {
	    errorStr := "Context value is not a *gorm.DB instance"
        c.JSON(http.StatusInternalServerError, gin.H{"error": errorStr})
		return nil, errors.New(errorStr)
	}

	return gormDb, nil
}

// GetTx retrieves the transaction-bound *gorm.DB instance from the Gin context.
func GetTx(c *gin.Context) (*gorm.DB, error) {
	tx, exists := c.Get("tx")
	if !exists {
	    errorStr := "Transaction not found in context"
        c.JSON(http.StatusInternalServerError, gin.H{"error": errorStr})
		return nil, errors.New(errorStr)
	}

	// Assert that the value is a *gorm.DB
	gormTx, ok := tx.(*gorm.DB)
	if !ok {
	    errorStr := "Context value is not a *gorm.DB instance"
        c.JSON(http.StatusInternalServerError, gin.H{"error": errorStr})
		return nil, errors.New(errorStr)
	}

	return gormTx, nil
}