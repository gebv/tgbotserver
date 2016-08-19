package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/nats-io/nats"
)

var (
	NAME    = "listener"
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

	listenerAddr = os.Getenv("LISTENADDR")

	logger *logrus.Entry
	nc     *nats.Conn
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

	logger.WithFields(logrus.Fields{
		"natsaddr":   natsaddr,
		"pubname":    pubname,
		"listenaddr": listenerAddr,
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
	// routes
	// ---------------------------

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		updatedata, err := ioutil.ReadAll(req.Body) // TODO: io.Copy

		if err != nil {
			logger.WithError(err).Errorln("read request")
		}

		// Note: one update

		if err := nc.Publish(pubname, updatedata); err != nil {
			logger.WithFields(logrus.Fields{
				"err":        err,
				"updatedata": string(updatedata),
			}).Errorln("published")
			return
		}

		logger.WithField("updatedata", string(updatedata)).Debug("published")
	})

	logger.Infoln("run http listener")

	// ---------------------------
	// run listener of HTTP
	// ---------------------------

	go func() {
		if err := http.ListenAndServe(listenerAddr, mux); err != nil {
			logger.WithError(err).Fatal("http listener")
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
