package tgx

import (
	"sync"
)

type SingleChatConf struct {
	BotToken    string `json:"bot_token"`
	ChatID      int64  `json:"chat_id"`
	ChatTopic   int    `json:"chat_topic"`
	Identifier  string `json:"identifier"`
	Description string `json:"description,omitempty"`
}

// Each conf should be valid, otherwise will return an error.
func Init(conf ...SingleChatConf) (tg *TgWrapper, err error) {
	tg = &TgWrapper{chatsByIdentifier: sync.Map{}}

	for _, c := range conf {
		_, err = tg.RegisterChat(c)
		if err != nil {
			return nil, err
		}
	}
	return tg, nil
}
