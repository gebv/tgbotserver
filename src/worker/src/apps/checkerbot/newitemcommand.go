package checkerbot

import (
	"context"
	"models"
	"strings"

	"github.com/inpime/sdata"

	"gopkg.in/telegram-bot-api.v4"
)

type NewItemCommand struct{}

func (NewItemCommand) Name() string {
	return "newvision"
}

func (*NewItemCommand) Match(ctx context.Context, s *models.Services) bool {
	update := models.UpdateFromContext(ctx)
	if update.IsCommand() && update.Command == "new" {
		return true
	}

	if update.Page() == "newvision" {
		return true
	}

	return false
}

// TODO:

func (*NewItemCommand) Handler(ctx context.Context, s *models.Services) (err error) {
	update := models.UpdateFromContext(ctx)

	if update.IsCommand() && update.Page() != "newvision" {
		update.SetPage("newvision")
		// TODO: куда перенаправить?
		update.SetSection("setToneOfVision")
	}

	if update.IsCommand() && update.Command == "new" {
		// TODO: куда перенаправить?== "new" {
		update.SetSection("setToneOfVision")
	}

	switch update.Section() {
	case "setToneOfVision":
		if update.IsCommand() {
			msg := tgbotapi.NewMessage(update.FromChatID, "Добавление новой записи. Выберите, пожалуйста, тональность вашего видения: позитивное или негативное?")
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("😃 Позитивное"),
					tgbotapi.NewKeyboardButton("😡 Негативное"),
				),
			)
			update.Api.Send(msg)
		} else if update.IsMessage() {
			tone := getToneFromText(update.Message.Text)

			if tone == UnknownTone {
				msg := tgbotapi.NewMessage(update.FromChatID, "Укажите, пожалуйста, характер новой записи: позитивное или негативное?")
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("😃 Позитивное"),
						tgbotapi.NewKeyboardButton("😡 Негативное"),
					),
				)
				update.Api.Send(msg)
				return nil
			}
			update.User.State.M("tmpvision").Set("tone", tone)
			update.User.Changed = true

			update.SetSection("setTags")

			msg := tgbotapi.NewMessage(update.FromChatID, "Выбрали: "+tone.String()+"\n\nДобавьте ключевые слова (через запятую) определяющие то о чем вы пишете.\nНапример название организации или места.")
			msg.ReplyMarkup = tgbotapi.NewHideKeyboard(true)
			update.Api.Send(msg)
		}
	case "setTags":
		tags := ""

		if update.IsMessage() {
			tags = update.Message.Text
		}

		if len(tags) == 0 {
			msg := tgbotapi.NewMessage(update.FromChatID, "Добавьте, пожалуйста, ключевые слова определяющие то о чем вы пишете. Например название организации или места.")
			update.Api.Send(msg)
			return nil
		}

		update.User.State.M("tmpvision").Set("tags", tags)
		update.User.Changed = true

		update.SetSection("setDescription")
		msg := tgbotapi.NewMessage(update.FromChatID, "Добавьте описание")
		msg.ReplyMarkup = tgbotapi.NewHideKeyboard(true)
		update.Api.Send(msg)
	case "setDescription":
		description := ""

		if update.IsMessage() {
			description = update.Message.Text
		}

		if len(description) == 0 {
			msg := tgbotapi.NewMessage(update.FromChatID, "Добавьте, пожалуйста, описание вашего видения")
			update.Api.Send(msg)
			return nil
		}

		update.User.State.M("tmpvision").Set("description", description)
		update.User.Changed = true

		update.SetSection("setPhotos")
		msg := tgbotapi.NewMessage(update.FromChatID, "Добавьте изображение")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Пропустить"),
			),
		)
		update.Api.Send(msg)
	case "setPhotos":
		withPhotos := false

		if update.Message.Photo != nil && len(*update.Message.Photo) > 0 {
			withPhotos = true
		} else if update.IsMessage() && strings.ToLower(update.Message.Text) == "пропустить" {
			withPhotos = false
		}

		if withPhotos {
			files := sdata.NewArray()
			for _, file := range *update.Message.Photo {
				files.Add(file)
			}

			update.User.State.M("tmpvision").Set("photos", files)
			update.User.Changed = true
		}

		defer func() {
			update.User.State.M("tmpvision").Clear()
			update.User.Changed = true
		}()

		vision, err := CreateVisionFromOptions(update.User.ID,
			update.User.State.M("tmpvision"),
			s.DB)

		if err != nil {
			s.Logger.WithError(err).Error("create vision")
			update.SetSection("setToneOfVision")

			msg := tgbotapi.NewMessage(update.FromChatID, `Ошибка создания, повторите позже.`)
			msg.ReplyMarkup = tgbotapi.NewHideKeyboard(true)
			update.Api.Send(msg)
			// TODO: куда перенаправить?
			update.SetSection("setToneOfVision")

			return err
		}

		// TODO: формат отображения vision

		msg := tgbotapi.NewMessage(update.FromChatID, `Готово. 
		
		`+vision.ViewAsMessage())
		msg.ReplyMarkup = tgbotapi.NewHideKeyboard(true)
		update.Api.Send(msg)
		// TODO: куда перенаправить?
		update.SetSection("setToneOfVision")

	default:
	}
	return nil
}

func formatVision(opt *sdata.StringMap) string {
	res := opt.String("tone") + "\n"
	res += opt.String("tags") + "\n"
	res += opt.String("description") + "\n"

	return res
}

// Helpful types

type Tone string

func (t Tone) String() string {
	return string(t)
}

var PositiveTone Tone = "tones:positive"
var NegativeTone Tone = "tones:negative"
var UnknownTone Tone = "tones:unknown"

func getToneFromText(text string) Tone {
	text = strings.ToLower(text)
	if strings.HasSuffix(text, "позитивное") {
		return PositiveTone
	}

	if strings.HasSuffix(text, "негативное") {
		return NegativeTone
	}

	return UnknownTone
}
