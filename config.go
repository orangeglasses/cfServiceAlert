package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/kelseyhightower/envconfig"
)

type alertServerConfig struct {
	CFUser                      string `envconfig:"cf_user" required:"true"`
	CFPassword                  string `envconfig:"cf_password" required:"true"`
	RulesPath                   string `envconfig:"rules_path" default:"rules.json"`
	PromURL                     string `envconfig:"prometheus_url" required:"true"`
	GrafanaApiKey               string `envconfig:"grafana_api_key" required:"true"`
	CheckInterval               int    `envconfig:"check_interval" default:"60"`
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
