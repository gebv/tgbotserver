package models

import (
	"github.com/inpime/sdata"
	"gopkg.in/telegram-bot-api.v4"
)

type RequestMethod string
type ChatType string

var (
	CommandRequest        RequestMethod = "request:command"
	InlineRequest         RequestMethod = "request:inline"
	InlineCallbackRequest RequestMethod = "request:inline_callback"
	MessageRequest        RequestMethod = "request:message"

	UnknownTypeChat ChatType = "chat:types:unknown"
	ChannelChat     ChatType = "chat:types:channel"
	GroupChat       ChatType = "chat:types:group"
	PrivateChat     ChatType = "chat:types:private"
	SuperGroupChat  ChatType = "chat:types:super_group"
)

func NewUpdate(user *User, update tgbotapi.Update, api *tgbotapi.BotAPI) *Update {
	res := &Update{
		Update: update,
		User:   user,
		Data:   sdata.NewStringMap(),
		Api:    api,
	}

	if update.Message != nil && update.Message.IsCommand() {
		res.Method = CommandRequest
		res.FromChatID = update.Message.Chat.ID
		res.ChatType = GetChatType(update.Message.Chat)
		res.Command = update.Message.Command()
		res.CommandArgs = update.Message.CommandArguments()
	} else if update.Message != nil {
		res.Method = MessageRequest
		res.FromChatID = update.Message.Chat.ID
		res.ChatType = GetChatType(update.Message.Chat)
	} else if update.InlineQuery != nil {
		res.Method = InlineRequest
		res.FromChatID = 0
	} else if update.CallbackQuery != nil {
		res.Method = InlineCallbackRequest
		res.InlineCallbackID = update.CallbackQuery.ID
		res.InlineCallbackValue = update.CallbackQuery.Data
		res.InlineMessageID = update.CallbackQuery.InlineMessageID

		if update.CallbackQuery.Message != nil {
			res.FromChatID = update.CallbackQuery.Message.Chat.ID
			res.ChatType = GetChatType(update.CallbackQuery.Message.Chat)
			res.InlineCallbackFromMessageID = update.CallbackQuery.Message.MessageID
		}
	}

	return res
}

type Update struct {
	tgbotapi.Update

	Api *tgbotapi.BotAPI

	Method RequestMethod

	User *User

	FromChatID int64
	ChatType   ChatType

	Command     string
	CommandArgs string

	InlineCallbackID            string
	InlineMessageID             string // only for inline query
	InlineCallbackValue         string
	InlineCallbackFromMessageID int

	Data *sdata.StringMap
}

// IsMessage returns true if the update type message text
func (c *Update) IsMessage() bool {
	return c.Method == MessageRequest
}

func (c *Update) IsCommand() bool {
	return c.Method == CommandRequest
}

func (c *Update) IsInline() bool {
	return c.Method == InlineRequest
}

func (c *Update) IsCallback() bool {
	return c.Method == InlineCallbackRequest
}

// Page
func (c *Update) Page() string {
	return c.User.Page
}

// SetPage
func (c *Update) SetPage(v string) {
	c.User.Changed = true
	c.User.Page = v
}

// Section
func (c *Update) Section() string {
	return c.User.Section
}

// SetSection
func (c *Update) SetSection(v string) {
	c.User.Changed = true
	c.User.Section = v
}
