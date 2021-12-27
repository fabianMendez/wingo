package main

import "github.com/fabianMendez/wingo"

type getPriceTask struct {
	fecha               string
	token               string
	origin, destination string
	vuelo               wingo.Vuelo
}

type vueloArchivado struct {
	wingo.Vuelo
	Services []wingo.Service `json:"services"`
}

// origin -> destination -> date -> flights
type flightsMap map[string]map[string]map[string][]vueloArchivado

type archiveTask struct {
	fecha               string
	vuelo               wingo.Vuelo
	origin, destination string
	services            []wingo.Service
}
