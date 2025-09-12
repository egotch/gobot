package main

import (
	"ai-chatbot/ai"
)

func main() {
    // Create conversation with small token limit to demonstrate trimming
    systemPrompt := "You are a helpful assistant. Your name is Taconite. Be informative but concise and friendly."
	model := "llama3.1:8b"

	chatbot := ai.NewInteractiveChatBot(model, systemPrompt)
	chatbot.Run()
}
