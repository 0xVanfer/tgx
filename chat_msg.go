package tgx

import (
	"fmt"

	"github.com/0xVanfer/tgx/internal/tgxerrors"
	"github.com/0xVanfer/tgx/internal/tgxutils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ChatMsg struct {
	Chat *Chat

	Msg *tgbotapi.Message

	Identifier  string
	Description string
}

// EditText edits the text of the message.
// If the msg is not found, it will send a new message with the given text.
func (msg *ChatMsg) EditText(text string) error {
	if msg == nil || msg.Chat == nil {
		return tgxerrors.ErrMsgNotFound
	}
	if msg.Msg == nil {
		newMsg := tgbotapi.NewMessage(msg.Chat.ChatID, text)
		newMsg.DisableWebPagePreview = msg.Chat.disableWebPagePreview
		if msg.Chat.ChatTopic > 0 {
			newMsg.ReplyToMessageID = msg.Chat.ChatTopic
		}
		_, err := msg.Chat.sendWithRetry(newMsg)
		return err
	}
	msgToEdit := tgbotapi.NewEditMessageText(msg.Msg.Chat.ID, msg.Msg.MessageID, text)
	msgToEdit.DisableWebPagePreview = msg.Chat.disableWebPagePreview

	_, err := msg.Chat.sendWithRetry(msgToEdit)
	return err
}

func (msg *ChatMsg) ReplaceWith(replacingMsg *tgbotapi.Message) error {
	msgToEdit := tgbotapi.NewEditMessageText(msg.Msg.Chat.ID, msg.Msg.MessageID, fmt.Sprintf("%v", replacingMsg.Text))
	msgToEdit.DisableWebPagePreview = msg.Chat.disableWebPagePreview
	msgToEdit.Entities = replacingMsg.Entities

	_, err := msg.Chat.sendWithRetry(msgToEdit)
	return err
}

// This function will only delete the tg msg, but the identifier will still be there.
// If you want to delete the identifier, use chat.DeleteMsgs(identifier) instead.
func (msg *ChatMsg) Delete() error {
	if msg == nil || msg.Msg == nil {
		return nil
	}
	if msg.Chat == nil {
		return tgxerrors.ErrMsgNotFound
	}
	msgToDelete := tgbotapi.NewDeleteMessage(msg.Msg.Chat.ID, msg.Msg.MessageID)
	// Allow msg not found.
	err := tgxutils.Retry(func() error {
		_, e := msg.Chat.Bot.Send(msgToDelete)
		if e.Error() == "Bad Request: message to delete not found" {
			return nil
		}
		return e
	}, msg.Chat.retry, msg.Chat.retryInterval)

	return err
}
