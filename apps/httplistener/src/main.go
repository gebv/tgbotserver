package main

import (
	"io/ioutil"
	"net/http"
	"os/signal"

	"os"

	"github.com/Sirupsen/logrus"
	"github.com/nats-io/nats"
)

var NAME = "httplistener"
var VERSION = "0.0.1"

func main() {
	natsAdds := os.Getenv("NATSADDR")
	clientID := os.Getenv("CLIENTID")
	logLevelStr := os.Getenv("LOGLEVEL")
	listenerAddr := os.Getenv("LISTENADDR")
	hostname := os.Getenv("HOSTNAME")
	pubname := os.Getenv("PUBNAME")

	logrus.WithFields(logrus.Fields{
		"_ref":         NAME,
		"_host":        hostname,
		"_nats_client": clientID,

		"nats_addr":    natsAdds,
		"version":      VERSION,
		"natsClientID": clientID,
		"logLevel":     logLevelStr,
		"addr":         listenerAddr,
		"pubname":      pubname,
	}).Infoln("httplistener init")

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

	defer nc.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		body, err := ioutil.ReadAll(req.Body)

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"_ref":         NAME,
				"_host":        hostname,
				"_nats_client": clientID,

				"err": err,
			}).Error("read request")
		}

		logrus.WithFields(logrus.Fields{
			"_ref":         NAME,
			"_host":        hostname,
			"_nats_client": clientID,

			"src_body": string(body),
		}).Debug("received a data")

		if err := nc.Publish(pubname, body); err != nil {
			logrus.WithFields(logrus.Fields{
				"_ref":         NAME,
				"_host":        hostname,
				"_nats_client": clientID,

				"err": err,
			}).Error("publish")
		}
	})

	logrus.WithFields(logrus.Fields{
		"_ref":         NAME,
		"_host":        hostname,
		"_nats_client": clientID,
	}).Infoln("Run http listener")

	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		logLevel = logrus.ErrorLevel
	}
	logrus.SetLevel(logLevel)

	go func() {
		if err := http.ListenAndServe(listenerAddr, mux); err != nil {
			logrus.WithFields(logrus.Fields{
				"_ref":         NAME,
				"_host":        hostname,
				"_nats_client": clientID,

				"err": err,
			}).Fatal("listener")
		}
	}()

	// Wait for a SIGINT (perhaps triggered by user with CTRL-C)
	// Run cleanup when signal is received
	signalChan := make(chan os.Signal, 1)
	done := make(chan bool)
	signal.Notify(signalChan, os.Interrupt)

	go func() {
		<-signalChan
		nc.Close()
		done <- true
	}()
	<-done

	os.Exit(0)
}
