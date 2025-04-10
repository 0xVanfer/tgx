package tgx

import (
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// The wrapper for telegram bot API to record the chats information.
//
// Use RegisterChat() to register a chat;
// Use GetChat() to get the chat information.
//
// More functions are avaliable under the return struct of GetChat().
type TgWrapper struct {
	chatsByIdentifier sync.Map // map[string]*Chat
}

func (tg *TgWrapper) GetChat(identifier string) (*Chat, error) {
	if identifier == "" {
		return nil, ErrIdentifierEmpty
	}
	chatI, exist := tg.chatsByIdentifier.Load(identifier)
	if !exist {
		return nil, ErrIdentifierNotFound
	}
	chat, ok := chatI.(*Chat)
	if !ok {
		return nil, ErrIdentifierNotFound
	}
	return chat, nil
}

func (tg *TgWrapper) RegisterChat(conf SingleChatConf) (*Chat, error) {
	if conf.Identifier == "" {
		return nil, ErrIdentifierEmpty
	}
	if conf.ChatID == 0 {
		return nil, ErrZeroChatID
	}
	if conf.BotToken == "" {
		return nil, ErrEmptyBotToken
	}

	_, exist := tg.chatsByIdentifier.Load(conf.Identifier)
	if exist {
		return nil, ErrIdentifierAlreadyExists
	}

	var bot *tgbotapi.BotAPI
	bot, err := tgbotapi.NewBotAPI(conf.BotToken)
	if err != nil {
		return nil, err
	}

	tgChat := &Chat{
		Bot:         bot,
		ChatID:      conf.ChatID,
		ChatTopic:   conf.ChatTopic,
		Identifier:  conf.Identifier,
		Description: conf.Description,

		disableWebPagePreview: true,
		retry:                 3,
		retryInterval:         time.Second,

		managedMsgs: sync.Map{},
	}

	tg.chatsByIdentifier.Store(conf.Identifier, tgChat)
	return tgChat, nil
}
