package test

import (
	"fmt"
	"testing"
)

func TestGeneralPrint(t *testing.T) {
	allBotsInfo := wrapper.GetAllRegisteredBots()
	for token, info := range allBotsInfo {
		fmt.Println("\n\nBot Token (key):     ", token)
		fmt.Println("Bot Token (info.Bot):", info.Bot.Token)
		for _, chat := range info.Chats {
			fmt.Println("\nChat Info:")
			fmt.Println("  Chat ID:", chat.ChatID)
			fmt.Println("  Chat Topic:", chat.ChatTopic)
			fmt.Println("  Identifier:", chat.Identifier)
			fmt.Println("  Description:", chat.Description)
		}
	}
}
