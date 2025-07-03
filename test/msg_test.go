package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/0xVanfer/tgx"
)

// Sending two simple messeges to the topic.
func TestSendTextMsg(t *testing.T) {
	_, _ = msgTopicChat.SendTextMsg(nil, "hello world")
	_, _ = msgTopicChat.SendTextMsgByComponents(nil, TestMsgComponents)
}

// Sending a long text to the topic.
// Expected to receive three msgs instead of one. Text will be spilt at length 4096 * n.
func TestSendLongText(t *testing.T) {
	var text string
	for range 2048 {
		text += "abcd"
	}
	text += "This start at 8192."
	_, _ = msgTopicChat.SendTextMsg(nil, text)
}

// Entities length must be shorter than 100 in one text.
// Since it's hard to decide how to spilt the entities, we just return error.
// If the range here is 100, no error will be returned.
// If the range is 101 or more, an error "tgx: entities length is too long" is expected.
func TestSendLongComponents(t *testing.T) {
	var components []tgx.MsgComponent
	// The amount of components with non empty entity type MUST be shorter then 100.
	for range 101 {
		components = append(components, tgx.MsgComponent{
			Text:        "abcd",
			EntitiyType: "bold",
		})
		components = append(components, tgx.MsgComponent{
			Text: "simple",
		})
	}
	_, err := msgTopicChat.SendTextMsgByComponents(nil, components)
	fmt.Println(err)
}

// Sending 2 pics from online and local.
func TestSendPhoto(t *testing.T) {
	photo0 := "https://ethereum.org/images/favicon.png"
	photo1 := "../internal/assets/favicon.png"

	msgTopicChat.SendTextMsg(nil, "this is a online photo")
	_, _ = msgTopicChat.SendPhoto(nil, photo0, false)

	msgTopicChat.SendTextMsg(nil, "this is a local photo")
	_, _ = msgTopicChat.SendPhoto(nil, photo1, true)
}

// Relatively complicated message handling.
// Will register the msg send and then edit it, replace it with another msg and finally delete it.
func TestEditMsg(t *testing.T) {
	msgIdentifierSimple := "msg_to_be_edited"
	msgIdentifierComplicated := "msg_complicated"

	// Send a simple text message
	msgSimple, _ := msgTopicChat.SendTextMsg(nil, "This is a msg to be edited")
	msgComplicated, _ := msgTopicChat.SendTextMsgByComponents(nil, TestMsgComponents)

	// Register the messages with their identifiers
	msgTopicChat.RegisterMsgs(msgSimple, msgIdentifierSimple, "")
	msgTopicChat.RegisterMsgs(msgComplicated, msgIdentifierComplicated, "")

	// Edit the simple message
	time.Sleep(time.Second * 2)
	simple, _ := msgTopicChat.GetMsg(msgIdentifierSimple)
	_ = simple.EditText("this is a edited msg")

	// Edit the message and make it same as the target one.
	time.Sleep(time.Second * 2)
	_ = simple.ReplaceWith(msgComplicated[0])

	// Finally delete the messages.
	time.Sleep(time.Second * 2)
	_ = simple.Delete()

	// Delete the registered messages by their identifiers.
	msgTopicChat.DeleteMsgs(msgIdentifierSimple)
	msgTopicChat.DeleteMsgs(msgIdentifierComplicated)
}
