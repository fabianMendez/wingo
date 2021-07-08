package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fabianMendez/bits/pkg/config"
	"github.com/fabianMendez/bits/pkg/email"
	"github.com/fabianMendez/wingo"
)

var logger = log.Default()

func formatMoney(n float64) string { return "$ " + humanize.FormatFloat("#,###.##", n) }

func showFlightInformation(client *wingo.Client, emailservice email.Service, flightsInformation wingo.FlightsInformation, origin, destination string) error {
	// fmt.Printf("%#v\n", flightsInformation)
	log.Println("Vuelos Ida:", len(flightsInformation.VueloIda))

	for _, flight := range flightsInformation.VueloIda {
		if len(flight.InfoVuelo.Vuelos) > 0 {
			log.Println("Vuelos:", len(flight.InfoVuelo.Vuelos))
			for _, vuelo := range flight.InfoVuelo.Vuelos {
				originalPrice := wingo.GetBundlePrice(wingo.OriginalPlanName, vuelo, 0)
				if originalPrice == 0 {
					log.Println("precio 0 - saltando")
					log.Printf("%#v\n", vuelo)
				}

				log.Printf("buscando tarifas servicios del vuelo %s - %s\n", vuelo.FlightNumber, vuelo.DepartureDate)
				serviceQuotes, err := client.RetrieveServiceQuotes([]wingo.FlightService{
					{Departure: flight.Fecha, From: origin, To: destination, FlightID: vuelo.LogicalFlightID},
				}, flightsInformation.Token)
				if err != nil {
					return err
				}
				log.Println("tarifas encontradas")

				log.Println("calculando precio")
				adminFares := wingo.GetAdminFares(serviceQuotes[0])
				precio := wingo.GetBundlePrice(wingo.OriginalPlanName, vuelo, adminFares)

				log.Println("Fecha:", vuelo.DepartureDate, "Precio:", formatMoney(precio))

				err = sendNotifications(emailservice, origin, destination, flight.Fecha, precio)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func sendNotifications(emailservice email.Service, origin, destination, departureDate string, price float64) error {
	subject := fmt.Sprintf("Precio del viaje %s-%s %s", origin, destination, departureDate)
	body := fmt.Sprintf("El precio del viaje %s-%s para la fecha %s es: <b>%s</b>", origin, destination, departureDate, formatMoney(price))
	wingoLink := fmt.Sprintf("https://booking.wingo.com/es/search/%s/%s/%s/1/0/0/1/COP/0/0", origin, destination, departureDate)
	body += fmt.Sprintf("<br>Link: %s", wingoLink)

	err := emailservice.Send(subject, body, nil, "")
	if err != nil {
		return err
	}

	return nil
}

func main() {
	client := wingo.NewClient(logger)
	emailservice, err := email.NewService(&config.EmailConfig{
		Host:     "",
		Port:     587,
		Username: "",
		Password: "",
	})
	if err != nil {
		log.Fatal(err)
	}

	if true {
		// now := time.Now()
		// flightsInformation, err := getInformationFlightsMonthly("BOG", "HAV", formatDate(now), 100)
		// origin, destination, departureDate := "BOG", "HAV", "2021-08-03"
		origin, destination, departureDate := "HAV", "BOG", "2021-07-20"

		logger.Printf("buscando vuelos %s-%s\n", origin, destination)
		flightsInformation, err := client.GetInformationFlightsMonthly(origin, destination, departureDate, 0)
		if err != nil {
			log.Fatal(err)
		}
		logger.Printf("vuelos encontrados")

		err = showFlightInformation(client, emailservice, flightsInformation, origin, destination)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		flightDate := time.Now()
		stop := flightDate.AddDate(1, 0, 0)

		for flightDate.Before(stop) {
			days := 4
			// flightsInformation, err := getInformationFlightsMonthly("HAV", "BOG")
			flightsInformation, err := client.GetInformationFlightsMonthly("BOG", "HAV", wingo.FormatDate(flightDate), days)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(wingo.FormatDate(flightDate))
			showFlightInformation(client, emailservice, flightsInformation, "BOG", "HAV")

			flightDate = flightDate.AddDate(0, 0, days)
		}
	}

}
