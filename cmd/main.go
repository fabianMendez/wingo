package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fabianMendez/bits/pkg/config"
	"github.com/fabianMendez/bits/pkg/email"
	"github.com/fabianMendez/wingo"
)

var (
	logger               = log.Default()
	notificationSettings = []notificationSetting{
		{
			origin:      "HAV",
			destination: "BOG",
			date:        "2021-07-20",
			email:       "",
		},
	}
)

const (
	months     = 10
	outdir     = "flights"
	maxWorkers = 10
)

type notificationSetting struct {
	origin      string
	destination string
	date        string
	email       string
}

func formatMoney(n float64) string { return "$ " + humanize.FormatFloat("#,###.##", n) }

func showFlightInformation(client *wingo.Client, emailservice email.Service, flightsInformation wingo.FlightsInformation, origin, destination string) error {
	// fmt.Printf("%#v\n", flightsInformation)
	// log.Println("Vuelos Ida:", len(flightsInformation.VueloIda))

	fechaVuelos := filtrarVuelos(flightsInformation.VueloIda)
	log.Println("Fechas:", len(fechaVuelos))

	for fecha, vuelos := range fechaVuelos {
		for _, vuelo := range vuelos {
			log.Printf("buscando tarifas servicios del vuelo %s - %s\n", vuelo.FlightNumber, vuelo.DepartureDate)
			serviceQuotes, err := client.RetrieveServiceQuotes([]wingo.FlightService{
				{Departure: fecha, From: origin, To: destination, FlightID: vuelo.LogicalFlightID},
			}, flightsInformation.Token)
			if err != nil {
				return err
			}
			log.Println("tarifas encontradas")

			log.Println("calculando precio")
			adminFares := wingo.GetAdminFares(serviceQuotes[0])
			precio := wingo.GetBundlePrice(wingo.OriginalPlanName, vuelo, adminFares)

			log.Println("Fecha:", vuelo.DepartureDate, "Precio:", formatMoney(precio))

			err = sendNotifications(emailservice, origin, destination, fecha, precio)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func filtrarVuelos(vuelos []wingo.VueloIda) map[string][]wingo.Vuelo {
	vuelosFiltrados := map[string][]wingo.Vuelo{}

	for _, flight := range vuelos {
		if len(flight.InfoVuelo.Vuelos) > 0 {
			for _, vuelo := range flight.InfoVuelo.Vuelos {
				price := wingo.GetBundlePrice(wingo.OriginalPlanName, vuelo, 0)
				if price != 0 {
					vuelosFiltrados[flight.Fecha] = append(vuelosFiltrados[flight.Fecha], vuelo)
					// log.Printf("buscando tarifas servicios del vuelo %s - %s\n", vuelo.FlightNumber, vuelo.DepartureDate)
					// []wingo.FlightService{ {Departure: flight.Fecha, From: origin, To: destination, FlightID: vuelo.LogicalFlightID}, }, flightsInformation.Token
				}

			}
		}
	}

	return vuelosFiltrados
}

type getPriceTask struct {
	fecha               string
	token               string
	origin, destination string
	vuelo               wingo.Vuelo
}

func printInformation(client *wingo.Client, fecha string, vuelo wingo.Vuelo, origin, destination, token string) ([]wingo.Service, error) {
	log.Printf("buscando tarifas servicios del vuelo %s-%s: %s - %s\n", origin, destination, vuelo.FlightNumber, vuelo.DepartureDate)
	serviceQuotes, err := client.RetrieveServiceQuotes([]wingo.FlightService{
		{Departure: fecha, From: origin, To: destination, FlightID: vuelo.LogicalFlightID},
	}, token)
	if err != nil {
		return nil, err
	}
	log.Printf("tarifas encontradas del vuelo %s-%s: %s - %s\n", origin, destination, vuelo.FlightNumber, vuelo.DepartureDate)

	log.Println("calculando precio")
	adminFares := wingo.GetAdminFares(serviceQuotes[0])
	precio := wingo.GetBundlePrice(wingo.OriginalPlanName, vuelo, adminFares)

	log.Printf("precio del vuelo %s-%s (%s - %s): %s\n", origin, destination, vuelo.FlightNumber, vuelo.DepartureDate, formatMoney(precio))

	return serviceQuotes[0].Services, nil
}

func printFlightInformation(client *wingo.Client, flightsInformation wingo.FlightsInformation, origin, destination string) error {
	// fmt.Printf("%#v\n", flightsInformation)
	// log.Println("Vuelos Ida:", len(flightsInformation.VueloIda))

	fechaVuelos := filtrarVuelos(flightsInformation.VueloIda)
	log.Println("Fechas:", len(fechaVuelos))

	for fecha, vuelos := range fechaVuelos {
		for _, vuelo := range vuelos {
			_, err := printInformation(client, fecha, vuelo, origin, destination, flightsInformation.Token)
			if err != nil {
				return err
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

	return sendNotificationEmail(emailservice, origin, destination, departureDate, subject, body)
}

func getNotificationEmails(origin, destination, date string) []string {
	var emails []string
	for _, setting := range notificationSettings {
		if origin == setting.origin && destination == setting.destination && date == setting.date {
			emails = append(emails, setting.email)
		}
	}
	return emails
}

func sendNotificationEmail(emailservice email.Service, origin, destination, date string, subject, body string) error {
	fmt.Println("===", subject, "===")

	emails := getNotificationEmails(origin, destination, date)
	for _, email := range emails {
		err := emailservice.Send(subject, body, nil, email)
		if err != nil {
			return err
		}
	}

	return nil
}

func sendNewFlightNotification(emailservice email.Service, origin, destination, date string, price float64) error {
	subject := fmt.Sprintf("Precio del viaje %s-%s %s", origin, destination, date)
	body := fmt.Sprintf("El precio del viaje %s-%s para la fecha %s es: <b>%s</b>", origin, destination, date, formatMoney(price))
	wingoLink := fmt.Sprintf("https://booking.wingo.com/es/search/%s/%s/%s/1/0/0/1/COP/0/0", origin, destination, date)
	body += fmt.Sprintf("<br>Link: %s", wingoLink)

	return sendNotificationEmail(emailservice, origin, destination, date, subject, body)
}

func sendPriceChangedNotification(emailservice email.Service, origin, destination, date string, oldPrice, newPrice float64) error {
	accion := "SUBIÓ"
	if oldPrice > newPrice {
		accion = "BAJÓ"
	}

	subject := fmt.Sprintf("%s precio del vuelo %s-%s %s", accion, origin, destination, date)
	body := fmt.Sprintf("El precio del vuelo %s-%s para la fecha %s %s desde <b>%s</b> a <b>%s</b>",
		origin, destination, date, accion, formatMoney(oldPrice), formatMoney(newPrice))

	wingoLink := fmt.Sprintf("https://booking.wingo.com/es/search/%s/%s/%s/1/0/0/1/COP/0/0", origin, destination, date)
	body += fmt.Sprintf("<br>Link: %s", wingoLink)

	return sendNotificationEmail(emailservice, origin, destination, date, subject, body)
}

func sendNotAvailableNotification(emailservice email.Service, origin, destination, date string, lastPrice float64) error {
	subject := fmt.Sprintf("El vuelo %s-%s %s ya NO está disponible", origin, destination, date)
	body := fmt.Sprintf("El vuelo %s-%s %s ya NO está disponible", origin, destination, date)

	wingoLink := fmt.Sprintf("https://booking.wingo.com/es/search/%s/%s/%s/1/0/0/1/COP/0/0", origin, destination, date)
	body += fmt.Sprintf("<br>Link: %s", wingoLink)

	return sendNotificationEmail(emailservice, origin, destination, date, subject, body)
}

func saveToFile(filename string, v interface{}) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, b, os.ModePerm)
}

func loadFromFile(filename string, v interface{}) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}

	return json.NewDecoder(f).Decode(v)
}

type flightsInformationResult struct {
	flightsInformation  wingo.FlightsInformation
	origin, destination wingo.Route
}

func convertToTasks(result flightsInformationResult) []getPriceTask {
	var tasks []getPriceTask

	fechaVuelos := filtrarVuelos(result.flightsInformation.VueloIda)
	for fecha, vuelos := range fechaVuelos {
		for _, vuelo := range vuelos {
			tasks = append(tasks, getPriceTask{
				fecha:       fecha,
				token:       result.flightsInformation.Token,
				origin:      result.origin.Code,
				destination: result.destination.Code,
				vuelo:       vuelo,
			})
		}
	}

	return tasks
}

func workgroup(fn func(), count int) *sync.WaitGroup {
	wg := sync.WaitGroup{}

	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			fn()
		}()
		wg.Add(1)
	}

	return &wg
}

