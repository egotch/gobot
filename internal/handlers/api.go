package handlers

import (
	"ai-chatbot-web/internal/database"
	"ai-chatbot-web/internal/models"
	"ai-chatbot-web/internal/services"
	"net/http"

	 "github.com/gin-gonic/gin"
)

type APIHandler struct {
	aiClient *services.AIClient
	db       *database.Database
}

// Handler interface defines methods for handling API requests.
func NewAPIHandler(db *database.Database, aiClient *services.AIClient) *APIHandler {
	return &APIHandler{
		db: 	 db,
		aiClient: aiClient,
	}
}

// HealthCheck handles the health check endpoint.
func (h *APIHandler) HealthCheck(c *gin.Context) {
    if err := h.db.Ping(); err != nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "error",
			"message": "Database connection failed",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "API is healthy",
		"model":   h.aiClient.GetModel(),
	})
}

// GetConversations retrieves conversations for a given user.
func (h *APIHandler) GetConversations(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		userID = "default_user"
	}

	var conversations []models.Conversation
	if err := h.db.DB.Where("user_id = ?", userID).Find(&conversations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to retrieve conversations",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"conversations": conversations,
		"count":         len(conversations),
	})
}

// Create a new conversation.
func (h *APIHandler) CreateConversation(c *gin.Context) {
	var req struct {
		Name   string `json:"name"`
		UserID string `json:"user_id"`
		SystemPrompt string `json:"system_prompt"`
	}

	// Bind JSON request body to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request payload",
		})
		return
	}

	if req.UserID == "" {
		req.UserID = "default_user"
	}

	if req.SystemPrompt == "" {
		req.SystemPrompt = "You are a helpful assistant."
	}

	conversation := models.Conversation{
		Name:         req.Name,
		UserID:       req.UserID,
		SystemPrompt: req.SystemPrompt,
		Model:        h.aiClient.GetModel(),
	}

	// Save the new conversation to the database
	if err := h.db.DB.Create(&conversation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to create conversation",
		})
		return
	}

	// add a system message to the conversation
	systemMessage := models.Message{
		ConversationID: conversation.ID,
		Role:           "system",
		Content:        req.SystemPrompt,
		TokenCount:    h.aiClient.EstimateTokens(req.SystemPrompt),
	}
	// Save the system message to the database
	if err := h.db.DB.Create(&systemMessage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to create system message",
		})
		return
	}
	
	// Return the created conversation
	c.JSON(http.StatusCreated, gin.H{
		"status":       "success",
		"conversation": conversation,
		"message":      "Conversation created successfully",
	})
}

// Get conversation with messages.
func (h *APIHandler) GetConversation(c *gin.Context) {
	conversationID := c.Param("id")

	var conversation models.Conversation
	if err := h.db.DB.Preload("Messages").First(&conversation, "id = ?", conversationID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "Conversation not found",
		})
		return
	}

	// Return the conversation
	c.JSON(http.StatusOK, gin.H{
		"status":       "success",
		"conversation": conversation,
	})
}

// Send a message in a conversation and get AI response.
func (h *APIHandler) SendMessage(c *gin.Context) {
    conversationID := c.Param("id")
    
    var req struct {
        Content string `json:"content" binding:"required"`
        Role    string `json:"role"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "invalid request format",
        })
        return
    }
    
    if req.Role == "" {
        req.Role = "user"
    }
    
    // Verify conversation exists
    var conversation models.Conversation
    if err := h.db.DB.First(&conversation, "id = ?", conversationID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "error": "conversation not found",
        })
        return
    }
    
    // Create user message
    userMessage := models.Message{
        ConversationID: conversationID,
        Role:           req.Role,
        Content:        req.Content,
        TokenCount:     h.aiClient.EstimateTokens(req.Content),
    }
    
    if err := h.db.DB.Create(&userMessage).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "failed to save message",
        })
        return
    }
    
    // Get all messages for context
    var messages []models.Message
    if err := h.db.DB.Where("conversation_id = ?", conversationID).Order("created_at ASC").Find(&messages).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "failed to load conversation history",
        })
        return
    }
    
    // Send to AI
    aiResponse, err := h.aiClient.SendMessage(messages)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "AI request failed: " + err.Error(),
        })
        return
    }
    
    // Save AI response
    assistantMessage := models.Message{
        ConversationID: conversationID,
        Role:           "assistant",
        Content:        aiResponse,
        TokenCount:     h.aiClient.EstimateTokens(aiResponse),
    }
    
    if err := h.db.DB.Create(&assistantMessage).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "failed to save AI response",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "user_message":      userMessage,
        "assistant_message": assistantMessage,
        "success":           true,
    })
}

// Delete a conversation
func (h *APIHandler) DeleteConversation(c *gin.Context) {
    conversationID := c.Param("id")
    
    // Delete messages first (due to foreign key constraint)
    if err := h.db.DB.Where("conversation_id = ?", conversationID).Delete(&models.Message{}).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "failed to delete messages",
        })
        return
    }
    
    // Delete conversation
    if err := h.db.DB.Delete(&models.Conversation{}, "id = ?", conversationID).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "failed to delete conversation",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "message": "conversation deleted successfully",
    })
}
