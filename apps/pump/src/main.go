package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
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

	cfg := tgbotapi.NewUpdate(0)
	cfg.Timeout = 60

	go func() {
		for {
			select {
			case <-time.After(time.Millisecond * 1):
				if err := getUpdates(nc, pubname, tgToken); err != nil {
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

func getUpdates(nc *nats.Conn, pubname, token string) error {
	var countRetry = 0
	var maxNumberAttempts = 5

RETRY:
	if countRetry >= maxNumberAttempts {
		return fmt.Errorf("error get updates, number of attempts %d", countRetry)
	}

	apiurl := fmt.Sprintf(APIEntrypoint, token, "getUpdates")

	req, err := http.NewRequest("POST", apiurl, nil)
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

	updates := []json.RawMessage{}
	json.Unmarshal(bytes, &updates)

	for _, update := range updates {
		if err := nc.Publish(pubname, update); err != nil {
			logrus.WithFields(logrus.Fields{
				"_ref": NAME,

				"err": err,
			}).Error("publish")
		}

		// TODO: if invalid connection then complete the procedure
	}

	// TODO: if error then to save the not processed messages

	return nil
}
