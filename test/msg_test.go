package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/0xVanfer/tgx"
)

func TestSendTextMsg(t *testing.T) {
	// Send a simple text message
	_, _ = chatDefault.SendTextMsg("hello world")
	_, _ = chatDefault.SendTextMsgByComponents(TestMsgComponents)
}

func TestSendLongText(t *testing.T) {
	var text string
	for range 1024 {
		text += "abcd"
	}
	text += "This start at 4096."
	_, _ = chatDefault.SendTextMsg(text)
}

func TestSendLongComponents(t *testing.T) {
	var components []tgx.MsgComponent
	// The amount of components with non empty entity type MUST be shorter then 100.
	for range 100 {
		components = append(components, tgx.MsgComponent{
			Text:        "abcd",
			EntitiyType: "bold",
		})
		components = append(components, tgx.MsgComponent{
			Text: "simple",
		})
	}
	_, err := chatDefault.SendTextMsgByComponents(components)
	fmt.Println(err)
}

func TestSendPhoto(t *testing.T) {
	photo0 := "https://ethereum.org/images/favicon.png"
	photo1 := "./favicon.png"

	chatDefault.SendTextMsg("this is a online photo")
	_, _ = chatDefault.SendPhoto(photo0, false)

	chatDefault.SendTextMsg("this is a local photo")
	_, _ = chatDefault.SendPhoto(photo1, true)
}

func TestEditMsg(t *testing.T) {
	msgIdentifierSimple := "msg_to_be_edited"
	msgIdentifierComplicated := "msg_complicated"

	// Send a simple text message
	msgSimple, _ := chatDefault.SendTextMsg("this is a msg to be edited")
	msgComplicated, _ := chatDefault.SendTextMsgByComponents(TestMsgComponents)

	chatDefault.RegisterMsgs(msgSimple, msgIdentifierSimple, "")
	chatDefault.RegisterMsgs(msgComplicated, msgIdentifierComplicated, "")

	time.Sleep(time.Second * 2)
	simple, _ := chatDefault.GetMsg(msgIdentifierSimple)
	_ = simple.EditText("this is a edited msg")

	time.Sleep(time.Second * 2)
	_ = simple.ReplaceWith(msgComplicated[0])

	time.Sleep(time.Second * 2)
	_ = simple.Delete()

	chatDefault.DeleteMsgs(msgIdentifierSimple)
	chatDefault.DeleteMsgs(msgIdentifierComplicated)
}
