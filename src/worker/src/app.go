package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Sirupsen/logrus"

	"time"

	models "models"

	pg "gopkg.in/pg.v4"
	"gopkg.in/telegram-bot-api.v4"
)

// Execute
func Execute(msg []byte, timeout time.Duration) error {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	if timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}

	defer cancel()

	return handler(ctx, msg)
}

func UnmarshalUpdate(msg []byte) (update tgbotapi.Update, err error) {
	return update, json.Unmarshal(msg, &update)
}

func handler(ctx context.Context, msg []byte) error {

	var (
		data *models.Update
	)

	update, err := UnmarshalUpdate(msg)

	if err != nil {
		return err
	}

	userCh := loadUser(ctx, update)

	// get a user

	select {
	case <-ctx.Done():
		return ctx.Err()
	case user := <-userCh:
		if user == nil {
			return fmt.Errorf("empty user")
		}

		data = models.NewUpdate(user, update, tgapi)
	}

	// TODO: how check dedline?

	// process data

	resCh := process(ctx, data)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-resCh:
		return err
	}
}

func loadUser(ctx context.Context, update tgbotapi.Update) <-chan *models.User {
	resCh := make(chan *models.User, 1)

	go func() {
		user := models.UserFromUpdate(update)

		if user == nil {
			logger.Error("the update is not from the user")

			close(resCh)
			return
		}

		// query to database

		created, err := db.Model(user).
			Where("id = ? ", user.ID).
			SelectOrCreate()

		if err != nil {
			logger.WithError(err).Error("load user")

			close(resCh)
			return
		}

		user.IsNew = created

		resCh <- user
	}()

	return resCh
}

func process(ctx context.Context, data *models.Update) <-chan error {
	resCh := make(chan error, 1)

	go func() {
		resCh <- appEntryPoint(ctx, data, db, logger)
	}()

	return resCh
}

type AppEntryPoint func(context.Context, *models.Update, *pg.DB, *logrus.Entry) error
type AppDBSchemaCreator func(*pg.DB) error
