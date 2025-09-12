package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)


func (bot *InteractiveChatbot) handleCommand(command string) bool {
	validCmd := false
    parts := strings.Fields(strings.ToLower(command))
	cmd := parts[0]
	switch cmd {
	case "help":
		bot.printWelcome()
		validCmd = true
	case "exit", "quit":
		bot.exitGracefully()
		validCmd = true
	case "new":
		name := "Conversation " + time.Now().Format("15:04")
		if len(parts) > 1 {
			name = strings.Join(parts[1:], " ")
		}
		id := fmt.Sprintf("conv_%d", time.Now().Unix())

		bot.conversations[id] = NewSmartConversation(id, name, bot.config.Model, bot.config.SystemPrompt, bot.config.MaxTokens)
		bot.conversation = bot.conversations[id]
		bot.currentID = id

		successColor.Printf("✨ Created new conversation: %s (%s)\n", name, id)
		validCmd = true

	case "list":
		systemColor.Println("📃 All Conversations:")
	for id, conv := range bot.conversations {
			current := ""
			if id == bot.currentID {
				current = " (current)"
			}
			fmt.Printf("	📝 %s%s: %s - %d messages\n", id, conv.Name, current, len(conv.Messages))
		}

		bot.listSavedConversations()

		validCmd = true

	case "clear":
		bot.conversation.Messages = make([]ChatMessage, 0)
		validCmd = true
	case "debug":
		bot.showDebugInfo()
		validCmd = true
	case "stats":
		bot.showStats()
		validCmd = true
	case "model":
		systemColor.Printf("model: %v\n", bot.config.Model)
		validCmd = true
	case "save":
		if len(parts) != 2{
			errorColor.Println("❌ Save failed: usage `save file-name`")
		}
		fileName := parts[1]
		if err := bot.saveConversation(fileName); err != nil {
			errorColor.Printf("❌ Save failed: %v\n", err)
		} else {
			successColor.Printf("💾 Saved as: %s.json\n", fileName)
		}
		validCmd = true
	case "stream":
		bot.config.StreamMode = !bot.config.StreamMode
		if bot.config.StreamMode {
			systemColor.Println("✨ Streaming mode enabled")
		} else {
			systemColor.Println("📦 Batch mode enabled")
		}
		validCmd = true
	default:
		validCmd = false
	}

	return validCmd
}

func (bot *InteractiveChatbot) exitGracefully() {

	systemColor.Println("👋 bye it was nice chatting!")
	bot.showStats()
	os.Exit(0)
}

func (bot *InteractiveChatbot) printWelcome() {
	fmt.Println()
	successColor.Println("🤖 Taconite - an interactive ai chat bot")
	systemColor.Printf("Model: %s\n", bot.config.Model)
	systemColor.Println("Commands:")
	fmt.Println("	help			- Show this help message")
	fmt.Println("	quit/exit		- Exit the chatbot")
	fmt.Println("	new <name>		- Create new conversation")
	fmt.Println("	clear			- Clear conversation history")
	fmt.Println("	debug			- Show debug information")
	fmt.Println("	stats			- Show conversation statistics")
	fmt.Println("	model			- Show/change current model")
	fmt.Println("	save [name]		- Save current conversation")
	fmt.Println()
	systemColor.Println("💡 Tip: Just type your message to chat!")
}

func (bot *InteractiveChatbot) showDebugInfo() {
	debugColor.Println("🔍 Debug Information:")
	debugColor.Printf("	Model: %s\n", bot.config.Model)
	debugColor.Printf("	Messsages in conversation: %d\n", len(bot.conversation.Messages))
	debugColor.Printf("	Estimated tokens: %d/%d\n", bot.conversation.TokenCount, bot.conversation.MaxTokens)
	debugColor.Printf("	Context usage: %.1f%%\n",
		float64(bot.conversation.TokenCount)/float64(bot.conversation.MaxTokens))
	debugColor.Println("	Recent messages:")
	start := max(len(bot.conversation.Messages) - 3, 0)

	for i := start; i < len(bot.conversation.Messages); i++ {
		msg := bot.conversation.Messages[i]
		debugColor.Printf("	%d. [%s] %.60s...\n", i+1, msg.Role, msg.Content)
	}
}

func (bot *InteractiveChatbot) showStats() {
	msgs := bot.conversation.Messages
	userMsgs := 0
	aiMsgs := 0
	systemMsgs := 0

	for _, msg := range msgs {
		switch msg.Role {
		case "user":
			userMsgs++
		case "assistant":
			aiMsgs++
		case "system":
			systemMsgs++
		}
	}

	systemColor.Println("📊 Conversation statistics:")
	fmt.Printf("	Total messages: %d\n", len(msgs))
	fmt.Printf("	User messages: %d\n", userMsgs)
	fmt.Printf("	AI responses: %d\n", aiMsgs)
	fmt.Printf("	System messags: %d\n", aiMsgs)
	fmt.Printf("	Estimated tokens: %d/%d\n", bot.conversation.TokenCount, bot.conversation.MaxTokens)
	fmt.Printf("	Context usage: %.1f%%\n",
		float64(bot.conversation.TokenCount)/float64(bot.conversation.MaxTokens))
}

func (bot *InteractiveChatbot) saveConversation(filename string) error {

	savedConvo := SavedConversation{
		Meta: ConversationMeta{
			ID:				bot.conversation.ID,
			Name:			bot.conversation.Name,
			Created: 		bot.conversation.Created,
			LastUsed: 		bot.conversation.LastUsed,
			MessageCount: len(bot.conversation.Messages),
		},
		Messages: bot.conversation.Messages,
		Config: bot.config,
	}

	filePath := filepath.Join(bot.saveDir, filename+".json")
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	return encoder.Encode(savedConvo)
}

func (bot *InteractiveChatbot) listSavedConversations() {
	files, err := filepath.Glob(filepath.Join(bot.saveDir, "*.json"))
	if err != nil || len(files) == 0 {
		return	
	}

	systemColor.Printf(" 💾 Saved Conversations:")
	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".json")

		// Try and read in metadata
		//BUG: missing function
		if savedConv, err := bot.readConversationMeta(file); err == nil {
           fmt.Printf("  💾 %s: %s (%d messages, %s)\n", 
                name, savedConv.Meta.Name, savedConv.Meta.MessageCount, 
                savedConv.Meta.LastUsed.Format("Jan 2 15:04"))	
		} else {
			fmt.Printf(" 💾 %s\n", name)
		}
	}
}
