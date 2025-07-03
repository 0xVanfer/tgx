package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/0xVanfer/tgx"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestMonitor(t *testing.T) {
	monitorTopicChat.RegisterHandleMsg("aaa", func(msg *tgbotapi.Message) (err error) {
		text := msg.Text
		if strings.Contains(text, "aaa") {
			_, err := monitorTopicChat.SendTextMsg(nil, "You sent a message containing 'aaa'.")
			return err
		}
		return nil
	})

	monitorTopicChat.RegisterHandleMsg("xxx", func(msg *tgbotapi.Message) (err error) {
		text := msg.Text
		if strings.Contains(text, "xxx") {
			_, err := monitorTopicChat.SendTextMsg(nil, "You sent a message containing 'xxx'.")
			return err
		}
		return nil
	})

	entireChat.RegisterHandleMsg("xxx", func(msg *tgbotapi.Message) (err error) {
		text := msg.Text
		if strings.Contains(text, "xxx") {
			components := []tgx.MsgComponent{
				{Text: "You sent a message containing "},
				{Text: "xxx", EntitiyType: "bold"},
				{Text: "."},
				{Text: fmt.Sprintf(" @%s", msg.From.UserName), EntitiyType: "mention"},
			}
			_, err := entireChat.SendTextMsgByComponents(nil, components)
			return err
		}
		return nil
	})

	monitorTopicChat.RegisterHandleCommand("timestamp", func(msg *tgbotapi.Message) (err error) {
		_, err = monitorTopicChat.SendTextMsg(nil, fmt.Sprintf("Current timestamp: %d", time.Now().Unix()))
		return
	})

	monitorEverywhere.RegisterHandleCommand("datetime", func(msg *tgbotapi.Message) (err error) {
		overrideInfo := monitorEverywhere.GetOverrideInfoFromMsg(msg)
		_, err = monitorEverywhere.SendTextMsg(overrideInfo, time.Now().UTC().Format(time.DateTime))
		return
	})

	wrapper.Monitor()

	fmt.Println("\n\nMonitor started. Please send messages containing 'aaa' or 'xxx' to the chat.")
	select {}
}
