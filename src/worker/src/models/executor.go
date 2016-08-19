package models

import (
	"context"

	"github.com/Sirupsen/logrus"
	pg "gopkg.in/pg.v4"
)

func Execute(h *Handlers, ctx context.Context, s *Services) error {
	for _, handler := range h.List() {
		if handler.Match(ctx, s) {
			s.Logger.Infof("matched %q", handler.Name())
			if err := handler.Handler(ctx, s); err != nil {
				s.Logger.WithError(err).Errorf("matched %q", handler.Name())
				return err
			}
		}
	}
	return nil
}

func UpdateFromContext(ctx context.Context) *Update {
	return ctx.Value("update").(*Update)
}

func ContextWithUpdate(ctx context.Context, update *Update) context.Context {
	return context.WithValue(ctx, "update", update)
}

type Services struct {
	DB     *pg.DB
	Logger *logrus.Entry
}
