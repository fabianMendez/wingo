package main

import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/fabianMendez/wingo/pkg/notifications"
)

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	uid := request.QueryStringParameters["uid"]

	if uid != "" {
		err := notifications.DeleteSetting(uid)
		if err != nil {
			log.Println(err)
		}
	}

	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "text/html",
		},
		Body: "<h1>Subscription has been cancelled</h1>",
	}, nil
}

func main() {
	lambda.Start(handler)
}
