package models

import (
	"gopkg.in/telegram-bot-api.v4"
)

func GetChatType(chat *tgbotapi.Chat) ChatType {
	if chat.IsChannel() {
		return ChannelChat
	}

	if chat.IsGroup() {
		return GroupChat
	}

	if chat.IsPrivate() {
		return PrivateChat
	}

	if chat.IsSuperGroup() {
		return SuperGroupChat
	}

	return UnknownTypeChat
}
