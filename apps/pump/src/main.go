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

var NAME = "pump"
var VERSION = "0.0.1"
var APIEntrypoint = "https://api.telegram.org/bot%s/%s"

func main() {
	natsAdds := os.Getenv("NATSADDR")
	clientID := os.Getenv("CLIENTID")
	logLevelStr := os.Getenv("LOGLEVEL")
	hostname := os.Getenv("HOSTNAME")
	pubname := os.Getenv("PUBNAME")
	tgToken := os.Getenv("TGTOKEN")

	logrus.WithFields(logrus.Fields{
		"_ref":         NAME,
		"_host":        hostname,
		"_nats_client": clientID,

		"nats_addr":    natsAdds,
		"version":      VERSION,
		"natsClientID": clientID,
		"logLevel":     logLevelStr,
		"tgtoken":      len(tgToken),
		"pubname":      pubname,
	}).Infoln(NAME + " init")

	// Setup logger
	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		logLevel = logrus.ErrorLevel
	}
	logrus.SetLevel(logLevel)

	// Setup nats
	nc, err := nats.Connect(natsAdds,
		nats.Name(clientID+"_"+hostname),
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"_ref":         NAME,
			"_host":        hostname,
			"_nats_client": clientID,

			"err": err,
		}).Fatal("connect to nats server")
	}

	// Init bot
	bot, err := tgbotapi.NewBotAPI(tgToken)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"_ref": NAME,
			"err":  err,
		}).Fatal("Init telegram bot")
	}

	if logrus.GetLevel() >= logrus.DebugLevel {
		bot.Debug = true
	}

	cfg := tgbotapi.NewUpdate(0) // TODO: save last updateID
	cfg.Timeout = 60

	go func() {
		for {
			select {
			case <-time.After(time.Millisecond * 1):
				if err := getUpdates(nc, &cfg, pubname, tgToken); err != nil {
					logrus.WithError(err).Error("request get updates, invalid status")
					time.Sleep(time.Second * 3)
				}
			}
		}
	}()

	signalChan := make(chan os.Signal, 2)
	done := make(chan bool)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		logrus.WithFields(logrus.Fields{
			"_ref":  NAME,
			"_host": hostname,
		}).Info("signal completion of the process")

		nc.Close()
		done <- true
	}()
	<-done

	os.Exit(0)
}

func getUpdates(nc *nats.Conn, cfg *tgbotapi.UpdateConfig, pubname, token string) error {
	var countRetry = 0
	var maxNumberAttempts = 5

RETRY:
	if countRetry >= maxNumberAttempts {
		return fmt.Errorf("error get updates, number of attempts %d", countRetry)
	}

	apiurl := fmt.Sprintf(APIEntrypoint, token, "getUpdates")

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

	req, err := http.NewRequest("POST", api.String(), nil)
	req.Header.Set("Content-Type", "application/json")

	// TODO: reuse http client
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		logrus.WithError(err).Error("request get updates, invalid status")
		countRetry++
		time.Sleep(time.Second * 3)
		goto RETRY
	}
	defer resp.Body.Close()

	// Note: array of updates

	if resp.StatusCode != http.StatusOK {

		if err != nil {
			logrus.WithError(err).Error("request get updates, invalid status")
			countRetry++
			time.Sleep(time.Second * 3)
			goto RETRY
		}

		return fmt.Errorf("error get updates, bad request %q", resp.Status)
	}

	bytes, _ := ioutil.ReadAll(resp.Body)

	updates := struct {
		Result []json.RawMessage `json:"result"`
	}{}
	json.Unmarshal(bytes, &updates)

	logrus.WithFields(logrus.Fields{
		"_ref": NAME,
		"data": string(bytes),
	}).Debug("get update")

	for _, update := range updates.Result {
		updateID := getUpdateID(update)

		if updateID >= cfg.Offset {
			cfg.Offset = updateID + 1

			if err := nc.Publish(pubname, update); err != nil {
				logrus.WithFields(logrus.Fields{
					"_ref": NAME,
					"err":  err,
				}).Error("publish")
			}
		}

		logrus.WithFields(logrus.Fields{
			"_ref": NAME,
		}).Debug("published")

		// TODO: if invalid connection then complete the procedure

	}

	// TODO: if error then to save the not processed messages

	return nil
}

func getUpdateID(raw []byte) int {
	v := struct {
		UpdateID int `json:"update_id"`
	}{}
	json.Unmarshal(raw, &v)
	return v.UpdateID
}
