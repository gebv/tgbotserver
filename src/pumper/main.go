package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/nats-io/nats"
	"gopkg.in/telegram-bot-api.v4"
)

var (
	NAME    = "pumper"
	VERSION = "0.0.2"

	hostname = os.Getenv("HOSTNAME")
	appname  = os.Getenv("APPNAME")
	loglevel = os.Getenv("LOGLEVEL")
	PID      = NAME + "_" +
		VERSION + "_" +
		appname + "_" +
		hostname

	natsaddr = os.Getenv("NATSADDR")
	pubname  = os.Getenv("PUBNAME")

	tgtoken         = os.Getenv("TG_TOKEN")
	pumpPeriodMSStr = os.Getenv("PUMPPERIODMS")

	logger *logrus.Entry
	tgapi  *tgbotapi.BotAPI
	nc     *nats.Conn

	countMaxRetry         = 5
	sleepTimeout          = 3 * time.Second        // seconds
	pumpPeriod            = time.Millisecond * 100 // default 100 ms
	TGAPIEntrypointFormat = "https://api.telegram.org/bot%s/%s"
)

func init() {
	logLevel, err := logrus.ParseLevel(loglevel)
	if err != nil {
		logLevel = logrus.ErrorLevel // Default logger
	}
	logrus.SetLevel(logLevel)

	logger = logrus.WithFields(logrus.Fields{
		"_pid":   PID,
		"_group": NAME,
		"_name":  appname,
	})

	// custom pumper period

	if len(pumpPeriodMSStr) > 0 {
		period, _ := strconv.Atoi(pumpPeriodMSStr)
		if period > 0 {
			pumpPeriod = time.Millisecond * time.Duration(period)
		}
	}

	tokensecret := ""
	if len(tgtoken) > 10 {
		tokensecret += tgtoken[:5]
		tokensecret += "..."
		tokensecret += tgtoken[len(tgtoken)-5 : len(tgtoken)]
	}

	logger.WithFields(logrus.Fields{
		"natsaddr": natsaddr,
		"pubname":  pubname,
		"loglevel": loglevel,

		"tg_token": tokensecret,
	}).Infoln("init with settings")
}

func main() {
	var (
		err error
	)

	// ---------------------------
	// connect to nats
	// ---------------------------

	nc, err = nats.Connect(natsaddr,
		nats.Name(PID),
		// TODO: more configs
	)
	if err != nil {
		logger.WithError(err).
			Fatalln("connect to nats server")
	}
	defer nc.Close()

	// ---------------------------
	// Telegram bot config
	// ---------------------------

	bot, err := tgbotapi.NewBotAPI(tgtoken)
	if err != nil {
		logger.WithError(err).Fatal("init telegram bot")
	}

	if logrus.GetLevel() >= logrus.DebugLevel {
		bot.Debug = true
	}

	cfg := tgbotapi.NewUpdate(0)
	cfg.Timeout = 60

	// ---------------------------
	// run pumper
	// ---------------------------

	go func() {
		for {
			select {
			case <-time.After(pumpPeriod):
				if err := getUpdates(&cfg); err != nil {
					logger.WithError(err).
						Error("request get updates")
					time.Sleep(sleepTimeout)
				}
			}
		}
	}()

	// ---------------------------
	// run listener of OS
	// ---------------------------

	osSignal := make(chan os.Signal, 2)
	close := make(chan struct{})
	signal.Notify(osSignal, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-osSignal
		logger.Infoln("signal completion of the process")

		nc.Close()
		close <- struct{}{}
	}()
	<-close

	os.Exit(0)
}

func getUpdates(cfg *tgbotapi.UpdateConfig) error {
	var countRetry = 0

RETRY:
	if countRetry >= countMaxRetry {
		return fmt.Errorf("error get updates, max count retry %d", countRetry)
	}

	apiurl := fmt.Sprintf(TGAPIEntrypointFormat, tgtoken, "getUpdates")

	// options

	api, err := url.Parse(apiurl)
	query := api.Query()
	if cfg.Offset != 0 {
		query.Add("offset", strconv.Itoa(cfg.Offset))
	}
	if cfg.Limit > 0 {
		query.Add("limit", strconv.Itoa(cfg.Limit))
	}
	if cfg.Timeout > 0 {
		query.Add("timeout", strconv.Itoa(cfg.Timeout))
	}
	api.RawQuery = query.Encode()

	// request

	req, err := http.NewRequest("POST", api.String(), nil)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		logger.WithError(err).Error("request, fail request")
		countRetry++
		time.Sleep(sleepTimeout)
		goto RETRY
	}
	defer resp.Body.Close()

	// Note: array of updates

	if resp.StatusCode != http.StatusOK {

		logger.WithFields(logrus.Fields{
			"status": resp.Status,
		}).Errorln("request, bad request")

		return fmt.Errorf("bad request %q", resp.Status)
	}

	bytes, _ := ioutil.ReadAll(resp.Body)

	updates := struct {
		Result []json.RawMessage `json:"result"`
	}{}
	json.Unmarshal(bytes, &updates)

	logger.WithField("data", string(bytes)).
		Debug("update data")

	for _, updatedata := range updates.Result {
		updateID := getUpdateID(updatedata)

		if updateID >= cfg.Offset {
			cfg.Offset = updateID + 1

			if err := nc.Publish(pubname, updatedata); err != nil {
				logger.WithFields(logrus.Fields{
					"err":        err,
					"updatedata": string(updatedata),
				}).Errorln("published")
				continue
			}
		}

		logger.WithField("updatedata", string(updatedata)).Debug("published")
	}

	return nil
}

func getUpdateID(raw []byte) int {
	v := struct {
		UpdateID int `json:"update_id"`
	}{}
	json.Unmarshal(raw, &v)
	return v.UpdateID
}
