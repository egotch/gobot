package main

import (
	"log"
	"os"
	
	"ai-chatbot-web/internal/database"
	"ai-chatbot-web/internal/handlers"
	"ai-chatbot-web/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize database
	db, err := database.NewDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize AI client
	aiClient := services.NewAIClient()

	// Initialize handlers
	handler := handlers.NewAPIHandler(db, aiClient)

	// Set up Gin router
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// API routes
	api := router.Group("/api/v1")
	{
		api.GET("/health", handler.HealthCheck)
		api.GET("/conversations", handler.GetConversations)
		api.POST("/conversations", handler.CreateConversation)
		api.GET("/conversations/:id", handler.GetConversation)
		api.POST("/conversations/:id/messages", handler.SendMessage)
		api.DELETE("/conversations/:id", handler.DeleteConversation)
	}

	// Serve static files
	router.Static("/static", "./web/static")
	router.LoadHTMLFiles("./web/templates/index.html")

	// Root route to serve the main HTML file
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{
			"title": "AI Chatbot",
		})
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server starting on port %s", port)
	log.Printf("📊 Health check: http://localhost:%s/api/v1/health", port)
	log.Printf("🤖 AI Model: %s", aiClient.GetModel())

if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}

}
