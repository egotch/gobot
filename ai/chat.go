package ai

import (
	"ai-chatbot/progress"
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

type ChatMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
	Time	time.Time `json:"time"`
}

type Config struct {
	Model		string `json:"model"`
	SystemPrompt	string `json:"system_prompt"`
	MaxTokens		int		`json:"max_tokens"`
	StreamMode		bool	`json:"stream_mode"`
	SaveDir			string	`json:"save_dir"`
}

type ConversationMeta struct {
	ID				string		`json:"id"`
	Name			string		`json:"name"`
	Created			time.Time	`json:"created"`
	LastUsed		time.Time	`json:"last_used"`
	MessageCount	int			`json:"message_count"`
}

type SavedConversation struct {
	Meta		ConversationMeta	`json:"meta"`
	Messages	[]ChatMessage		`json:"messages"`
	Config		Config				`json:"config"`
}

type InteractiveChatbot struct {
	config			Config	
	conversation	*SmartConversation
	conversations	map[string]*SmartConversation
	currentID		string
	saveDir			string
}

// SmartConversation manages conversation with token limits
type SmartConversation struct {
	ID			string
	Name		string
    Messages   []ChatMessage
    Model      string
    MaxTokens  int // Maximum tokens to keep in context
    TokenCount int // Current estimated token count
	Created		time.Time
	LastUsed	time.Time
}

func NewInteractiveChatBot(model string, systemPrompt string) *InteractiveChatbot {

	config := Config{
		Model: 			"llama3.1:8b",
		SystemPrompt:  "You are a helpful assistant. Your name is Taconite. Be informative but concise and friendly.",
		MaxTokens: 		4000,
		StreamMode: 	true,
		SaveDir:		"./conversations",
	}

	// Create the save dir
	os.MkdirAll(config.SaveDir, 0755)

	bot := &InteractiveChatbot{
		config: config,
		conversations: make(map[string]*SmartConversation),
		currentID: "default",
		saveDir: config.SaveDir,
	}

	bot.conversations["default"] = NewSmartConversation("default", "Default Chat", config.Model, config.SystemPrompt, config.MaxTokens)
	bot.conversation = bot.conversations["default"]

	return bot
}

func NewSmartConversation(id, name, model, systemPrompt string, maxTokens int) *SmartConversation {
    conv := &SmartConversation{
		ID:			id,
		Name:		name,
        Messages:	make([]ChatMessage, 0),
        Model:     	model,
        MaxTokens: 	maxTokens,
		Created: 	time.Now(),
		LastUsed: 	time.Now(),
    }
    
    if systemPrompt != "" {
        conv.AddMessage("system", systemPrompt)
    }
    
    return conv
}

func (c *SmartConversation) AddMessage(role, content string) {
	message := ChatMessage{
		Role: role,
		Content: content,
		Time: time.Now(),
	}
    c.Messages = append(c.Messages, message)
    c.TokenCount += c.estimateTokens(content)
	c.LastUsed = message.Time
    c.trimToFitContext()
}

// Rough token estimation (4 chars ≈ 1 token)
func (c *SmartConversation) estimateTokens(text string) int {
    return len(text) / 4
}

// Trim old messages to stay within token limit
func (c *SmartConversation) trimToFitContext() {
    if c.TokenCount <= c.MaxTokens {
        return // No trimming needed
    }
    
    // Always keep system prompt if it exists
    systemMsgExists := len(c.Messages) > 0 && c.Messages[0].Role == "system"
    startIdx := 0
    if systemMsgExists {
        startIdx = 1
    }
    
    // Remove old messages until we're under limit
    for c.TokenCount > c.MaxTokens && len(c.Messages) > startIdx+2 {
        // Remove the oldest non-system message
        removedMsg := c.Messages[startIdx]
        c.Messages = append(c.Messages[:startIdx], c.Messages[startIdx+1:]...)
        c.TokenCount -= c.estimateTokens(removedMsg.Content)
        
        fmt.Printf("🗑️  Trimmed old message: [%s] %.30s...\n", removedMsg.Role, removedMsg.Content)
    }
}

func (bot *InteractiveChatbot) sendMessage(userInput string) {
	// Add user message to conversation
	bot.conversation.AddMessage("user", userInput)

	// Show thinking indicator while model is spinning
	fmt.Print("🤖 ")
	aiColor.Print("AI: ")

	var err error

	// Send message to Ollama and await response
	if bot.config.StreamMode{
		_, err = bot.conversation.SendToOllamaStream()
	} else {
		ctx, cancel := context.WithCancel(context.Background())
		// go progress.ShowSpinnerProgress(ctx)
		go progress.ShowColorfulProgress(ctx)

		response, batchErr := bot.conversation.SendToOllamaBatch()

		// stop the spinner as soon as response comes back
		cancel()
		err = batchErr

		if err == nil {
			fmt.Print("\r🤖 ") // clear thinking indicator with carriage return
			aiColor.Print("AI: ")
			fmt.Println(response)
		}
	}

	if err != nil {
		errorColor.Printf("❌ Error: %v\n", err)

		// Give some high level troubleshooting
		if strings.Contains(err.Error(), "connection refused") {
			systemColor.Println("💡 Tip: Make sure Ollama is running with `ollama serve`")
		} else if strings.Contains(err.Error(), "timeout") {
			systemColor.Printf("💡 Tip: the model might be processing.  Try a shorter message.")
		}
		return
	}

}

// Run is the main interactive loop runner
// for the chat bot
func (bot *InteractiveChatbot) Run() {

	bot.printWelcome()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println()
		userColor.Print("👤 You: ")

		// read the input
		if !scanner.Scan() {
			break
		}

		userInput := strings.TrimSpace(scanner.Text())

		// skip blank input
		if userInput == "" {
			continue
		}

		// check if it's a command
		if bot.handleCommand(userInput) {
			continue
		}

		bot.sendMessage(userInput)

	}
}
