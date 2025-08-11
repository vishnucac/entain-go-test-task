package main

import (
	"entain-task/internal/db"
	"entain-task/internal/handlers"
	"entain-task/internal/models"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_NAME", "entain"),
		getEnv("DB_PORT", "5432"),
	)

	var err error
	db.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}

	if err := models.AutoMigrate(db.DB); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	seedUsers()

	r := gin.Default()

	r.POST("/user/:userId/transaction", handlers.PostTransaction)
	r.GET("/user/:userId/balance", handlers.GetBalance)

	port := getEnv("PORT", "8080")
	log.Printf("Listening on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func seedUsers() {
	for _, id := range []uint64{1, 2, 3} {
		var user models.User
		if err := db.DB.First(&user, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				u := models.User{ID: id, Balance: 0.00}
				if err := db.DB.Create(&u).Error; err != nil {
					log.Fatalf("failed to seed user %d: %v", id, err)
				}
			} else {
				log.Fatalf("failed to query user %d: %v", id, err)
			}
		}
	}
}
