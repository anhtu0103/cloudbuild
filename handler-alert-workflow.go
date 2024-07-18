package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

type (
	AdaptiveCard struct {
		Attachments []CardAttachment `json:"attachments"`
	}
	CardAttachment struct {
		ContentType string      `json:"contentType" default:"application/vnd.microsoft.card.adaptive"`
		Content     CardContent `json:"content"`
	}
	CardContent struct {
		Schema  string     `json:"$schema" default:"http://adaptivecards.io/schemas/adaptive-card.json"`
		Type    string     `json:"type" default:"AdaptiveCard"`
		Version string     `json:"version" default:"1.6"`
		Body    []CardBody `json:"body"`
	}
	CardBody struct {
		Type    string       `json:"type"`
		Wrap    bool         `json:"wrap,omitempty"`
		Text    string       `json:"text,omitempty"`
		Actions []CardAction `json:"actions,omitempty"`
	}
	CardAction struct {
		Type  string `json:"type"`
		Title string `json:"title"`
		URL   string `json:"url"`
		Style string `json:"style"`
	}
)

func handlerAlertWorkflow(w http.ResponseWriter, r *http.Request) {
	teamsWebhookURL := os.Getenv("TEAMS_CHANNEL_WORKFLOW")
	if teamsWebhookURL == "" {
		log.Fatalln("`TEAMS_CHANNEL_WORKFLOW` is not set in the environment")
	}

	if contentType := r.Header.Get("Content-Type"); r.Method != "POST" ||
		contentType != "application/json" {
		log.Printf("\ninvalid method / content-type: %s / %s \n", r.Method, contentType)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid request"))
		return
	}

	var notification Notification
	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		log.Fatalln(err)
	}

	card := toTeamsWorkFlow(notification)
	payload, err := json.Marshal(card)
	if err != nil {
		log.Fatalln(err)
	}

	res, err := http.Post(teamsWebhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		log.Println("payload", string(payload))
		log.Fatalln("unexpected status code", res.StatusCode)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(card)
	if err != nil {
		log.Fatalln(err)
	}
}

func toTeamsWorkFlow(notification Notification) AdaptiveCard {
	conditionName := notification.Incident.ConditionName
	policyName := notification.Incident.PolicyName
	if policyName == "" {
		policyName = "-"
	}
	if conditionName == "" {
		conditionName = "-"
	}
	var started time.Time
	var ended time.Time

	if st := notification.Incident.StartedAt; st > 0 {
		started = time.Unix(st, 0)
	}

	if et := notification.Incident.EndedAt; et > 0 {
		ended = time.Unix(et, 0)
	}

	msgText := "  \n*Incident ID*: " + notification.Incident.IncidentID +
		"  \n*Condition*: " + conditionName
	if notification.Incident.Summary != "" {
		msgText += "  \n*Summary*: " + notification.Incident.Summary
	}
	if notification.Incident.State == "open" {
		msgText += "  \n*Status*: Opened"
	} else {
		msgText += "  \n*Status*: Closed"
	}
	if !started.IsZero() {
		msgText += "  \n*Started at*: " + started.String()
		if !ended.IsZero() {
			duration := strings.TrimSpace(humanize.RelTime(started, ended, "", ""))
			msgText += "  \n*Ended at*: " + fmt.Sprintf("%s (%s)", ended.String(), duration)
		}
	}
	card := AdaptiveCard{
		Attachments: []CardAttachment{
			{
				ContentType: "application/vnd.microsoft.card.adaptive",
				Content: CardContent{
					Schema:  "http://adaptivecards.io/schemas/adaptive-card.json",
					Type:    "AdaptiveCard",
					Version: "1.0",
					Body: []CardBody{
						{
							Type: "TextBlock",
							Wrap: true,
							Text: msgText,
						},
						{
							Type: "ActionSet",
							Actions: []CardAction{
								{
									Type:  "Action.OpenUrl",
									Title: "View Incident",
									URL:   notification.Incident.URL,
									Style: "positive",
								},
							},
						},
					},
				},
			},
		},
	}

	return card
}
