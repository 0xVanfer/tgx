package tgx

import (
	"fmt"

	"github.com/0xVanfer/tgx/internal/tgxutils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ChatMsg struct {
	chat *Chat

	Msg *tgbotapi.Message

	Identifier  string
	Description string
}

func (msg *ChatMsg) EditText(text string) error {
	msgToEdit := tgbotapi.NewEditMessageText(msg.Msg.Chat.ID, msg.Msg.MessageID, text)
	msgToEdit.DisableWebPagePreview = msg.chat.disableWebPagePreview

	_, err := msg.chat.sendWithRetry(msgToEdit)
	return err
}

func (msg *ChatMsg) ReplaceWith(replacingMsg *tgbotapi.Message) error {
	msgToEdit := tgbotapi.NewEditMessageText(msg.Msg.Chat.ID, msg.Msg.MessageID, fmt.Sprintf("%v", replacingMsg.Text))
	msgToEdit.DisableWebPagePreview = msg.chat.disableWebPagePreview
	msgToEdit.Entities = replacingMsg.Entities

	_, err := msg.chat.sendWithRetry(msgToEdit)
	return err
}

// This function will only delete the tg msg, but the identifier will still be there.
// If you want to delete the identifier, use chat.DeleteMsgs(identifier) instead.
func (msg *ChatMsg) Delete() error {
	msgToDelete := tgbotapi.NewDeleteMessage(msg.Msg.Chat.ID, msg.Msg.MessageID)
	// Allow msg not found.
	err := tgxutils.Retry(func() error {
		_, e := msg.chat.Bot.Send(msgToDelete)
		if e.Error() == "Bad Request: message to delete not found" {
			return nil
		}
		return e
	}, msg.chat.retry, msg.chat.retryInterval)

	return err
}
