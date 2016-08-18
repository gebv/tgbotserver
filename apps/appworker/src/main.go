package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/nats-io/nats"
	"gopkg.in/telegram-bot-api.v4"
)

var (
	NAME    = "appworker"
	VERSION = "0.0.1"

	dbAddr    string
	dbName    string
	dbUser    string
	dbPwd     string
	dbNetwork string
	hostname  string
	clientID  string

	api *tgbotapi.BotAPI
)

func main() {
	natsAdds := os.Getenv("NATSADDR")
	clientID = os.Getenv("CLIENTID")
	logLevelStr := os.Getenv("LOGLEVEL")
	listenerAddr := os.Getenv("ADDRESS")
	hostname = os.Getenv("HOSTNAME")
	subname := os.Getenv("SUBNAME")
	tgToken := os.Getenv("TGTOKEN")
	dbAddr = os.Getenv("DBADDR")
	dbName = os.Getenv("DBNAME")
	dbUser = os.Getenv("DBUSER")
	dbPwd = os.Getenv("DBPASS")
	dbNetwork = os.Getenv("DBNETWORK")

	logrus.WithFields(logrus.Fields{
		"_ref":         NAME,
		"_host":        hostname,
		"_nats_client": clientID,

		"nats_addr":    natsAdds,
		"version":      VERSION,
		"natsClientID": clientID,
		"logLevel":     logLevelStr,
		"addr":         listenerAddr,
		"subname":      subname,
		"tgtoken":      len(tgToken),
	}).Infoln("httplistener init")

	// setup logger

	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		logLevel = logrus.ErrorLevel
	}
	logrus.SetLevel(logLevel)

	// setup database

	if err := setupDatabase(dbNetwork, dbAddr, dbName, dbUser, dbPwd); err != nil {
		logrus.WithFields(logrus.Fields{
			"_ref":         NAME,
			"_host":        hostname,
			"_nats_client": clientID,

			"err": err,
		}).Fatal("setup database")
	}

	if err := createSchema(db); err != nil {
		logrus.WithFields(logrus.Fields{
			"_ref":         NAME,
			"_host":        hostname,
			"_nats_client": clientID,

			"err": err,
		}).Fatal("create schema database")
	}

	// setup NATS

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

	mcb := func(m *nats.Msg) {
		logrus.WithFields(logrus.Fields{
			"_ref":         NAME,
			"_host":        hostname,
			"_nats_client": clientID,

			"msg": string(m.Data),
		}).Debug("Received a message")

		if err := Handler(m.Data, time.Second*1); err != nil {
			logrus.WithFields(logrus.Fields{
				"_ref":         NAME,
				"_host":        hostname,
				"_nats_client": clientID,

				"err": err,
			}).Error("handler")
		}
	}

	// init api

	api, err = tgbotapi.NewBotAPI(tgToken)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"_ref":         NAME,
			"_host":        hostname,
			"_nats_client": clientID,

			"err": err,
		}).Fatal("init telegram api")
	}

	// handler started

	sub, err := nc.QueueSubscribe(subname, clientID, mcb)

	if err != nil {
		sub.Unsubscribe()
		nc.Close()
		logrus.WithFields(logrus.Fields{
			"_ref":         NAME,
			"_host":        hostname,
			"_nats_client": clientID,

			"err":     err,
			"subname": subname,
		}).Fatal("Subscribe")
	}

	signalChan := make(chan os.Signal, 2)
	done := make(chan bool)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan

		logrus.WithFields(logrus.Fields{
			"_ref":    NAME,
			"_host":   hostname,
			"_client": clientID,
		}).Info("signal completion of the process")

		sub.Unsubscribe()
		nc.Close()
		done <- true
	}()
	<-done

	os.Exit(0)
}
