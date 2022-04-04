package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/prometheus/common/model"
	"github.com/vedhavyas/hashring"
)

type alertServer struct {
	cfClient   *cfclient.Client
	promClient *PrometheusClient
	appGuid    string
	node       string
	nodes      int
	alertRules alertRules
}

func (a *alertServer) Start(checkInterval int64) {
	ticker := time.NewTicker(time.Second * time.Duration(checkInterval))
	go func() {
		for range ticker.C {
			a.scanServices()
		}
	}()
}

func (a *alertServer) statusHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "I am Node %v of %v", a.node, a.nodes)
}

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

			if ruleSet, ok := a.alertRules[service.Label]; ok {
				log.Printf("Checking %v service with guid: %v\n", service.Label, serviceInstance.Guid)
				ruleSet.Process(a, serviceInstance)
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
		return nil, fmt.Errorf("Error rendering prometheus query: %v", err)
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
