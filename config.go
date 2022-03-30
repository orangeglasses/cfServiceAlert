package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/kelseyhightower/envconfig"
)

type alertServiceConfig struct {
	CFUser        string `envconfig:"cf_user" required:"true"`
	CFPassword    string `envconfig:"cf_password" required:"true"`
	RulesPath     string `envconfig:"rules_path" default:"rules.json"`
	PromURL       string `envconfig:"prometheus_url" required:"true"`
	GrafanaApiKey string `envconfig:"grafana_api_key" required:"true"`
}

type alertRule struct {
	Promq    string `json:"prometheus_query"`
	Treshold string `json:"treshold"`
	Above    bool   `json:"above"`
}

type alertRules map[string]alertRule

func alertServiceConfigLoad() (alertServiceConfig, alertRules, error) {
	var config alertServiceConfig
	rules := make(alertRules)

	err := envconfig.Process("", &config)
	if err != nil {
		return alertServiceConfig{}, nil, err
	}

	inBuf, err := ioutil.ReadFile(config.RulesPath)
	if err != nil {
		return alertServiceConfig{}, nil, err
	}

	err = json.Unmarshal(inBuf, &rules)
	if err != nil {
		return alertServiceConfig{}, nil, err
	}

	return config, rules, nil
}
