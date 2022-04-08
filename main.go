package main

import (
	"fmt"
	"log"
	"net/http"
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
		Username:   config.CFUser,
		Password:   config.CFPassword,
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
		notificationSerivceClient: *NewNotificationServiceClient(config.NotificationServiceUrl, config.NotificationServiceUser, config.NotificaitonServicePassword),
	}

	as.Start(int64(config.CheckInterval))

	http.HandleFunc("/status", as.statusHandler)
	http.ListenAndServe(fmt.Sprintf(":%v", appEnv.Port), nil)
}
