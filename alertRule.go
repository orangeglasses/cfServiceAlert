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
	Name           string `json:"name"`
	Promq          string `json:"prometheus_query"`
	Treshold       string `json:"treshold"`
	NotifyInterval string `json:"notification_interval"`
	Above          bool   `json:"above"`
	Subject        string `json:"subject"`
	Message        string `json:"message"`
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
			exceeded, err := rule.TresholdExceeded(*sample)
			if err != nil {
				log.Println("Error checking treshold: ", err)
				continue
			}

			if exceeded {
				msg, err := rule.GenerateMessageForSpace(*a.cfClient, serviceInstance, sample.Value, a.environment)
				if err != nil {
					log.Println("Error generating notification: ", err)
				}

				err = a.notificationSerivceClient.Send(msg)
				if err != nil {
					log.Println("Error sending notificaiton: ", err)
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

func (rule *alertRule) GenerateMessageForSpace(client cfclient.Client, serviceInstance cfclient.V3ServiceInstance, sampleValue model.SampleValue, environment string) (NotificationMessage, error) {
	space, err := client.GetSpaceByGuid(serviceInstance.Relationships["space"].Data.GUID)
	if err != nil {
		return NotificationMessage{}, err
	}

	org, err := space.Org()
	if err != nil {
		return NotificationMessage{}, err
	}

	floatMetricValue, _ := strconv.ParseFloat(sampleValue.String(), 32)

	templData := struct {
		AlertName       string
		InstanceId      string
		InstanceName    string
		EnvironmentName string
		SpaceName       string
		OrgName         string
		Treshold        string
		MetricValue     string
	}{
		AlertName:       rule.Name,
		InstanceId:      serviceInstance.Guid,
		InstanceName:    serviceInstance.Name,
		EnvironmentName: environment,
		SpaceName:       space.Name,
		OrgName:         org.Name,
		Treshold:        rule.Treshold,
		MetricValue:     fmt.Sprintf("%.2f", floatMetricValue),
	}

	log.Printf("Generating notification for service %s in space: %s(%s)\n", serviceInstance.Name, space.Name, space.Guid)

	var renderedMessage bytes.Buffer
	msgTmpl, err := template.New("msg").Parse(rule.Message)
	if err := msgTmpl.Execute(&renderedMessage, templData); err != nil {
		return NotificationMessage{}, fmt.Errorf("Error rendering message: %v", err)
	}

	var renderedSubject bytes.Buffer
	subjTmpl, err := template.New("subject").Parse(rule.Subject)
	if err := subjTmpl.Execute(&renderedSubject, templData); err != nil {
		return NotificationMessage{}, fmt.Errorf("Error rendering subject: %v", err)
	}

	msg := NotificationMessage{
		Id:        fmt.Sprintf("%s-%s\n", serviceInstance.Guid, rule.Name),
		Subject:   renderedSubject.String(),
		Message:   renderedMessage.String(),
		ExpiresIn: rule.NotifyInterval,
		Target: NotificationMessageTarget{
			Type:        "space",
			Environment: environment,
			Id:          space.Guid,
		},
	}

	return msg, nil
}
