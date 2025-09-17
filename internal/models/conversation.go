// internal/models/conversation.go
package models

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type Conversation struct {
    ID          string    `json:"id" gorm:"primaryKey"`
    Name        string    `json:"name"`
    UserID      string    `json:"user_id"`
    Model       string    `json:"model"`
    SystemPrompt string   `json:"system_prompt"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    Messages    []Message `json:"messages,omitempty" gorm:"foreignKey:ConversationID"`
}

type Message struct {
    ID             string    `json:"id" gorm:"primaryKey"`
    ConversationID string    `json:"conversation_id"`
    Role           string    `json:"role"` // system, user, assistant
    Content        string    `json:"content"`
    TokenCount     int       `json:"token_count"`
    CreatedAt      time.Time `json:"created_at"`
}

type User struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    Username  string    `json:"username" gorm:"unique"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate hook for generating UUIDs
func (c *Conversation) BeforeCreate(tx *gorm.DB) error {
    if c.ID == "" {
        c.ID = uuid.New().String()
    }
    return nil
}

func (m *Message) BeforeCreate(tx *gorm.DB) error {
    if m.ID == "" {
        m.ID = uuid.New().String()
    }
    return nil
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.ID == "" {
        u.ID = uuid.New().String()
    }
    return nil
}
