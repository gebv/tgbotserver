package main

import (
	"time"

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

// NewUser new user
func NewUser(id int) *User {
	return &User{
		ID:           id,
		State:        sdata.NewStringMap(),
		Created:      time.Now(),
		IsOpenDialog: true,
		DialogID:     id,
	}
}

func UserFromUpdate(update tgbotapi.Update) *User {
	var user *tgbotapi.User

	if update.InlineQuery != nil {
		user = update.InlineQuery.From
	} else if update.CallbackQuery != nil {
		user = update.CallbackQuery.From
	} else if update.Message != nil {
		user = update.Message.From
	} else {
		return nil
	}

	return TransformToUser(user)
}

// TransformToUser transform telegram-user in user
func TransformToUser(user *tgbotapi.User) *User {
	return &User{
		ID: user.ID,

		FirstName: user.FirstName,
		LastName:  user.LastName,
		UserName:  user.UserName,

		State: sdata.NewStringMap(),
	}
}

// LoadUserOrCreate
func LoadUserOrCreate(user *User) error {
	_, err := db.Model(user).
		Where("id = ? ", user.ID).
		SelectOrCreate()

	if err != nil {
		return err
	}

	return nil
}

type User struct {
	ID int `sql:"id"`

	FirstName string `sql:"fname"`
	LastName  string `sql:"lname"`
	UserName  string `sql:"uname"`
	Phone     string `sql:"phone"`

	State *sdata.StringMap // values ​​stored

	IsOpenDialog bool `sql:"is_open_dialog"`
	DialogID     int  `sql:"dialog_id"`

	Page    string `sql:"page"`
	Section string `sql:"section"` // section of page

	Created time.Time `sql:"created"`
	Updated time.Time `sql:"updated"`
}

// ------------
// Update
// ------------

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
		res.ChatType = getChatType(update.Message.Chat)
	} else if update.Message != nil {
		res.Method = MessageRequest
		res.FromChatID = update.Message.Chat.ID
		res.ChatType = getChatType(update.Message.Chat)
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
			res.ChatType = getChatType(update.CallbackQuery.Message.Chat)
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

	InlineCallbackID            string
	InlineMessageID             string // only for inline query
	InlineCallbackValue         string
	InlineCallbackFromMessageID int

	Data *sdata.StringMap
}

// Page
func (c *Update) Page() string {
	return c.User.Page
}

// SetPage
func (c *Update) SetPage(v string) {
	c.User.Page = v
}

// Section
func (c *Update) Section() string {
	return c.User.Section
}

// SetSection
func (c *Update) SetSection(v string) {
	c.User.Section = v
}
