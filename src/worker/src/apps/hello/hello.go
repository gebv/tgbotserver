package hello

import (
	"context"
	"models"

	"gopkg.in/telegram-bot-api.v4"
)

type Hello struct{}

func (*Hello) Name() string {
	return "hello"
}

func (*Hello) Match(ctx context.Context, s *models.Services) bool {
	update := models.UpdateFromContext(ctx)
	if !update.IsMessage() {
		return false
	}
	return true
}

func (*Hello) Handler(ctx context.Context, s *models.Services) error {
	update := models.UpdateFromContext(ctx)
	update.Api.Send(tgbotapi.NewMessage(update.FromChatID, "Hello '"+update.User.FirstName+"'"))
	return nil
}
