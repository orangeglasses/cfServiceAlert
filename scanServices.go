package main

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"text/template"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/common/model"
	"github.com/vedhavyas/hashring"
)

func (a *alertServer) scanServices() {
	app, _ := a.cfClient.GetAppByGuid(a.appGuid)
	a.nodes = app.Instances

	ring := hashring.New(3, nil)
	for i := a.nodes; i > 0; i-- {
		ring.Add(fmt.Sprintf("%v", i-1))
	}

	serviceInstances, _ := a.cfClient.ListV3ServiceInstances()
	log.Printf("Processing %v instances\n", len(serviceInstances))

	for _, serviceInstance := range serviceInstances {
		n, _ := ring.Locate(serviceInstance.Guid)
		if n == a.node {
			//log.Println("processing servive guid: ", serviceInstance.Guid)

			if serviceInstance.Relationships["service_plan"].Data.GUID == "" {
				//this is probably a CUPS, skip it.
				continue
			}

			servicePlan, err := a.cfClient.GetServicePlanByGUID(serviceInstance.Relationships["service_plan"].Data.GUID)
			if err != nil {
				continue
			}
			service, err := a.cfClient.GetServiceByGuid(servicePlan.ServiceGuid)
			if err != nil {
				log.Println("Error getting Service: ", err)
			}

			if rule, ok := a.alertRules[service.Label]; ok {
				log.Printf("Checking %v service with guid: %v\n", service.Label, serviceInstance.Guid)
				vres, err := a.GetMetric(rule.Promq, serviceInstance.Guid)
				if err != nil {
					log.Println(err)
					continue
				}

				for _, sample := range vres {
					fmt.Println("Value: ", sample.Value)
					fmt.Printf("Metric name: %+v\n", sample.Metric["__name__"])
					fmt.Printf("Metric: %+v\n", sample.Metric)

					exceeded, err := rule.TresholdExceeded(*sample)
					if err != nil {
						log.Println("Error checking treshold: ", err)
						continue
					}

					if exceeded {
						err := rule.SendNotificationForSpace(*a.cfClient, serviceInstance, fmt.Sprintf("%v", sample.Metric["__name__"]))
						if err != nil {
							log.Println("Error sending alert: ", err)
						}
					}
				}
			}
		}
	}
}

func (a *alertServer) GetMetric(queryTemplate, instanceId string) (model.Vector, error) {
	var renderedQuery bytes.Buffer
	templData := struct {
		InstanceId string
	}{InstanceId: instanceId}

	t, err := template.New("pq").Parse(queryTemplate)
	if err := t.Execute(&renderedQuery, templData); err != nil {
		log.Printf("Error rendering prometheus query: %v", err)

	}

	res, err := a.promClient.Query(renderedQuery.String())
	if err != nil {
		return nil, fmt.Errorf("Error querying prometheus: %v\n", err.Error())
	}

	if res.Type().String() != "vector" {
		return nil, fmt.Errorf("Prometheus query did not return a vector")
	}

	return res.(model.Vector), nil
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

func (rule *alertRule) SendNotificationForSpace(client cfclient.Client, service cfclient.V3ServiceInstance, metric string) error {
	space, err := client.GetSpaceByGuid(service.Relationships["space"].Data.GUID)
	if err != nil {
		return err
	}

	log.Printf("Generating notification for service %s, metric: %s, space: %s(%s)\n", service.Name, metric, space.Name, space.Guid)

	return nil
}
