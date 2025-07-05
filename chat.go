package tgx

import (
	"sync"
	"time"

	"github.com/0xVanfer/tgx/internal/tgxerrors"
	"github.com/0xVanfer/tgx/internal/tgxutils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Chat struct {
	Bot *tgbotapi.BotAPI

	ChatID      int64
	ChatTopic   int
	Identifier  string
	Description string

	// I don't like the web page preview, so I set it to true by default.
	// If you want to enable it, use SetDisableWebPagePreview() to set it
	disableWebPagePreview bool

	// Set retry times and interval.
	retry         int
	retryInterval time.Duration

	// map[identifier] = msg info || []msg info
	managedMsgs sync.Map

	// To handle commands.
	// map[command string] = func(msg *tgbotapi.Message) (err error)
	handleCommandFuncs map[string]func(msg *tgbotapi.Message) (err error)

	// To handle normal messages.
	// Each different logic can be registered with a different function.
	// To manage the functions, use a map to make registered functions easy to find.
	handleMsgFuncs map[string]func(msg *tgbotapi.Message) (err error)
}

type ChatAndTopic struct {
	ChatID    int64
	ChatTopic int
}

// Turn the msg into a ChatAndTopic struct.
func (chat *Chat) GetOverrideInfoFromMsg(msg *tgbotapi.Message) *ChatAndTopic {
	if msg == nil {
		return nil
	}
	if msg.ReplyToMessage != nil {
		return &ChatAndTopic{
			ChatID:    msg.Chat.ID,
			ChatTopic: msg.ReplyToMessage.MessageID,
		}
	}
	return &ChatAndTopic{
		ChatID:    msg.Chat.ID,
		ChatTopic: 0,
	}
}

// Send text message to the chat. If targetChatOverride is not nil, it will override the chat ID and topic.
//
// If text is too long, it will be split into multiple messages.
// The maximum length of a single message is 4096 characters.
func (chat *Chat) SendTextMsg(targetChatOverride *ChatAndTopic, text string) (msgsSent []*tgbotapi.Message, err error) {
	var spiltText []string
	for len(text) > 4096 {
		spiltText = append(spiltText, text[:4096])
		text = text[4096:]
	}
	if text != "" {
		spiltText = append(spiltText, text)
	}

	for _, t := range spiltText {
		msg, e := chat.sendTextMsg(targetChatOverride, t, nil)
		if e != nil {
			return nil, e
		}
		msgsSent = append(msgsSent, msg)
	}
	return
}

// Send text message with entities. If targetChatOverride is not nil, it will override the chat ID and topic.
//
// The entities should be defined in the MsgComponent struct. And total entities length must be no longer than 100.
func (chat *Chat) SendTextMsgByComponents(targetChatOverride *ChatAndTopic, components ...[]MsgComponent) (msgsSent []*tgbotapi.Message, err error) {
	for _, component := range components {
		text, entities := CompileMsgComponents(component...)
		if len(text) == 0 {
			msgsSent = append(msgsSent, nil)
			continue
		}
		if len(text) > 4096 {
			return nil, tgxerrors.ErrTextTooLong
		}
		if len(entities) > 100 {
			return nil, tgxerrors.ErrTooManyEntities
		}
		msgSent, err := chat.sendTextMsg(targetChatOverride, text, entities)
		if err != nil {
			return nil, err
		}
		msgsSent = append(msgsSent, msgSent)
	}
	return
}

// Send a photo to the chat. If targetChatOverride is not nil, it will override the chat ID and topic.
//
// If sending a local file, photoPath should be the path to the file.
// If sending a online file, photoPath should be the URL to the file.
func (chat *Chat) SendPhoto(targetChatOverride *ChatAndTopic, photoPath string, isLocal bool) (msgSent *tgbotapi.Message, err error) {
	var photo tgbotapi.RequestFileData
	if isLocal {
		photo = tgbotapi.FilePath(photoPath)
	} else {
		photo = tgbotapi.FileURL(photoPath)
	}

	chatID, topic := chat.decideChatAndTopic(targetChatOverride)

	msg := tgbotapi.NewPhoto(chatID, photo)
	if topic != 0 {
		msg.ReplyToMessageID = topic
	}
	return chat.sendWithRetry(msg)
}

func (chat *Chat) RegisterHandleCommand(command string, handleFunc func(msg *tgbotapi.Message) (err error)) {
	// Map must be initialized when the chat is created.
	chat.handleCommandFuncs[command] = handleFunc
}

func (chat *Chat) HandleCommand(msg *tgbotapi.Message) error {
	commandStr := msg.Command()
	funcx, exist := chat.handleCommandFuncs[commandStr]
	if !exist {
		return nil
	}
	return funcx(msg)
}

