package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/fabianMendez/bits/syncbits"
	"github.com/fabianMendez/wingo"
	"github.com/fabianMendez/wingo/pkg/storage"
)

const (
	outdir     = "flights"
	maxWorkers = 10
)

var headers = map[string]string{
	"Access-Control-Allow-Origin":  "*",
	"Access-Control-Allow-Methods": "GET",
	"Access-Control-Max-Age":       "3600",
	"Access-Control-Allow-Headers": "Content-Type",
	"Content-Type":                 "application/json",
}

type History struct {
}

type vueloArchivado struct {
	wingo.Vuelo
	Services []wingo.Service `json:"services"`
}

func calculatePrice(vuelo wingo.Vuelo, services []wingo.Service) float64 {
	adminFares := wingo.GetAdminFares(wingo.ServiceQuote{
		Services: services,
	})

	return wingo.GetBundlePrice(wingo.OriginalPlanName, vuelo, adminFares)
}

func routeHistory(origin, destination, date, flightNumber string) (map[string]float64, error) {
	githubStorage, err := storage.NewGithubFromEnv()
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("%s/%s/%s/%s/%s.json", outdir, origin, destination, date, flightNumber)
	twoWeeksAgo := time.Now().AddDate(0, 0, -15)
	hashes, err := githubStorage.Commits(path, twoWeeksAgo)
	if err != nil {
		log.Println("could not get commits: ", err)
		return nil, err
	}
	vuelos := map[string]float64{}

	type task struct {
		sha  string
		date string
	}
	ch := make(chan task, maxWorkers)
	mutex := &sync.Mutex{}

	wg := syncbits.Workgroup(func() {
		for t := range ch {
			hash := t.sha
			date := t.date
			content, err := githubStorage.ReadRef(path, hash)
			if err != nil {
				fmt.Fprintln(os.Stderr, "could not read ref: ", err)
				return
			}
			var vuelo vueloArchivado
			err = json.Unmarshal(content, &vuelo)
			if err != nil {
				fmt.Fprintln(os.Stderr, "could not decode archived flight: ", err)
				return
			}
			price := calculatePrice(vuelo.Vuelo, vuelo.Services)
			fmt.Println(date, price)
			mutex.Lock()
			vuelos[date] = price
			mutex.Unlock()
		}
	}, maxWorkers)

	for date, hash := range hashes {
		ch <- task{hash, date}
	}
	close(ch)
	wg.Wait()

	return vuelos, nil
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	log.Println("fetching route history: ", request.Path)

	// BOG/HAV/2022-04-14
	params := strings.Split(request.Path, "/")
	nparams := 4
	params = params[len(params)-nparams:]
	if len(params) != nparams {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Headers:    headers,
			Body:       "Wrong arguments",
		}, nil
	}

	vuelos, err := routeHistory(params[0], params[1], params[2], params[3])
	if err != nil {
		log.Println(err)
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
			Body:       err.Error(),
		}, nil
	}

	log.Println("route information successfully retrieved")
	body, err := json.Marshal(vuelos)
	if err != nil {
		log.Println(err)
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
			Body:       err.Error(),
		}, nil
	}

	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    headers,
		Body:       string(body),
	}, nil
}

func main() {
	lambda.Start(handler)
}
