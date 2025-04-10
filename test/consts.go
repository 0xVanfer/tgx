package test

import "github.com/0xVanfer/tgx"

const BotToken = ""

const (
	IdentifierVanfer = "vanfer"
)

var TestChats = []tgx.SingleChatConf{
	{
		BotToken:   BotToken,
		ChatID:     0,
		Identifier: IdentifierVanfer,
	},
}

var TestMsgComponents = []tgx.MsgComponent{
	{
		Text: "Clean Text ",
	},
	{
		Text:        "bold text",
		EntitiyType: "bold",
	},
	{
		Text: " ",
	},
	{
		Text:        "italic text",
		EntitiyType: "italic",
	},
	{
		Text: " ",
	},
	{
		Text:        "underline text",
		EntitiyType: "underline",
	},
	{
		Text: " ",
	},
	{
		Text:        "strikethrough text",
		EntitiyType: "strikethrough",
	},
	{
		Text: "\n",
	},
	{
		Text:        "text link",
		EntitiyType: "text_link",
		URL:         "https://google.com",
	},
}
