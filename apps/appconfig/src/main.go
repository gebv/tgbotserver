package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sirupsen/logrus"
)

var NAME = "appconfig"
var VERSION = "0.0.1"
var APIEntrypoint = "https://api.telegram.org/bot%s/%s"

func main() {
	webhook := os.Getenv("TGWEBHOOK")
	token := os.Getenv("TGTOKEN")
	hostname := os.Getenv("HOSTNAME")
	logLevelStr := os.Getenv("LOGLEVEL")

	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		logLevel = logrus.ErrorLevel
	}
	logrus.SetLevel(logLevel)

	logrus.WithFields(logrus.Fields{
		"_ref":  NAME,
		"_host": hostname,

		"version":  VERSION,
		"logLevel": logLevelStr,
		"webhook":  webhook,
		"tgtoken":  len(token),
	}).Infoln("appconfig init")

	if err := setWebhook(token, webhook); err != nil {
		logrus.WithFields(logrus.Fields{
			"_ref": NAME,
			"err":  err,
		}).Fatal("set webhook")
	}

	signalChan := make(chan os.Signal, 2)
	done := make(chan bool)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		logrus.WithFields(logrus.Fields{
			"_ref":  NAME,
			"_host": hostname,
		}).Info("signal completion of the process")

		if err := setWebhook(token, ""); err != nil {
			logrus.WithFields(logrus.Fields{
				"_ref": NAME,
				"err":  err,
			}).Fatal("clear webhook")
		}
		done <- true
	}()
	<-done

	os.Exit(0)

}

func setWebhook(token, webhook string) error {
	apiurl := fmt.Sprintf(APIEntrypoint, token, "setWebhook")

	api, err := url.Parse(apiurl)
	query := api.Query()
	query.Add("url", webhook)
	api.RawQuery = query.Encode()

	req, err := http.NewRequest("POST", api.String(), nil)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {

		return err
	}
	defer resp.Body.Close()

	bytes, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {

		logrus.WithFields(logrus.Fields{
			"_ref":        NAME,
			"err":         err,
			"url_request": api.String(),
			"body":        string(bytes),
			"webhook":     webhook,
			"status":      resp.Status,
		}).Error("fail updated webhook")

		return err
	}

	logrus.WithFields(logrus.Fields{
		"_ref":        NAME,
		"webhook":     webhook,
		"url_request": api.String(),
		"body":        string(bytes),
	}).Info("successfully updated webhook")

	return nil
}
