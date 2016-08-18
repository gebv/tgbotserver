package handlers

import (
	"context"

	"gopkg.in/telegram-bot-api.v4"
)

type EchoHandler struct {
}

func (*EchoHandler) Name() string {
	return "echo"
}

func (*EchoHandler) Match(ctx context.Context, s *Services) bool {
	return true
}

func (*EchoHandler) Handler(ctx context.Context, s *Services) error {
	update := UpdateFromContext(ctx)
	update.Api.Send(tgbotapi.NewMessage(update.FromChatID, "ECHO: "+update.Message.Text))
	return nil
}