type vueloArchivado struct {
	wingo.Vuelo
	Services []wingo.Service `json:"services"`
}

// origin -> destination -> date -> flights
type flightsMap map[string]map[string]map[string][]vueloArchivado

func loadSavedFlights(savedRoutes []wingo.Route) (flightsMap, error) {
	flights := flightsMap{}

	for _, origin := range savedRoutes {
		for _, destination := range origin.Routes {
			dirname := filepath.Join(outdir, origin.Code, destination.Code)
			direntries, err := os.ReadDir(dirname)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return nil, err
			}

			for _, dentry := range direntries {
				if !dentry.IsDir() {
					continue
				}

				if flights[origin.Code] == nil {
					flights[origin.Code] = map[string]map[string][]vueloArchivado{}
				}

				if flights[origin.Code][destination.Code] == nil {
					flights[origin.Code][destination.Code] = map[string][]vueloArchivado{}
				}

				datepath := filepath.Join(dirname, dentry.Name())
				fentries, err := os.ReadDir(datepath)
				if err != nil {
					return nil, err
				}

				for _, fentry := range fentries {
					var varchivado vueloArchivado

					flightpath := filepath.Join(datepath, fentry.Name())
					f, err := os.Open(flightpath)
					if err != nil {
						return nil, err
					}
					defer f.Close()

					err = json.NewDecoder(f).Decode(&varchivado)
					if err != nil {
						return nil, err
					}

					flights[origin.Code][destination.Code][dentry.Name()] = append(
						flights[origin.Code][destination.Code][dentry.Name()],
						varchivado,
					)
				}
			}
		}
	}

	return flights, nil
}

