package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/fabianMendez/wingo/pkg/email"
	"github.com/fabianMendez/wingo/pkg/notifications"
)

var baseURL = os.Getenv("URL")

var headers = map[string]string{
	"Access-Control-Allow-Origin":  "*",
	"Access-Control-Allow-Methods": "POST",
	"Access-Control-Max-Age":       "3600",
	"Access-Control-Allow-Headers": "Content-Type",
	"Content-Type":                 "application/json",
}

func createSubscription(ctx context.Context, body []byte) error {
	var setting notifications.Setting
	err := json.Unmarshal(body, &setting)
	if err != nil {
		return err
	}

	setting.Confirmed = false
	uid, err := notifications.SaveSetting(setting)
	if err != nil {
		return err
	}

	link := baseURL + "/.netlify/functions/confirm_subscription?uid=" + uid
	// `Please confirm your subscription`
	err = email.SendMessage(ctx, `Por favor confirma tu suscripción`, email.TplConfirmSubscription, map[string]interface{}{
		"subscription": setting,
		"link":         link,
	}, setting.Email)
	if err != nil {
		return fmt.Errorf("could not send message: %w", err)
	}

	log.Println("Message sent")

	return nil
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	log.Println("creating subscription")

	err := createSubscription(ctx, []byte(request.Body))
	if err != nil {
		log.Println(err)
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
			Body:       err.Error(),
		}, nil
	}

	log.Println("subscription successfully created")

	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    headers,
	}, nil
}

func main() {
	lambda.Start(handler)
}
