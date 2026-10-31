package database

import "errors"

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrNoteNotFound         = errors.New("note not found")
	ErrConversationNotFound = errors.New("conversation not found")
)
