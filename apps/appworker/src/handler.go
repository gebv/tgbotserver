package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	h "handlers"
	"model"

	"github.com/Sirupsen/logrus"

	"gopkg.in/telegram-bot-api.v4"
)

// Handler
func Handler(msg []byte, timeout time.Duration) error {
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
		data *model.Update
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

		data = model.NewUpdate(user, update, api)
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

func loadUser(ctx context.Context, update tgbotapi.Update) <-chan *model.User {
	resCh := make(chan *model.User, 1)

	go func() {
		user := model.UserFromUpdate(update)

		if user == nil {
			logrus.WithFields(logrus.Fields{
				"_ref":    NAME,
				"_host":   hostname,
				"_client": clientID,
			}).Error("the update is not from the user")

			close(resCh)
			return
		}

		// db

		_, err := db.Model(user).
			Where("id = ? ", user.ID).
			SelectOrCreate()

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"_ref":    NAME,
				"_host":   hostname,
				"_client": clientID,

				"err": err,
			}).Error("load user")

			close(resCh)
			return
		}

		resCh <- user
	}()

	return resCh
}

func process(ctx context.Context, data *model.Update) <-chan error {
	resCh := make(chan error, 1)

	go func() {
		resCh <- h.Execute(ctx, data, db)
	}()

	return resCh
}
