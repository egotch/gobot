package database

import (
	"fmt"
	"log"
	"os"
	
	"ai-chatbot-web/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	DB *gorm.DB
}

// NewDatabase initializes the database connection based on environment variables.
func NewDatabase() (*Database, error) {
	dbType := os.Getenv("DB_TYPE")

	if dbType == "" {
		dbType = "sqlite"
	}

	var db *gorm.DB
	var err error

	switch dbType {
	case "sqlite":
		dbPath := os.Getenv("DB_PATH")
		if dbPath == "" {
			dbPath = "./data.chatbot.db"
		}

		// ensure dirs exist
		if err := os.MkdirAll("./data", 0755); err != nil {
			return nil, fmt.Errorf("failed to create data directory: %v", err)
		}

		db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})

	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_PORT"),
			)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	default:
		return nil, fmt.Errorf("Unsupported DB_TYPE: %s", dbType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %v", err)
	}

	// auto migrate the schema
	if err := db.AutoMigrate(&models.User{}, &models.Conversation{}, &models.Message{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	log.Printf("✅ Database connected using %s", dbType)

	return &Database{DB: db}, nil

}

// Close closes the database connection.
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB from gorm DB: %v", err)
	}
	return sqlDB.Close()
}

// Ping checks the database connection.
func (d *Database) Ping() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB from gorm DB: %v", err)
	}
	return sqlDB.Ping()
}	
