package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/cloudfoundry-community/go-cfenv"
)

type alertServer struct {
	cfClient   *cfclient.Client
	promClient *PrometheusClient
	appGuid    string
	node       string
	nodes      int
	alertRules alertRules
}

func (a *alertServer) statusHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "I am Node %v of %v", a.node, a.nodes)
}

func main() {
	appEnv, _ := cfenv.Current()
	config, rules, err := alertServiceConfigLoad()
	if err != nil {
		log.Fatalf("Error loading config: %v", err.Error())
	}

	c := &cfclient.Config{
		ApiAddress: appEnv.CFAPI,
		Username:   config.CFUser,
		Password:   config.CFPassword,
	}

	fmt.Printf("%+v", c)

	cfClient, err := cfclient.NewClient(c)
	if err != nil {
		log.Fatal("Failed logging into cloudfoundry", err)
	}

	promClient, err := NewPrometheusClient(config.PromURL, config.GrafanaApiKey)
	if err != nil {
		log.Fatal("Error creating Prometheus client", err)
	}

	as := &alertServer{
		cfClient:   cfClient,
		promClient: promClient,
		appGuid:    appEnv.AppID,
		node:       strconv.Itoa(appEnv.Index),
		nodes:      0,
		alertRules: rules,
	}

	ticker := time.NewTicker(time.Second * 15)
	go func() {
		for range ticker.C {
			as.scanServices()
		}
	}()

	http.HandleFunc("/status", as.statusHandler)
	http.ListenAndServe(fmt.Sprintf(":%v", appEnv.Port), nil)
}
