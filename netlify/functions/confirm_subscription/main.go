package main

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/fabianMendez/wingo/pkg/notifications"
)

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var setting notifications.Setting
	uid := request.QueryStringParameters["uid"]

	setting, err := notifications.GetSetting(uid)
	if err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Headers: map[string]string{
				"Content-Type": "text/html",
			},
			Body: "<h1>Subscription not found</h1>",
		}, nil
	}

	setting.Confirmed = true
	err = notifications.UpdateSetting(uid, setting)
	if err != nil {
		return nil, err
	}

	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "text/html",
		},
		Body: "<h1>Subscription has been confirmed</h1>",
	}, nil
}

func main() {
	lambda.Start(handler)
}
