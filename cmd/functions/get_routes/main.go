package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type fileContentsResponse struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	SHA         string `json:"sha"`
	Size        int64  `json:"size"`
	URL         string `json:"url"`
	HTMLURL     string `json:"html_url"`
	GitURL      string `json:"git_url"`
	DownloadURL string `json:"download_url"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	Encoding    string `json:"encoding"`
	Links       struct {
		Self string `json:"self"`
		Git  string `json:"git"`
		HTML string `json:"html"`
	} `json:"_links"`
}

func getRoutes() ([]byte, error) {
	u := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", "fabianMendez", "wingo", "routes.json")
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Authorization", "token ghtoken")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not send request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed: %s", resp.Status)
	}
	defer resp.Body.Close()

	var response fileContentsResponse

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, fmt.Errorf("could not decode response: %w", err)
	}

	b, err := base64.StdEncoding.DecodeString(response.Content)
	if err != nil {
		return nil, fmt.Errorf("could not decode base64: %w", err)
	}

	return b, nil
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	b, err := getRoutes()
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
			"Content-Type":                 "application/json",
		},
		Body: string(b),
	}, nil
}

func main() {
	lambda.Start(handler)
}
