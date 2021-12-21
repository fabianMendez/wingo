package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/fabianMendez/bits/pkg/config"
	"github.com/fabianMendez/bits/pkg/email"
	"github.com/fabianMendez/wingo/pkg/notifications"
)

var baseURL = os.Getenv("URL")

const body = `<h1>Confirm your subscription</h1>
<br>
<p>Use the following link to confirm your subscription to receive notifications about price updates in the route:
<a href="{{.baseURL}}/.netlify/functions/confirm_notification_setting?uid={{.uid}}"> Confirm</a>
</p>
<br>
<p>If you did not request this subscription, please ignore this message.</p>
`

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var setting notifications.Setting
	err := json.Unmarshal([]byte(request.Body), &setting)
	if err != nil {
		return nil, err
	}

	setting.Confirmed = false

	uid, err := notifications.SaveSetting(setting)
	if err != nil {
		return nil, err
	}

	// send confirmation email
	err = emailservice.Send(`Please confirm your subscription`, body, map[string]string{
		"uid":     uid,
		"baseURL": baseURL,
	}, setting.Email)
	if err != nil {
		return nil, err
	}

	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "POST",
			"Access-Control-Max-Age":       "3600",
			"Access-Control-Allow-Headers": "Content-Type",
		},
	}, nil
}

var emailservice email.Service

func main() {
	var err error
	emailservice, err = email.NewService(&config.EmailConfig{
		Host:     "",
		Port:     587,
		Username: "",
		Password: "",
	})
	if err != nil {
		log.Fatal(err)
	}

	lambda.Start(handler)
}
