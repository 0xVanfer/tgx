package tgx

import (
	"sync"
)

// The structure to initialize the Telegram wrapper.
//
// ChatID and ChatTopic can be read by the following steps:
//
// 0. If you want to register the entire group, or the group does not allow topics, set ChatTopic to -1 and ignore the logics about topic in following steps.
//
// 1. Find the topic (or group) you want to register.
//
// 2. Selete any message in the topic (or group), right click and select "Copy Message Link".
//
// 3. The link will look like this: https://t.me/c/123456789/2/21 (If topic is not allowed, the link will look like this: https://t.me/c/123456789/21).
//
// 4. The ChatID is -100123456789 (adding -100 at the beginning), and the ChatTopic is 2. 21 is the message ID, which is not used in this package.
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
