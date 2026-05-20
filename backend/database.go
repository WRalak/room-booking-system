package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/go-redis/redis/v8"
	"backend/models"
)

func initDB() (*gorm.DB, error) {
	dsn := config.DBURL
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL not set in environment")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connected successfully")

	// Run migrations
	err = db.AutoMigrate(
		&models.User{},
		&models.Room{},
		&models.Booking{},
		&models.Payment{},
		&models.Notification{},
		&models.Review{},
		&models.RecurringBooking{},
		&models.Invoice{},
		&models.AuditLog{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Migrations completed successfully")
	return db, nil
}

func initRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	return client
}
