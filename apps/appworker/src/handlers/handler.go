package handlers

import (
	"context"
	"sync"

	"github.com/Sirupsen/logrus"

	"model"

	pg "gopkg.in/pg.v4"
)

var handlers []Handler
var handlermutex sync.RWMutex

func init() {
	handlers = []Handler{
		&EchoHandler{},
	}
}

type Services struct {
	DB *pg.DB
}

type Handler interface {
	Name() string
	Match(context.Context, *Services) bool
	Handler(context.Context, *Services) error
}

func Execute(ctx context.Context, update *model.Update, db *pg.DB) error {
	handlermutex.RLock()
	defer handlermutex.RUnlock()

	ctx = context.WithValue(ctx, "update", update)
	services := &Services{
		DB: db,
	}

	for _, handler := range handlers {
		if handler.Match(ctx, services) {
			logrus.Info("matched %q", handler.Name())
			if err := handler.Handler(ctx, services); err != nil {
				logrus.WithError(err).Error("matched %q", handler.Name())
				return err
			}
		}
	}

	return nil
}

func UpdateFromContext(ctx context.Context) *model.Update {
	return ctx.Value("update").(*model.Update)
}
