package tgx

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// The amount of components with non empty entity type MUST be length then 100.
type MsgComponent struct {
	Text        string // The original text
	EntitiyType string // The type of the entity, can be "bold", "italic", "underline", "strikethrough", "text_link"
	URL         string // The URL for the text_link entity

	// User        tgbotapi.User // The user for the text_mention entity
}

func (msg *MsgComponent) Length() int { return len(msg.Text) }

func CompileMsgComponents(components ...MsgComponent) (text string, entities []tgbotapi.MessageEntity) {
	for _, component := range components {
		currentLength := len(text)
		switch component.EntitiyType {
		case "bold":
			text += component.Text
			entities = append(entities, tgbotapi.MessageEntity{
				Type:   component.EntitiyType,
				Offset: currentLength,
				Length: len(component.Text),
			})
		case "italic":
			text += component.Text
			entities = append(entities, tgbotapi.MessageEntity{
				Type:   component.EntitiyType,
				Offset: currentLength,
				Length: len(component.Text),
			})
		case "underline":
			text += component.Text
			entities = append(entities, tgbotapi.MessageEntity{
				Type:   component.EntitiyType,
				Offset: currentLength,
				Length: len(component.Text),
			})
		case "strikethrough":
			text += component.Text
			entities = append(entities, tgbotapi.MessageEntity{
				Type:   component.EntitiyType,
				Offset: currentLength,
				Length: len(component.Text),
			})
		case "text_link":
			text += component.Text
			entities = append(entities, tgbotapi.MessageEntity{
				Type:   component.EntitiyType,
				Offset: currentLength,
				Length: len(component.Text),
				URL:    component.URL,
			})
		case "mention":
			text += component.Text
			entities = append(entities, tgbotapi.MessageEntity{
				Type:   component.EntitiyType,
				Offset: currentLength,
				Length: len(component.Text),
			})
		default:
			text += component.Text
		}
	}
	return
}