func calculatePrice(vuelo wingo.Vuelo, services []wingo.Service) float64 {
	adminFares := wingo.GetAdminFares(wingo.ServiceQuote{
		Services: services,
	})

	return wingo.GetBundlePrice(wingo.OriginalPlanName, vuelo, adminFares)
}

func findFlight(savedFlights flightsMap, origin, destination, date, flightNumber string) (vueloArchivado, bool) {
	for savedOrigin, originMap := range savedFlights {
		if origin != savedOrigin {
			continue
		}

		for savedDestination, destinationMap := range originMap {
			if destination != savedDestination {
				continue
			}

			for savedDate, savedFlights := range destinationMap {
				if savedDate != date {
					continue
				}

				for _, savedFlight := range savedFlights {
					if savedFlight.FlightNumber == flightNumber {
						return savedFlight, true
					}
				}
			}
		}
	}

	return vueloArchivado{}, false
}

func processFlight(emailservice email.Service, savedFlights flightsMap, date, origin, destination string, flight vueloArchivado) error {
	previous, previousFound := findFlight(savedFlights, origin, destination, date, flight.FlightNumber)

	price := calculatePrice(flight.Vuelo, flight.Services)
	// 1. Antes NO disponible y ahora disponible?
	if !previousFound {
		err := sendNewFlightNotification(emailservice, origin, destination, date, price)
		if err != nil {
			return err
		}
	} else {
		savedPrice := calculatePrice(previous.Vuelo, previous.Services)
		// 2. Antes disponible y ahora diferente precio?
		if price != savedPrice {
			err := sendPriceChangedNotification(emailservice, origin, destination, date, savedPrice, price)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func processUnavailableFlights(emailservice email.Service, savedFlights flightsMap, actualFlights flightsMap) error {
	// 3. Antes disponible y ahora NO disponible?
	for origin, originMap := range savedFlights {
		for destination, destinationMap := range originMap {
			for date, savedFlights := range destinationMap {
				for _, savedFlight := range savedFlights {
					_, actualFound := findFlight(actualFlights, origin, destination, date, savedFlight.FlightNumber)
					if !actualFound {
						flightpath := filepath.Join(outdir, origin, destination, date, savedFlight.FlightNumber+".json")
						_ = os.Remove(flightpath)

						savedPrice := calculatePrice(savedFlight.Vuelo, savedFlight.Services)
						err := sendNotAvailableNotification(emailservice, origin, destination, date, savedPrice)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}

func main() {
	starttime := time.Now()

	defer func() {
		fmt.Fprintln(os.Stderr, "Duración:", time.Since(starttime))
	}()

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

	var savedRoutes []wingo.Route
	_ = loadFromFile("routes.json", &savedRoutes)

	savedFlights, err := loadSavedFlights(savedRoutes)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Rutas guardadas:", len(savedRoutes))

	routes, err := client.GetRoutes()
	if err != nil {
		log.Fatal(err)
	} else {
		err = saveToFile("routes.json", routes)
		if err != nil {
			log.Fatal(err)
		}

		type archiveTask struct {
			fecha               string
			vuelo               wingo.Vuelo
			origin, destination string
			services            []wingo.Service
		}

		archiveTasks := []archiveTask{}
		archiveTasksMutex := sync.Mutex{}

		getPriceTaskChan := make(chan getPriceTask, maxWorkers)
		threadsG := workgroup(func() {
			for task := range getPriceTaskChan {
				services, err := printInformation(client, task.fecha, task.vuelo, task.origin, task.destination, task.token)
				if err != nil {
					log.Fatal(err)
				}

				archiveTasksMutex.Lock()
				archiveTasks = append(archiveTasks, archiveTask{
					fecha:       task.fecha,
					vuelo:       task.vuelo,
					origin:      task.origin,
					destination: task.destination,
					services:    services,
				})
				archiveTasksMutex.Unlock()

			}
		}, maxWorkers)

		wg := sync.WaitGroup{}
		count := 0
		for _, origin := range routes {
			for _, destination := range origin.Routes {
				fmt.Println(origin.Name, "=>", destination.Name)
				count++

				go func(origin, destination wingo.Route) {
					defer wg.Done()

					startDate := time.Now()
					endDate := startDate.AddDate(0, months, 0)
					days := int(endDate.Sub(startDate).Hours() / 24)
					flightsInformation, err := client.GetInformationFlightsMonthly(origin.Code, destination.Code, wingo.FormatDate(startDate), days)
					if err != nil {
						log.Fatal(err)
					}

					result := flightsInformationResult{
						flightsInformation: flightsInformation,
						origin:             origin,
						destination:        destination,
					}

					for _, task := range convertToTasks(result) {
						getPriceTaskChan <- task
					}

				}(origin, destination)
				wg.Add(1)
			}
		}

		fmt.Println("--------------------------------------")
		fmt.Println("Waiting to load", count, "flights information")
		fmt.Println("--------------------------------------")
		wg.Wait()
		close(getPriceTaskChan)
		fmt.Println("------------------------------------")
		fmt.Println("Finished loading flights information")
		fmt.Println("------------------------------------")

		threadsG.Wait()

		actualFlights := flightsMap{}

		for _, task := range archiveTasks {
			if actualFlights[task.origin] == nil {
				actualFlights[task.origin] = map[string]map[string][]vueloArchivado{}
			}

			if actualFlights[task.origin][task.destination] == nil {
				actualFlights[task.origin][task.destination] = map[string][]vueloArchivado{}
			}

			actualFlights[task.origin][task.destination][task.fecha] = append(actualFlights[task.origin][task.destination][task.fecha], vueloArchivado{
				Vuelo:    task.vuelo,
				Services: task.services,
			})
		}

		for origin, destinationMap := range actualFlights {
			for destination, dateMap := range destinationMap {
				for date, flights := range dateMap {
					for _, flight := range flights {
						fname := fmt.Sprintf("%s/%s/%s/%s/%s.json", outdir, origin, destination, date, flight.FlightNumber)

						_ = os.MkdirAll(filepath.Dir(fname), os.ModePerm)

						b, err := json.MarshalIndent(flight, "", "  ")
						if err != nil {
							log.Fatal(err)
						}

						err = os.WriteFile(fname, b, os.ModePerm)
						if err != nil {
							fmt.Fprintln(os.Stderr, err)
						}

						err = processFlight(emailservice, savedFlights, date, origin, destination, flight)
						if err != nil {
							fmt.Fprintln(os.Stderr, err)
						}
					}
				}
			}
		}

		err = processUnavailableFlights(emailservice, savedFlights, actualFlights)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		return
	}
}