func (chat *Chat) RegisterHandleMsg(funcIdentifier string, handleFunc func(msg *tgbotapi.Message) (err error)) {
	// Map must be initialized when the chat is created.
	chat.handleMsgFuncs[funcIdentifier] = handleFunc
}

func (chat *Chat) HandleMsg(msg *tgbotapi.Message) map[string]error {
	errorRes := make(map[string]error)
	for identidier, funcx := range chat.handleMsgFuncs {
		if funcx == nil {
			continue
		}
		err := funcx(msg)
		if err != nil {
			errorRes[identidier] = err
		}
	}
	return errorRes
}

// ========== Msg Related ==========

// An identifier can represent a single message or a list of messages.
// Use RegisterMsg() to register a single message.
// Use RegisterMsgs() to register a list of messages. (share the same identifier and description)
func (chat *Chat) RegisterMsg(msg *tgbotapi.Message, identifier string, description string) (*ChatMsg, error) {
	_, loaded := chat.managedMsgs.Load(identifier)
	if loaded {
		return nil, tgxerrors.ErrIdentifierAlreadyExists
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
		return nil, tgxerrors.ErrIdentifierAlreadyExists
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
		return nil, tgxerrors.ErrIdentifierNotFound
	}
	msg, ok := msgI.(*ChatMsg)
	if ok {
		return msg, nil
	}
	msgs, ok := msgI.([]*ChatMsg)
	if ok && len(msgs) > 0 {
		return msgs[0], nil
	}
	return nil, tgxerrors.ErrIdentifierNotFound
}

// GetMsgs returns the messages with the given identifier.
// If the identifier is a single message, it will be returned as a slice with a single element.
// So you can always use this function to get the messages.
func (chat *Chat) GetMsgs(identifier string) ([]*ChatMsg, error) {
	msgI, exist := chat.managedMsgs.Load(identifier)
	if !exist {
		return nil, tgxerrors.ErrIdentifierNotFound
	}
	msgs, ok := msgI.([]*ChatMsg)
	if ok {
		return msgs, nil
	}
	msg, ok := msgI.(*ChatMsg)
	if ok {
		return []*ChatMsg{msg}, nil
	}
	return nil, tgxerrors.ErrIdentifierNotFound
}

func (chat *Chat) DeleteMsg(msg *tgbotapi.Message) error {
	if msg == nil || msg.Chat == nil {
		return tgxerrors.ErrMsgNotFound
	}
	deletingMsg := tgbotapi.NewDeleteMessage(msg.Chat.ID, msg.MessageID)
	_, err := chat.sendWithRetry(deletingMsg)
	if err != nil {
		return err
	}
	return nil
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

func (chat *Chat) DeleteMsgByID(chatID int64, msgID int) error {
	msg := tgbotapi.NewDeleteMessage(chatID, msgID)
	_, err := chat.sendWithRetry(msg)
	if err != nil {
		return err
	}
	return nil
}

// ========== Setters ==========

func (chat *Chat) SetRetry(retry int)                      { chat.retry = retry }
func (chat *Chat) SetRetryInterval(interval time.Duration) { chat.retryInterval = interval }
func (chat *Chat) SetDisableWebPagePreview(disable bool)   { chat.disableWebPagePreview = disable }

// ========== Internal ==========

func (chat *Chat) sendWithRetry(msg tgbotapi.Chattable) (msgSent *tgbotapi.Message, err error) {
	err = tgxutils.Retry(func() error {
		newMsg, e := chat.Bot.Send(msg)
		msgSent = &newMsg
		return e
	}, chat.retry, chat.retryInterval)
	return
}

// Internal function.
// Chat must be valid; text length must < 4096; entities length must < 100.
func (chat *Chat) sendTextMsg(targetChatOverride *ChatAndTopic, text string, entities []tgbotapi.MessageEntity) (msgSent *tgbotapi.Message, err error) {
	chatID, topic := chat.decideChatAndTopic(targetChatOverride)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.DisableWebPagePreview = chat.disableWebPagePreview
	if topic > 0 {
		msg.ReplyToMessageID = topic
	}
	msg.Entities = entities
	return chat.sendWithRetry(msg)
}

func (chat *Chat) decideChatAndTopic(targetChatOverride *ChatAndTopic) (chatID int64, topic int) {
	if targetChatOverride != nil {
		chatID = targetChatOverride.ChatID
		topic = max(targetChatOverride.ChatTopic, 0)
	} else {
		chatID = chat.ChatID
		topic = max(chat.ChatTopic, 0)
	}
	return
}
