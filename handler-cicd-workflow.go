package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func handlerCICDWorkflow(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var b Build
	err := decoder.Decode(&b)
	if err != nil {
		fmt.Fprintf(w, "Failed %s!\n", err.Error())
		return
	}

	// Set webhook url
	channel := "TEAMS_CHANNEL_WORKFLOW"
	webhookUrl := os.Getenv(channel)

	// Setup message card.
	msgText := "  \n*Cloud Build*: " + b.TriggerName
	msgText += "  \n*Tag*: " + b.Tag
	msgText += "  \n*Status*: " + string(b.Status)

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
							Wrap: false,
							Text: msgText,
						},
						{
							Type: "ActionSet",
							Actions: []CardAction{
								{
									Type:  "Action.OpenUrl",
									Title: "View Logs",
									URL:   b.LogUrl,
									Style: "positive",
								},
							},
						},
					},
				},
			},
		},
	}

	payload, err := json.Marshal(card)
	if err != nil {
		log.Fatalln(err)
	}
	res, err := http.Post(webhookUrl, "application/json", bytes.NewBuffer(payload))
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
