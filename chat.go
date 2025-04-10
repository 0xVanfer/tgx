package tgx

import (
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Chat struct {
	Bot *tgbotapi.BotAPI

	ChatID      int64
	ChatTopic   int
	Identifier  string
	Description string

	disableWebPagePreview bool
	retry                 int
	retryInterval         time.Duration

	// map[identifier] = msg info || []msg info
	managedMsgs sync.Map
}

// Send text message to the chat.
// If text is too long, it will be split into multiple messages.
// The maximum length of a single message is 4096 characters.
func (chat *Chat) SendTextMsg(text string) (msgsSent []*tgbotapi.Message, err error) {
	var spiltText []string
	for len(text) > 4096 {
		spiltText = append(spiltText, text[:4096])
		text = text[4096:]
	}
	if text != "" {
		spiltText = append(spiltText, text)
	}

	for _, t := range spiltText {
		msg, e := chat.sendTextMsg(t, nil)
		if e != nil {
			return nil, e
		}
		msgsSent = append(msgsSent, msg)
	}
	return
}

// Send text message with entities.
// The entities SHOULD be defined in the MsgComponent struct.
func (chat *Chat) SendTextMsgByComponents(components ...[]MsgComponent) (msgsSent []*tgbotapi.Message, err error) {
	for _, component := range components {
		text, entities := CompileMsgComponents(component...)
		if len(text) == 0 {
			msgsSent = append(msgsSent, nil)
			continue
		}
		if len(text) > 4096 {
			return nil, ErrTextTooLong
		}
		if len(entities) > 100 {
			return nil, ErrTooManyEntities
		}
		msgSent, err := chat.sendTextMsg(text, entities)
		if err != nil {
			return nil, err
		}
		msgsSent = append(msgsSent, msgSent)
	}
	return
}

// Send a photo to the chat.
// If sending a local file, photoPath should be the path to the file.
// If sending a online file, photoPath should be the URL to the file.
func (chat *Chat) SendPhoto(photoPath string, isLocal bool) (msgSent *tgbotapi.Message, err error) {
	var photo tgbotapi.RequestFileData
	if isLocal {
		photo = tgbotapi.FilePath(photoPath)
	} else {
		photo = tgbotapi.FileURL(photoPath)
	}

	msg := tgbotapi.NewPhoto(chat.ChatID, photo)
	if chat.ChatTopic != 0 {
		msg.ReplyToMessageID = chat.ChatTopic
	}
	return chat.sendWithRetry(msg)
}

// ========== Msg Related ==========

// An identifier can represent a single message or a list of messages.
// Use RegisterMsg() to register a single message.
// Use RegisterMsgs() to register a list of messages. (share the same identifier and description)
func (chat *Chat) RegisterMsg(msg *tgbotapi.Message, identifier string, description string) (*ChatMsg, error) {
	_, loaded := chat.managedMsgs.Load(identifier)
	if loaded {
		return nil, ErrIdentifierAlreadyExists
	}
	newMsg := &ChatMsg{
		chat:        chat,
		Msg:         msg,
		Identifier:  identifier,
		Description: description,
	}
	chat.managedMsgs.Store(identifier, newMsg)
	return newMsg, nil
}

// An identifier can represent a single message or a list of messages.
// Use RegisterMsg() to register a single message.
// Use RegisterMsgs() to register a list of messages. (share the same identifier and description)
func (chat *Chat) RegisterMsgs(msgs []*tgbotapi.Message, identifier string, description string) ([]*ChatMsg, error) {
	_, loaded := chat.managedMsgs.Load(identifier)
	if loaded {
		return nil, ErrIdentifierAlreadyExists
	}
	newMsgs := make([]*ChatMsg, 0)
	for _, msg := range msgs {
		newMsg := &ChatMsg{
			chat:        chat,
			Msg:         msg,
			Identifier:  identifier,
			Description: description,
		}
		newMsgs = append(newMsgs, newMsg)
	}

	chat.managedMsgs.Store(identifier, newMsgs)
	return newMsgs, nil
}

// GetMsg returns the message with the given identifier.
// If you can make sure that the identifier is a single message, you can use this function.
// If you are not sure, use GetMsgs() instead.
func (chat *Chat) GetMsg(identifier string) (*ChatMsg, error) {
	msgI, exist := chat.managedMsgs.Load(identifier)
	if !exist {
		return nil, ErrIdentifierNotFound
	}
	msg, ok := msgI.(*ChatMsg)
	if ok {
		return msg, nil
	}
	msgs, ok := msgI.([]*ChatMsg)
	if ok && len(msgs) > 0 {
		return msgs[0], nil
	}
	return nil, ErrIdentifierNotFound
}

// GetMsgs returns the messages with the given identifier.
// If the identifier is a single message, it will be returned as a slice with a single element.
// So you can always use this function to get the messages.
func (chat *Chat) GetMsgs(identifier string) ([]*ChatMsg, error) {
	msgI, exist := chat.managedMsgs.Load(identifier)
	if !exist {
		return nil, ErrIdentifierNotFound
	}
	msgs, ok := msgI.([]*ChatMsg)
	if ok {
		return msgs, nil
	}
	msg, ok := msgI.(*ChatMsg)
	if ok {
		return []*ChatMsg{msg}, nil
	}
	return nil, ErrIdentifierNotFound
}

// DeleteMsgs deletes the tg messages with the given identifier, and free the identidier.
func (chat *Chat) DeleteMsgs(identifier string) error {
	msgs, err := chat.GetMsgs(identifier)
	if err != nil {
		return err
	}
	for _, msg := range msgs {
		_ = msg.Delete()
	}
	chat.managedMsgs.Delete(identifier)
	return nil
}

// ========== Setters ==========

func (chat *Chat) SetRetry(retry int)                      { chat.retry = retry }
func (chat *Chat) SetRetryInterval(interval time.Duration) { chat.retryInterval = interval }
func (chat *Chat) SetDisableWebPagePreview(disable bool)   { chat.disableWebPagePreview = disable }

// ========== Internal ==========

func (chat *Chat) sendWithRetry(msg tgbotapi.Chattable) (msgSent *tgbotapi.Message, err error) {
	err = retry(func() error {
		newMsg, e := chat.Bot.Send(msg)
		msgSent = &newMsg
		return e
	}, chat.retry, chat.retryInterval)
	return
}

// Internal function.
// Chat must be valid; text length must < 4096; entities length must < 100.
func (chat *Chat) sendTextMsg(text string, entities []tgbotapi.MessageEntity) (msgSent *tgbotapi.Message, err error) {
	msg := tgbotapi.NewMessage(chat.ChatID, text)
	msg.DisableWebPagePreview = chat.disableWebPagePreview
	if chat.ChatTopic != 0 {
		msg.ReplyToMessageID = chat.ChatTopic
	}
	msg.Entities = entities
	return chat.sendWithRetry(msg)
}
