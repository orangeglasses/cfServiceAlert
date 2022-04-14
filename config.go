package main

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/kelseyhightower/envconfig"
)

type alertServerConfig struct {
	CFUser                      string `envconfig:"cf_user" required:"true"`
	CFPassword                  string `envconfig:"cf_password" required:"true"`
	CFClient                    string `envconfig:"cf_client"`
	CFSecret                    string `envconfig:"cf_secret"`
	RulesPath                   string `envconfig:"rules_path" default:"rules.json"`
	PromURL                     string `envconfig:"prometheus_url" required:"true"`
	GrafanaApiKey               string `envconfig:"grafana_api_key" required:"true"`
	CheckInterval               int    `envconfig:"check_interval" default:"120"`
	Environment                 string `envconfig:"environment_name" required:"true"`
	NotificationServiceUrl      string `envconfig:"notification_url" required:"true"`
	NotificationServiceUser     string `envconfig:"notification_user" required:"true"`
	NotificaitonServicePassword string `envconfig:"notification_password" required:"true"`
}

func alertServerConfigLoad() (alertServerConfig, alertRules, error) {
	var config alertServerConfig
	rules := make(alertRules)

	err := envconfig.Process("", &config)
	if err != nil {
		return alertServerConfig{}, nil, err
	}

	if config.CFUser == "" && config.CFClient == "" {
		log.Fatal("Please set CF_USER/CF_PASSWORD or CF_CLIENT/CF_SECRET")
	}

	if config.CFUser != "" && config.CFClient != "" {
		log.Println("Both CF_USER and CF_CLIENT are set. I'll use CF_CLIENT and ignore CF_USER.")
	}

	inBuf, err := ioutil.ReadFile(config.RulesPath)
	if err != nil {
		return alertServerConfig{}, nil, err
	}

	err = json.Unmarshal(inBuf, &rules)
	if err != nil {
		return alertServerConfig{}, nil, err
	}

	return config, rules, nil
}
