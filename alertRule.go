package main

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"text/template"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/common/model"
)

type alertRule struct {
	Name     string `json:"name"`
	Promq    string `json:"prometheus_query"`
	Treshold string `json:"treshold"`
	Above    bool   `json:"above"`
	Message  string `json:"message"`
}

type alertRuleSet []alertRule

type alertRules map[string]alertRuleSet

func (rs alertRuleSet) Process(a *alertServer, serviceInstance cfclient.V3ServiceInstance) {
	for _, rule := range rs {
		vres, err := a.GetMetric(rule.Promq, serviceInstance.Guid)
		if err != nil {
			log.Println(err)
			continue
		}

		for _, sample := range vres {
			fmt.Println("Value: ", sample.Value)

			exceeded, err := rule.TresholdExceeded(*sample)
			if err != nil {
				log.Println("Error checking treshold: ", err)
				continue
			}

			if exceeded {
				err := rule.SendNotificationForSpace(*a.cfClient, serviceInstance)
				if err != nil {
					log.Println("Error sending alert: ", err)
				}
			}
		}
	}
}

func (rule *alertRule) TresholdExceeded(sample model.Sample) (bool, error) {
	treshold, err := strconv.Atoi(rule.Treshold)
	if err != nil {
		return false, err
	}

	if rule.Above {
		if int(sample.Value) > treshold {
			//log.Println("treshold exceeded (higher)")
			return true, nil
		}
	} else {
		if int(sample.Value) < treshold {
			//log.Println("treshold exceeded (lower)")
			return true, nil
		}
	}

	return false, nil
}

func (rule *alertRule) SendNotificationForSpace(client cfclient.Client, serviceInstance cfclient.V3ServiceInstance) error {
	space, err := client.GetSpaceByGuid(serviceInstance.Relationships["space"].Data.GUID)
	if err != nil {
		return err
	}

	org, err := space.Org()
	if err != nil {
		return err
	}

	var renderedMessage bytes.Buffer
	templData := struct {
		InstanceId   string
		InstanceName string
		SpaceName    string
		OrgName      string
	}{
		InstanceId:   serviceInstance.Guid,
		InstanceName: serviceInstance.Name,
		SpaceName:    space.Name,
		OrgName:      org.Name,
	}

	t, err := template.New("msg").Parse(rule.Message)
	if err := t.Execute(&renderedMessage, templData); err != nil {
		return fmt.Errorf("Error rendering message: %v", err)
	}

	log.Printf("Generating notification for service %s in space: %s(%s)\n", serviceInstance.Name, space.Name, space.Guid)
	log.Println(renderedMessage.String())

	return nil
}
