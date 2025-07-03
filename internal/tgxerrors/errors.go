package tgxerrors

import (
	"errors"
)

// Common errors
var (
	ErrIdentifierEmpty         = errors.New("tgx: identifier is empty")       // Identifier is "".
	ErrIdentifierNotFound      = errors.New("tgx: identifier not found")      // Identifier not found when getting.
	ErrIdentifierAlreadyExists = errors.New("tgx: identifier already exists") // Identifier already exist when registering.

	ErrTextTooLong     = errors.New("tgx: text length is too long")     // Text length > 4096.
	ErrTooManyEntities = errors.New("tgx: entities length is too long") // Entities length > 100.

	ErrZeroChatID    = errors.New("tgx: chat_id is 0")
	ErrEmptyBotToken = errors.New("tgx: bot_token is empty")

	ErrMsgNotFound = errors.New("tgx: msg or msg.Chat is nil")
)
