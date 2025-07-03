package tgx

import (
	"sync"
	"time"

	"github.com/0xVanfer/tgx/internal/tgxerrors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// The wrapper for telegram bot API to record the chats information.
//
// Use RegisterChat() to register a chat;
// Use GetChat() to get the chat information.
//
// More functions are avaliable under the return struct of GetChat().
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
	if conf.ChatID == 0 {
		return nil, tgxerrors.ErrZeroChatID
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
	updatesConf.Timeout = 10

	bots := tg.GetAllRegisteredBots()
	for _, info := range bots {
		go func(b *botInfo) {
			bot := b.Bot
			if bot == nil {
				return
			}
			updates := bot.GetUpdatesChan(updatesConf)
			for update := range updates {
				if update.Message == nil || update.Message.Chat == nil {
					continue
				}
				// Search for the chat by chat ID.
				for _, chat := range b.Chats {
					if chat.ChatID != update.Message.Chat.ID {
						continue
					}
					// Consider the messageID of the msg reply to is the topic.
					// Replies under topic will be ignored.
					var topic int
					if update.Message.ReplyToMessage != nil {
						topic = update.Message.ReplyToMessage.MessageID
					}
					// If chat topic is set to negative, ignore the topic. It will handle all messages under the chatID.
					if chat.ChatTopic >= 0 && topic != chat.ChatTopic {
						continue
					}

					chat.HandleCommand(update.Message)
					chat.HandleMsg(update.Message)
				}
			}
		}(info)
	}
}
