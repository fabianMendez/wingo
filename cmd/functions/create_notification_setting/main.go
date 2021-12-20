package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/fabianMendez/wingo/pkg/notifications"
)

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var setting notifications.Setting
	err := json.Unmarshal([]byte(request.Body), &setting)
	if err != nil {
		return nil, err
	}

	setting.Confirmed = false

	err = notifications.SaveSetting(setting)
	if err != nil {
		return nil, err
	}

	// TODO: send confirmation email

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

func main() {
	lambda.Start(handler)
}
