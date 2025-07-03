package test

import "github.com/0xVanfer/tgx"

var (
	wrapper *tgx.TgWrapper

	entireChat        *tgx.Chat
	generalChat       *tgx.Chat
	msgTopicChat      *tgx.Chat
	monitorTopicChat  *tgx.Chat
	monitorEverywhere *tgx.Chat
)

func init() {
	var err error
	wrapper, err = tgx.Init(TestChats...)
	if err != nil {
		panic(err)
	}

	entireChat, err = wrapper.GetChat(IdentifierTestEntireChat)
	if err != nil {
		panic(err)
	}
	generalChat, err = wrapper.GetChat(IdentifierTestGeneral)
	if err != nil {
		panic(err)
	}
	msgTopicChat, err = wrapper.GetChat(IdentifierTestSendMsg)
	if err != nil {
		panic(err)
	}
	monitorTopicChat, err = wrapper.GetChat(IdentifierTestMonitorCommand)
	if err != nil {
		panic(err)
	}
	monitorEverywhere, err = wrapper.GetChat(IdentifierEverywhere)
	if err != nil {
		panic(err)
	}
}

// Change to your own bot token
var BotToken = ""

const (
	IdentifierTestEntireChat     = "test_entire_chat"
	IdentifierTestGeneral        = "test_general"
	IdentifierTestSendMsg        = "test_send_msg"
	IdentifierTestMonitorCommand = "test_monitor_command"

	IdentifierEverywhere = "test_everywhere"
)

var TestChats = []tgx.SingleChatConf{
	{
		BotToken:   BotToken,
		ChatID:     0, // 0 to say it can be used everywhere
		ChatTopic:  0,
		Identifier: IdentifierEverywhere,
	},
	{
		BotToken:   BotToken,
		ChatID:     -1001111111111, // Change to your own chat ID
		ChatTopic:  -1,
		Identifier: IdentifierTestEntireChat,
	},
	{
		BotToken:   BotToken,
		ChatID:     -1001111111111,
		ChatTopic:  0,
		Identifier: IdentifierTestGeneral,
	},
	{
		BotToken:   BotToken,
		ChatID:     -1001111111111,
		ChatTopic:  2, // Change to your own topic ID
		Identifier: IdentifierTestSendMsg,
	},
	{
		BotToken:   BotToken,
		ChatID:     -1001111111111,
		ChatTopic:  5,
		Identifier: IdentifierTestMonitorCommand,
	},
}

var TestMsgComponents = []tgx.MsgComponent{
	{Text: "Clean Text "},
	{Text: "bold text", EntitiyType: "bold"},
	{Text: " "},
	{Text: "italic text", EntitiyType: "italic"},
	{Text: " "},
	{Text: "underline text", EntitiyType: "underline"},
	{Text: " "},
	{Text: "strikethrough text", EntitiyType: "strikethrough"},
	{Text: "\n"},
	{Text: "text link", EntitiyType: "text_link", URL: "https://google.com"},
	{Text: " @Vanfer0x", EntitiyType: "mention"},
}
