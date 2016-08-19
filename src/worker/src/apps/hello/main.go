package hello

import (
	"context"
	"models"

	"github.com/Sirupsen/logrus"

	pg "gopkg.in/pg.v4"
	"gopkg.in/telegram-bot-api.v4"
)

var h = models.NewHandlers()
var NAME = "hello"

func init() {
	h.Add(&Hello{})
}

func Execute(ctx context.Context, update *models.Update, db *pg.DB, logger *logrus.Entry) error {
	ctx = context.WithValue(ctx, "update", update)

	s := &models.Services{
		DB:     db,
		Logger: logger.WithField("_bot", NAME),
	}

	if update.IsMessage() {
		update.Api.Send(tgbotapi.NewChatAction(update.FromChatID, "typing"))
	}

	if err := models.Execute(h, ctx, s); err != nil {
		s.Logger.WithError(err).Errorln("execute app")
		return err
	}

	// after routing

	if userAfter := models.UpdateFromContext(ctx).User; userAfter.Changed {
		err := s.DB.Update(userAfter)

		if err != nil {
			s.Logger.WithError(err).Errorln("save a user after routing")
		}

		return err
	}

	return nil
}
