package main

import (
	"os"
	"os/signal"

	"github.com/Sirupsen/logrus"
	"github.com/nats-io/nats"
)

var NAME = "appworker"
var VERSION = "0.0.1"

func main() {
	natsAdds := os.Getenv("NATSADDR")
	clientID := os.Getenv("CLIENTID")
	logLevelStr := os.Getenv("LOGLEVEL")
	listenerAddr := os.Getenv("ADDRESS")
	hostname := os.Getenv("HOSTNAME")
	subname := os.Getenv("SUBNAME")
	tgToken := os.Getenv("TGTOKEN")

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

	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		logLevel = logrus.ErrorLevel
	}
	logrus.SetLevel(logLevel)

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
	}

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

	// Wait for a SIGINT (perhaps triggered by user with CTRL-C)
	// Run cleanup when signal is received
	signalChan := make(chan os.Signal, 1)
	done := make(chan bool)
	signal.Notify(signalChan, os.Interrupt)

	go func() {
		<-signalChan
		sub.Unsubscribe()
		nc.Close()
		done <- true
	}()
	<-done

	os.Exit(0)
}
