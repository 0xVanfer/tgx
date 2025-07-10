package tgx

import (
	"fmt"
	"sync"
	"time"

	"github.com/0xVanfer/tgx/internal/tgxerrors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// The wrapper for telegram bot API to record the chats information.
//
// Supports:
// - Sending and managing messages in the registered chats.
// - Monitor all registered bots for incoming commands.
//
// Use RegisterChat() to register a chat;
// Use GetChat() to get the chat information.
type TgWrapper struct {
	chatsByIdentifier sync.Map // map[identifier(string)]*Chat

	allRelatedBots sync.Map // map[bot token(string)][]*Chat
}

// Get the chat information by identifier.
func (tg *TgWrapper) GetChat(identifier string) (*Chat, error) {
	if identifier == "" {
		return nil, tgxerrors.ErrIdentifierEmpty
	}
	chatI, exist := tg.chatsByIdentifier.Load(identifier)
	if !exist {
		return nil, tgxerrors.ErrIdentifierNotFound
	}
	chat, ok := chatI.(*Chat)
	if !ok {
		return nil, tgxerrors.ErrIdentifierNotFound
	}
	return chat, nil
}

// Register a chat with the given configuration.
func (tg *TgWrapper) RegisterChat(conf SingleChatConf) (*Chat, error) {
	if conf.Identifier == "" {
		return nil, tgxerrors.ErrIdentifierEmpty
	}
	if conf.BotToken == "" {
		return nil, tgxerrors.ErrEmptyBotToken
	}

	// Identifier should be unique.
	_, exist := tg.chatsByIdentifier.Load(conf.Identifier)
	if exist {
		return nil, tgxerrors.ErrIdentifierAlreadyExists
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

		managedMsgs:        sync.Map{},
		handleCommandFuncs: make(map[string]func(msg *tgbotapi.Message) (err error)),
		handleMsgFuncs:     make(map[string]func(msg *tgbotapi.Message) (err error)),
	}

	tg.chatsByIdentifier.Store(conf.Identifier, tgChat)

	if bots, ok := tg.allRelatedBots.Load(conf.BotToken); !ok {
		tg.allRelatedBots.Store(conf.BotToken, []*Chat{tgChat})
	} else {
		tg.allRelatedBots.Store(conf.BotToken, append(bots.([]*Chat), tgChat))
	}
	return tgChat, nil
}

type botInfo struct {
	Bot   *tgbotapi.BotAPI
	Chats []*Chat
}

func (b *botInfo) hasHandler() bool {
	for _, chat := range b.Chats {
		if len(chat.handleCommandFuncs) > 0 || len(chat.handleMsgFuncs) > 0 {
			return true
		}
	}
	return false
}

// By reading 'tg.allRelatedBots', we can get all registered bots.
func (tg *TgWrapper) GetAllRegisteredBots() (bots map[string]*botInfo) {
	bots = make(map[string]*botInfo)
	tg.allRelatedBots.Range(func(key any, value any) bool {
		// make sure at least one chat is registered for the bot
		botToken, ok := key.(string)
		if !ok {
			return true
		}
		chats, ok := value.([]*Chat)
		if !ok {
			return true
		}
		if len(chats) == 0 {
			return true
		}
		if _, exist := bots[botToken]; !exist {
			bots[botToken] = &botInfo{
				Bot:   chats[0].Bot,
				Chats: chats,
			}
		} else {
			bots[botToken].Chats = append(bots[botToken].Chats, chats...)
		}

		return true
	})
	return
}

// Monitor all registered bots for incoming commands.
func (tg *TgWrapper) Monitor() {
	updatesConf := tgbotapi.NewUpdate(0)
	updatesConf.Timeout = 10 // TODO: make it configurable?

	bots := tg.GetAllRegisteredBots()
	for _, info := range bots {
		if info.Bot == nil {
			// No bot registered, skip this bot.
			continue
		}
		if !info.hasHandler() {
			// No handler registered, skip this bot.
			continue
		}
		go func(b *botInfo) {
			bot := b.Bot
			updates := bot.GetUpdatesChan(updatesConf)
			for update := range updates {
				if update.Message == nil || update.Message.Chat == nil {
					continue
				}
				// Actually will not use this.
				if update.Message.From.ID == bot.Self.ID {
					continue
				}
				// Search for the chat by chat ID.
				for _, chat := range b.Chats {
					// If chat.ChatID is 0, should handle all messages.
					// Otherwise, only handle messages in the chat with the same chat ID.
					if chat.ChatID != 0 && chat.ChatID != update.Message.Chat.ID {
						continue
					}
					// Consider the messageID of the msg reply to is the topic.
					// Replies under topic will be ignored.
					var topic int
					if update.Message.ReplyToMessage != nil {
						topic = update.Message.ReplyToMessage.MessageID
					}
					// If chat topic is set to negative, ignore the topic. It will handle all messages under the chatID.
					if chat.ChatID != 0 && chat.ChatTopic >= 0 && topic != chat.ChatTopic {
						continue
					}

					err := chat.HandleCommand(update.Message)
					if err != nil {
						fmt.Printf("Error in chat [%s]: %s\n", chat.Identifier, err.Error())
					}
					errors := chat.HandleMsg(update.Message)
					for _, err := range errors {
						fmt.Printf("Error in chat [%s]: %s\n", chat.Identifier, err.Error())
					}
				}
			}
		}(info)
	}
}
