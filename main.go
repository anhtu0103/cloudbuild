package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"

	goteamsnotify "github.com/atc0005/go-teams-notify/v2"
	"github.com/atc0005/go-teams-notify/v2/messagecard"
)

type (
	Status string
	Build  struct {
		LogUrl      string `json:"logUrl"`
		RepoName    string ` json:"repoName"`
		Status      Status ` json:"status"`
		Tag         string ` json:"tag"`
		TriggerName string ` json:"triggerName"`
	}
)

const (
	QUEUED    Status = "QUEUED"
	WORKING   Status = "WORKING"
	SUCCESS   Status = "SUCCESS"
	FAILURE   Status = "FAILURE"
	CANCELLED Status = "CANCELLED"
)

func main() {
	log.Print("starting server...")
	http.HandleFunc("/", handler)

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	// Start HTTP server.
	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var b Build
	err := decoder.Decode(&b)
	if err != nil {
		fmt.Fprintf(w, "Failed %s!\n", err.Error())
		return
	}

	// Initialize a new Microsoft Teams client.
	mstClient := goteamsnotify.NewTeamsClient()

	// Set webhook url
	channel := "TEAMS_CHANNEL"
	webhookUrl := os.Getenv(channel)

	// // Destination for OpenUri action.
	targetURL := b.LogUrl
	targetURLDesc := "View Logs"

	// Setup message card.
	msgCard := messagecard.NewMessageCard()
	msgCard.Title = fmt.Sprintf("Cloud Build (%s)", b.TriggerName)
	msgCard.Text = "**Tag** " + b.Tag +
		"<br>**Status** " + string(b.Status)
	msgCard.ThemeColor = status2Theme(b.Status)

	// Setup Action for message card.
	pa, err := messagecard.NewPotentialAction(
		messagecard.PotentialActionOpenURIType,
		targetURLDesc,
	)
	if err != nil {
		log.Fatal("error encountered when creating new action:", err)
	}
	pa.PotentialActionOpenURI.Targets =
		[]messagecard.PotentialActionOpenURITarget{
			{
				OS:  "default",
				URI: targetURL,
			},
		}
	// Add the Action to the message card.
	if err := msgCard.AddPotentialAction(pa); err != nil {
		log.Fatal("error encountered when adding action to message card:", err)
	}

	// Send the message with default timeout/retry settings.
	if err := mstClient.Send(webhookUrl, msgCard); err != nil {
		log.Printf("failed to send message: %v", err)
	}

	fmt.Fprintf(w, "Success build!")
}

func status2Theme(status Status) string {
	switch status {
	case QUEUED:
		return "#808080"
	case WORKING:
		return "#FFA500"
	case SUCCESS:
		return "#008000"
	case FAILURE:
		return "#FF0000"
	case CANCELLED:
		return "#FFFFFF"
	}
	return ""
}

func iterateStructFields(input interface{}) {
	value := reflect.ValueOf(input)
	numFields := value.NumField()

	fmt.Printf("Number of fields: %d\n", numFields)

	structType := value.Type()

	for i := 0; i < numFields; i++ {
		field := structType.Field(i)
		fieldValue := value.Field(i)

		fmt.Printf("Field %d: %s (%s) = %v\n", i+1, field.Name, field.Type, fieldValue)
	}
}
