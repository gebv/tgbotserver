package checkerbot

import (
	"fmt"
	"strings"

	pg "gopkg.in/pg.v4"

	"github.com/Sirupsen/logrus"
	"github.com/inpime/sdata"
	"github.com/satori/go.uuid"
	"gopkg.in/telegram-bot-api.v4"
)

type VisionStatus string

var CreateWaitApproval VisionStatus = "status:create:wait_approval"
var WaitApproval VisionStatus = "status:update:wait_approval"
var ApprovalVision VisionStatus = "status:approval"
var DeclinedVision VisionStatus = "status:declined"
var HiddenUser VisionStatus = "status:hidden_user" // статус если пользователь
// сркыл не дождавшись подтверждения от менеджеров

func NewVision() *Vision {
	return &Vision{
		Status: CreateWaitApproval, // default status
	}
}

type Vision struct {
	ID      string `sql:"id"`
	OwnerID int    `sql:"owner_id"`

	Tone        string   `sql:"tone"` // positive OR negative
	Description string   `sql:"description"`
	Tags        []string `sql:"tags" pg:",array"`

	Status       VisionStatus `sql:"status"`
	StatusReason string       `sql:"status_reason"` // для проблемных
	Enabled      bool         `sql:"enabled"`       // опция пользователя
}

func (v Vision) ViewAsMessage() string {
	return fmt.Sprintf(`/vision_%s
%s`,
		v.String(),
		v.Description)
}

func (v Vision) String() string {
	return v.ID[:7]
}

type PhotoVision struct {
	TableName struct{} `sql:"vision_photos"`
	ID        string   `sql:"id"`
	ExtID     string   `sql:"ext_id"`
	VisionID  string   `sql:"vision_id"` // relation of vision

	Width  int `sql:"w"`
	Height int `sql:"h"`
	Size   int `sql:"size"`
}

func (v PhotoVision) String() string {
	return v.ID[:7]
}

func CreateVisionFromOptions(owenrid int, opt *sdata.StringMap, db *pg.DB) (*Vision, error) {
	tone := opt.String("tone")
	description := opt.String("description")

	tagsarr := strings.Split(opt.String("tags"), ",")
	tags := make([]string, len(tagsarr))
	for index, tag := range tagsarr {
		tags[index] = strings.ToLower(strings.TrimSpace(tag))

	}
	vision := &Vision{
		ID:          uuid.NewV4().String(),
		OwnerID:     owenrid,
		Tone:        tone,
		Tags:        tags,
		Description: description,
		Status:      CreateWaitApproval,
	}

	tx, err := db.Begin()

	if err != nil {
		return vision, err
	}

	if _, err := tx.Model(vision).Create(); err != nil {
		return vision, err
	}

	for _, file := range opt.A("photos").Data() {
		tgphoto, ok := file.(tgbotapi.PhotoSize)
		if !ok {
			logrus.Error("not expected struct file type %T", file)
			continue
		}

		photo := &PhotoVision{
			ID:       uuid.NewV4().String(),
			ExtID:    tgphoto.FileID,
			VisionID: vision.ID,
			Width:    tgphoto.Width,
			Height:   tgphoto.Height,
			Size:     tgphoto.FileSize,
		}

		if _, err := tx.Model(photo).Create(); err != nil {
			logrus.WithError(err).Error("create photo for vision")
			continue
		}
	}

	return vision, tx.Commit()
}
