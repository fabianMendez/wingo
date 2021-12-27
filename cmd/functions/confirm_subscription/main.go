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
				"Content-Type": "text/html; charset=utf-8",
			},
			Body: `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Document</title>
</head>
<body>
	<h1>Suscripción no encontrada</h1>
</body>
</html>
		`,
		}, nil
	}

	if !setting.Confirmed {
		setting.Confirmed = true
		err = notifications.UpdateSetting(uid, setting)
		if err != nil {
			return nil, err
		}
	}

	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "text/html; charset=utf-8",
		},
		Body: `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Document</title>
</head>
<body>
	<h1>La suscripción ha sido confirmada</h1>
</body>
</html>
		`,
	}, nil
}

func main() {
	lambda.Start(handler)
}
