package checkerbot

import (
	"context"
	"models"

	"gopkg.in/telegram-bot-api.v4"
)

var (
	WelcomeMessage = `Сообщение для новых пользователей`
)

type TipForNewUser struct{}

func (*TipForNewUser) Name() string {
	return "TipForNewUser"
}

func (*TipForNewUser) Match(ctx context.Context, s *models.Services) bool {
	update := models.UpdateFromContext(ctx)
	if !update.User.IsNew {
		return false
	}
	return true
}

func (*TipForNewUser) Handler(ctx context.Context, s *models.Services) error {
	update := models.UpdateFromContext(ctx)

	if update.IsMessage() {
		update.Api.Send(tgbotapi.NewMessage(update.FromChatID, WelcomeMessage))
	}

	return nil
}
