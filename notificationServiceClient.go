package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

type NotificationServiceClient struct {
	url        *url.URL
	httpClient http.Client
	username   string
	password   string
}

type NotificationMessage struct {
	Id        string                    `json:"id"`
	Subject   string                    `json:"subject"`
	Message   string                    `json:"message"`
	ExpiresIn string                    `json:"validity,omitempty"`
	Target    NotificationMessageTarget `json:"target"`
}

type NotificationMessageTarget struct {
	Type        string `json:"type"`
	Environment string `json:"environment,omitempty"`
	Id          string `json:"id"`
}

func NewNotificationServiceClient(urlString, username, password string) *NotificationServiceClient {
	parsedUrl, err := url.Parse(urlString)
	if err != nil {
		log.Fatal("Unable to parse notification service URL")
	}

	return &NotificationServiceClient{
		url: parsedUrl,
		httpClient: http.Client{
			Timeout: 15 * time.Second,
		},
		username: username,
		password: password,
	}
}

func (c *NotificationServiceClient) Send(msg NotificationMessage) error {
	msgBody, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("Error unmarshalling message to json: %v", err)
	}

	req, _ := http.NewRequest("POST", c.url.String()+"/send", bytes.NewReader(msgBody))
	req.SetBasicAuth(c.username, c.password)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Error calling notification Service: %v", err.Error())
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("Unable to login to notification service")
	case http.StatusConflict:
		return fmt.Errorf("Message was sent before. It won't be sent again")
	case http.StatusOK:
		return nil
	default:
		return fmt.Errorf("Notification service return status code: %v", resp.StatusCode)
	}

}
