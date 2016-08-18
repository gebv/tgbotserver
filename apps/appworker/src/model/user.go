package model

import (
	"time"

	"github.com/inpime/sdata"
	"gopkg.in/telegram-bot-api.v4"
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
