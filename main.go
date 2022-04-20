package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/cloudfoundry-community/go-cfenv"
)

func main() {
	appEnv, _ := cfenv.Current()
	config, rules, err := alertServerConfigLoad()
	if err != nil {
		log.Fatalf("Error loading config: %v", err.Error())
	}

	c := &cfclient.Config{
		ApiAddress: appEnv.CFAPI,
	}

	if config.CFUser != "" {
		c.Username = config.CFUser
		c.Password = config.CFPassword
	} else {
		c.ClientID = config.CFClient
		c.ClientSecret = config.CFSecret
	}

	cfClient, err := cfclient.NewClient(c)
	if err != nil {
		log.Fatal("Failed logging into cloudfoundry", err)
	}

	promClient, err := NewPrometheusClient(config.PromURL, config.GrafanaApiKey)
	if err != nil {
		log.Fatal("Error creating Prometheus client", err)
	}

	as := &alertServer{
		cfClient:                  cfClient,
		promClient:                promClient,
		appGuid:                   appEnv.AppID,
		node:                      strconv.Itoa(appEnv.Index),
		alertRules:                rules,
		environment:               config.Environment,
		notificationSerivceClient: *NewNotificationServiceClient(config.NotificationServiceUrl, config.NotificationServiceUser, config.NotificationServicePassword),
	}

	as.Start(int64(config.CheckInterval))

	signals := make(chan os.Signal, 2)
	signal.Notify(signals, os.Interrupt)
	signal.Notify(signals, os.Kill)

	<-signals
}
