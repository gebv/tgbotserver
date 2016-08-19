package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/nats-io/nats"
	"gopkg.in/pg.v4"
	"gopkg.in/telegram-bot-api.v4"
	// set your APPLICATION
	app "apps/checkerbot"
)

var (
	NAME     = "worker"
	VERSION  = "0.0.2"
	hostname = os.Getenv("HOSTNAME")
	appname  = os.Getenv("APPNAME")
	loglevel = os.Getenv("LOGLEVEL")
	PID      = NAME + "_" + VERSION + "_" + appname + "_" + hostname

	natsaddr = os.Getenv("NATSADDR")
	subname  = os.Getenv("SUBNAME")

	tgtoken = os.Getenv("TG_TOKEN")
	dbaddr  = os.Getenv("DB_ADDR")
	dbname  = os.Getenv("DB_NAME")
	dbuser  = os.Getenv("DB_USER")
	dbpass  = os.Getenv("DB_PASS")

	logger *logrus.Entry
	tgapi  *tgbotapi.BotAPI
	nc     *nats.Conn
	db     *pg.DB

	appCreateSchema AppDBSchemaCreator
	appEntryPoint   AppEntryPoint
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

	tokensecret := ""
	if len(tgtoken) > 10 {
		tokensecret += tgtoken[:5]
		tokensecret += "..."
		tokensecret += tgtoken[len(tgtoken)-5 : len(tgtoken)]
	}

	logger.WithFields(logrus.Fields{
		"natsaddr": natsaddr,
		"subname":  subname,
		"loglevel": loglevel,

		"tg_token": tokensecret,
		"db_name":  dbname,
		"db_user":  dbuser,
	}).Infoln("init with settings")
}

func main() {
	var (
		err error
	)

	// ---------------------------
	// Init application
	// ---------------------------

	appCreateSchema = app.CreateSchema
	appEntryPoint = app.Execute

	// ---------------------------
	// Init telegram bot
	// ---------------------------

	tgapi, err = tgbotapi.NewBotAPI(tgtoken)

	if err != nil {
		logger.WithError(err).Fatalln("init telegram api")
	}

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
	// setup database
	// ---------------------------

	if err := setupDatabase(dbuser, dbpass, dbaddr, dbname); err != nil {
		logger.WithError(err).
			Fatalln("connect and create schema to database")
	}

	// ---------------------------
	// Subscribe
	// ---------------------------

	mcb := func(m *nats.Msg) {
		logger.WithField("data", string(m.Data)).Debug("received a message")

		if err := Execute(m.Data, time.Second*1); err != nil {
			logger.WithError(err).Error("handler")
		}
	}

	sub, err := nc.QueueSubscribe(subname, appname, mcb)

	if err != nil {
		sub.Unsubscribe()
		nc.Close()
		logger.WithError(err).Fatalln("subscribe")
	}

	// ---------------------------
	// run listener of OS
	// ---------------------------

	osSignal := make(chan os.Signal, 2)
	close := make(chan struct{})
	signal.Notify(osSignal, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-osSignal
		logger.Infoln("signal completion of the process")

		sub.Unsubscribe()
		nc.Close()
		close <- struct{}{}
	}()
	<-close

	os.Exit(0)
}
