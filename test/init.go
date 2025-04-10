package test

import "github.com/0xVanfer/tgx"

var (
	wrapper     *tgx.TgWrapper
	chatDefault *tgx.Chat
)

func init() {
	var err error
	wrapper, err = tgx.Init(TestChats...)
	if err != nil {
		panic(err)
	}
	chatDefault, err = wrapper.GetChat(IdentifierVanfer)
	if err != nil {
		panic(err)
	}
}
