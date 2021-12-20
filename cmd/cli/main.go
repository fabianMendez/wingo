package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fabianMendez/bits/pkg/config"
	"github.com/fabianMendez/bits/pkg/email"
	"github.com/fabianMendez/wingo"
	"github.com/fabianMendez/wingo/pkg/date"
	"github.com/fabianMendez/wingo/pkg/notifications"
)

var (
	logger = log.Default()
)

const (
	outdir     = "flights"
	maxWorkers = 10
)

func formatMoney(n float64) string { return "$ " + humanize.FormatFloat("#,###.##", n) }

func showFlightInformation(client *wingo.Client, emailservice email.Service, notificationSettings []notifications.Setting, flightsInformation wingo.FlightsInformation, origin, destination string) error {
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

			err = sendNotifications(emailservice, notificationSettings, origin, destination, fecha, precio)
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

	// log.Println("calculando precio")
	// adminFares := wingo.GetAdminFares(serviceQuotes[0])
	// precio := wingo.GetBundlePrice(wingo.OriginalPlanName, vuelo, adminFares)

	// log.Printf("precio del vuelo %s-%s (%s - %s): %s\n", origin, destination, vuelo.FlightNumber, vuelo.DepartureDate, formatMoney(precio))

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

func sendNotifications(emailservice email.Service, notificationSettings []notifications.Setting, origin, destination, departureDate string, price float64) error {
	subject := fmt.Sprintf("Precio del viaje %s-%s %s", origin, destination, departureDate)
	body := fmt.Sprintf("El precio del viaje %s-%s para la fecha %s es: <b>%s</b>", origin, destination, departureDate, formatMoney(price))
	wingoLink := fmt.Sprintf("https://booking.wingo.com/es/search/%s/%s/%s/1/0/0/1/COP/0/0", origin, destination, departureDate)
	body += fmt.Sprintf("<br>Link: %s", wingoLink)

	return sendNotificationEmail(emailservice, notificationSettings, origin, destination, departureDate, subject, body)
}

func getNotificationEmails(notificationSettings []notifications.Setting, origin, destination, date string) []string {
	var emails []string
	for _, setting := range notificationSettings {
		if origin == setting.Origin && destination == setting.Destination && date == setting.Date {
			emails = append(emails, setting.Email)
		}
	}
	return emails
}

func sendNotificationEmail(emailservice email.Service, notificationSettings []notifications.Setting, origin, destination, date string, subject, body string) error {
	fmt.Println("===", subject, "===")
	fmt.Println(body)

	emails := getNotificationEmails(notificationSettings, origin, destination, date)
	for _, email := range emails {
		fmt.Println("["+email+"]:", subject)
		// err := emailservice.Send(subject, body, nil, email)
		// if err != nil {
		// 	return err
		// }
	}

	return nil
}

func sendNewFlightNotification(emailservice email.Service, notificationSettings []notifications.Setting, origin, destination, date string, price float64) error {
	subject := fmt.Sprintf("Precio del viaje %s-%s %s", origin, destination, date)
	body := fmt.Sprintf("El precio del viaje %s-%s para la fecha %s es: <b>%s</b>", origin, destination, date, formatMoney(price))
	wingoLink := fmt.Sprintf("https://booking.wingo.com/es/search/%s/%s/%s/1/0/0/1/COP/0/0", origin, destination, date)
	body += fmt.Sprintf("<br>Link: %s", wingoLink)

	return sendNotificationEmail(emailservice, notificationSettings, origin, destination, date, subject, body)
}

func sendPriceChangedNotification(emailservice email.Service, notificationSettings []notifications.Setting, origin, destination, date string, oldPrice, newPrice float64) error {
	accion := "SUBIÓ"
	if oldPrice > newPrice {
		accion = "BAJÓ"
	}

	subject := fmt.Sprintf("%s precio del vuelo %s-%s %s", accion, origin, destination, date)
	body := fmt.Sprintf("El precio del vuelo %s-%s para la fecha %s %s desde <b>%s</b> a <b>%s</b>",
		origin, destination, date, accion, formatMoney(oldPrice), formatMoney(newPrice))

	wingoLink := fmt.Sprintf("https://booking.wingo.com/es/search/%s/%s/%s/1/0/0/1/COP/0/0", origin, destination, date)
	body += fmt.Sprintf("<br>Link: %s", wingoLink)

	return sendNotificationEmail(emailservice, notificationSettings, origin, destination, date, subject, body)
}

func sendNotAvailableNotification(emailservice email.Service, notificationSettings []notifications.Setting, origin, destination, date string, lastPrice float64) error {
	subject := fmt.Sprintf("El vuelo %s-%s %s ya NO está disponible", origin, destination, date)
	body := fmt.Sprintf("El vuelo %s-%s %s ya NO está disponible", origin, destination, date)

	wingoLink := fmt.Sprintf("https://booking.wingo.com/es/search/%s/%s/%s/1/0/0/1/COP/0/0", origin, destination, date)
	body += fmt.Sprintf("<br>Link: %s", wingoLink)

	return sendNotificationEmail(emailservice, notificationSettings, origin, destination, date, subject, body)
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

func convertToTasks(flightsInformation wingo.FlightsInformation, origin, destination string) []getPriceTask {
	var tasks []getPriceTask

	fechaVuelos := filtrarVuelos(flightsInformation.VueloIda)
	for fecha, vuelos := range fechaVuelos {
		for _, vuelo := range vuelos {
			tasks = append(tasks, getPriceTask{
				fecha:       fecha,
				token:       flightsInformation.Token,
				origin:      origin,
				destination: destination,
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

func loadSavedFlights(savedRoutes []wingo.Route, startDate, stopDate time.Time) (flightsMap, error) {
	wg := new(sync.WaitGroup)
	flightsMutex := new(sync.Mutex)
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

			wg.Add(1)
			go func(origin, destination string, direntries []fs.DirEntry) {
				defer wg.Done()

				for _, dentry := range direntries {
					if !dentry.IsDir() {
						continue
					}

					flightsMutex.Lock()
					if flights[origin] == nil {
						flights[origin] = map[string]map[string][]vueloArchivado{}
					}

					if flights[origin][destination] == nil {
						flights[origin][destination] = map[string][]vueloArchivado{}
					}
					flightsMutex.Unlock()

					datepath := filepath.Join(dirname, dentry.Name())
					fentries, err := os.ReadDir(datepath)
					if err != nil {
						// return nil, err
						continue
					}

					for _, fentry := range fentries {
						var varchivado vueloArchivado

						flightpath := filepath.Join(datepath, fentry.Name())
						f, err := os.Open(flightpath)
						if err != nil {
							// return nil, err
							continue
						}
						defer f.Close()

						err = json.NewDecoder(f).Decode(&varchivado)
						if err != nil {
							// return nil, err
							continue
						}

						datestr := dentry.Name()
						date, err := date.Parse(datestr)
						if err != nil {
							continue
						}

						if date.Before(startDate) || date.After(stopDate) {
							continue
						}

						flightsMutex.Lock()
						flights[origin][destination][datestr] = append(
							flights[origin][destination][datestr],
							varchivado,
						)
						flightsMutex.Unlock()
					}
				}
			}(origin.Code, destination.Code, direntries)
		}
	}
	wg.Wait()

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

func processFlight(emailservice email.Service, notificationSettings []notifications.Setting, savedFlights flightsMap, date, origin, destination string, flight vueloArchivado) error {
	previous, previousFound := findFlight(savedFlights, origin, destination, date, flight.FlightNumber)

	price := calculatePrice(flight.Vuelo, flight.Services)
	// 1. Antes NO disponible y ahora disponible?
	if !previousFound {
		err := sendNewFlightNotification(emailservice, notificationSettings, origin, destination, date, price)
		if err != nil {
			return err
		}
	} else {
		savedPrice := calculatePrice(previous.Vuelo, previous.Services)
		// 2. Antes disponible y ahora diferente precio?
		if price != savedPrice {
			err := sendPriceChangedNotification(emailservice, notificationSettings, origin, destination, date, savedPrice, price)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func processUnavailableFlights(emailservice email.Service, notificationSettings []notifications.Setting, savedFlights flightsMap, actualFlights flightsMap) error {
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
						err := sendNotAvailableNotification(emailservice, notificationSettings, origin, destination, date, savedPrice)
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

func processSchedule(emailservice email.Service, notificationSettings []notifications.Setting,
	savedFlights flightsMap, date, origin, destination string, flight vueloArchivado) error {
	previous, previousFound := findFlight(savedFlights, origin, destination, date, flight.FlightNumber)

	price := calculatePrice(flight.Vuelo, flight.Services)
	// 1. Antes NO disponible y ahora disponible?
	if !previousFound {
		err := sendNewFlightNotification(emailservice, notificationSettings, origin, destination, date, price)
		if err != nil {
			return err
		}
	} else {
		savedPrice := calculatePrice(previous.Vuelo, previous.Services)
		// 2. Antes disponible y ahora diferente precio?
		if price != savedPrice {
			err := sendPriceChangedNotification(emailservice, notificationSettings, origin, destination, date, savedPrice, price)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func processUnavailableSchedules(emailservice email.Service, notificationSettings []notifications.Setting, savedFlights flightsMap, actualFlights flightsMap) error {
	// 3. Antes disponible y ahora NO disponible?
	for origin, originMap := range savedFlights {
		for destination, destinationMap := range originMap {
			for date, savedFlights := range destinationMap {
				for _, savedFlight := range savedFlights {
					_, actualFound := findFlight(actualFlights, origin, destination, date, savedFlight.FlightNumber)
					if !actualFound {
						// flightpath := filepath.Join(outdir, origin, destination, date, savedFlight.FlightNumber+".json")
						// _ = os.Remove(flightpath)

						savedPrice := calculatePrice(savedFlight.Vuelo, savedFlight.Services)
						err := sendNotAvailableNotification(emailservice, notificationSettings, origin, destination, date, savedPrice)
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

type archiveTask struct {
	fecha               string
	vuelo               wingo.Vuelo
	origin, destination string
	services            []wingo.Service
}

func retrieveServices(client *wingo.Client, getPriceTaskChan chan getPriceTask, archiveTasksChan chan<- archiveTask) {
	for task := range getPriceTaskChan {
		services, err := printInformation(client, task.fecha, task.vuelo, task.origin, task.destination, task.token)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		archiveTasksChan <- archiveTask{
			fecha:       task.fecha,
			vuelo:       task.vuelo,
			origin:      task.origin,
			destination: task.destination,
			services:    services,
		}
	}
}

func getInformationFlightsMonthly(client *wingo.Client, origin, destination string, startDate, endDate time.Time) []getPriceTask {
	daysAfter := int(endDate.Sub(startDate).Hours() / 24)
	fmt.Printf("Start date %s-%s: %s", origin, destination, date.Format(startDate))
	fmt.Printf("Days %s-%s: %d\n", origin, destination, daysAfter)
	flightsInformation, err := client.GetInformationFlightsMonthly(origin, destination, date.Format(startDate), daysAfter)
	if err != nil {
		log.Fatal(err)
	}

	tasks := convertToTasks(flightsInformation, origin, destination)
	fmt.Printf("Got flightsInformation %s-%s: %d\n", origin, destination, len(tasks))
	return tasks

	// for _, task := range tasks {
	// 	getPriceTaskChan <- task
	// }
	// fmt.Println("tasks sent")
}

func checkDate(startDate, endDate time.Time, dateStr string) error {
	f, err := date.Parse(dateStr)
	if err != nil {
		return err
	}
	if f.Before(startDate) {
		return fmt.Errorf("%v is before %v", f, startDate)
	} else if f.After(endDate) {
		return fmt.Errorf("%v is after %v", f, startDate)
	}
	return nil
}

func cleanFlightNumber(flightNumber string) string {
	i := strings.Index(flightNumber, "-")
	if i != -1 {
		return flightNumber[i+1:]
	}

	return flightNumber
}

func addFlightToMap(fmap flightsMap, origin, destination, date string, flight vueloArchivado) {
	if fmap[origin] == nil {
		fmap[origin] = map[string]map[string][]vueloArchivado{}
	}

	if fmap[origin][destination] == nil {
		fmap[origin][destination] = map[string][]vueloArchivado{}
	}

	fmap[origin][destination][date] = append(fmap[origin][destination][date], flight)
}

func processNotificationSettings(client *wingo.Client, emailservice email.Service, settings []notifications.Setting,
	savedFlights flightsMap, startDate, stopDate time.Time) error {

	type getFlightScheduleTask struct{ origin, destination string }
	getFlightsScheduleTasksChan := make(chan getFlightScheduleTask, maxWorkers)

	flightsInformation := []wingo.FlightInformation{}

	wg := workgroup(func() {
		for t := range getFlightsScheduleTasksChan {
			information, err := client.GetFlightScheduleInformation(t.origin, t.destination,
				date.Format(startDate), date.Format(stopDate))
			if err != nil {
				log.Println(err)
				continue
			}

			flightsInformation = append(flightsInformation, information.FlightInformation...)
		}
	}, maxWorkers)

	routes := map[string][]string{}
	for _, setting := range settings {
		if routes[setting.Origin] == nil {
			routes[setting.Origin] = []string{}
		}

		found := false
		for _, val := range routes[setting.Origin] {
			if val == setting.Destination {
				found = true
				break
			}
		}

		if !found {
			routes[setting.Origin] = append(routes[setting.Origin], setting.Destination)
			// routes[setting.Origin] = []string{setting.Destination}
		}
	}

	for origin, destinations := range routes {
		for _, destination := range destinations {
			getFlightsScheduleTasksChan <- getFlightScheduleTask{origin, destination}
		}
	}
	close(getFlightsScheduleTasksChan)
	wg.Wait()

	actualFlights := flightsMap{}

	for _, flightInf := range flightsInformation {
		flight := vueloArchivado{
			Vuelo: wingo.Vuelo{
				FlightNumber: cleanFlightNumber(flightInf.FlightNumber),
			},
			Services: []wingo.Service{},
		}
		fecha, err := date.Parse(flightInf.EffectiveDate)
		if err != nil {
			log.Println(err)
			continue
		}

		date := date.Format(*fecha)
		addFlightToMap(actualFlights, flightInf.Origin, flightInf.Destination, date, flight)
		processSchedule(emailservice, settings, savedFlights, date, flightInf.Origin, flightInf.Destination, flight)
	}

	// return processUnavailableSchedules(emailservice, settings, savedFlights, actualFlights)
	return nil
}

func main() {
	starttime := time.Now()
	defer func() {
		fmt.Fprintln(os.Stderr, "Duración:", time.Since(starttime))
	}()

	months, err := strconv.Atoi(os.Getenv("WINGO_MONTHS"))
	if err != nil {
		months = 1
	}

	startMonths, err := strconv.Atoi(os.Getenv("WINGO_START_MONTHS"))
	if err != nil {
		startMonths = 0
	}

	now := time.Now()
	startDate := time.Date(now.Year(), now.Month()+time.Month(startMonths), now.Day(), 0, 0, 0, 0, time.UTC)
	stopDate := startDate.AddDate(0, months, 0)

	fmt.Println("--------------------------------------")
	fmt.Println("  Start:", date.Format(startDate), "End:", date.Format(stopDate))
	fmt.Println("--------------------------------------")

	client := wingo.NewClient(logger)
	defer func() {
		fmt.Println("Request Count:", client.RequestCount)
	}()

	emailservice, err := email.NewService(&config.EmailConfig{
		Host:     "",
		Port:     587,
		Username: "",
		Password: "",
	})
	if err != nil {
		log.Fatal(err)
	}

	notificationSettings, err := notifications.LoadAllSettings()
	if err != nil {
		log.Fatal(err)
	}

	var savedRoutes []wingo.Route
	_ = loadFromFile("routes.json", &savedRoutes)

	logger.Println("Cargando rutas guardadas")
	savedFlights, err := loadSavedFlights(savedRoutes, startDate, stopDate)
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) >= 2 && os.Args[1] == "fast" {
		fmt.Fprintln(os.Stderr, "Settings count:", len(notificationSettings))
		processNotificationSettings(client, emailservice, notificationSettings, savedFlights, startDate, stopDate)
		return
	}

	logger.Println("Rutas guardadas:", len(savedRoutes))

	routes, err := client.GetRoutes()
	if err != nil {
		log.Fatal(err)
	}

	err = saveToFile("routes.json", routes)
	if err != nil {
		log.Fatal(err)
	}

	var getPriceTasks []getPriceTask

	// wg := sync.WaitGroup{}
	type getInformationFlightsTask struct {
		origin, destination string
		startDate, endDate  time.Time
	}
	getInformationFlightsChan := make(chan getInformationFlightsTask, maxWorkers)

	wg := workgroup(func() {
		for t := range getInformationFlightsChan {
			tasks := getInformationFlightsMonthly(client, t.origin, t.destination, t.startDate, t.endDate)
			for _, t2 := range tasks {
				if err := checkDate(t.startDate, t.endDate, t2.fecha); err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
			}
			getPriceTasks = append(getPriceTasks, tasks...)
		}
	}, maxWorkers)

	fmt.Println("Routes from api:", len(routes))
	count := 0
	for _, origin := range routes {
		for _, destination := range origin.Routes {
			fmt.Println(origin.Name, "=>", destination.Name)
			count++

			func(startDate time.Time) {
				for startDate.Before(stopDate) {
					endDate := startDate.AddDate(0, 1, 0)
					getInformationFlightsChan <- getInformationFlightsTask{
						origin:      origin.Code,
						destination: destination.Code,
						startDate:   startDate,
						endDate:     endDate,
					}
					startDate = endDate
				}
			}(startDate)
		}
	}
	close(getInformationFlightsChan)

	fmt.Println("--------------------------------------")
	fmt.Println("Waiting to load", count, "flights information")
	fmt.Println("--------------------------------------")
	wg.Wait()

	fmt.Println("------------------------------------")
	fmt.Println("Finished loading flights information")
	fmt.Println("------------------------------------")

	actualFlights := flightsMap{}
	actualFlightsMutex := new(sync.Mutex)
	archiveTaskChan := make(chan archiveTask, maxWorkers)
	wgArchive := workgroup(func() {
		for task := range archiveTaskChan {
			flight := vueloArchivado{
				Vuelo:    task.vuelo,
				Services: task.services,
			}

			actualFlightsMutex.Lock()
			addFlightToMap(actualFlights, task.origin, task.destination, task.fecha, flight)
			actualFlightsMutex.Unlock()

			// save to file
			fname := fmt.Sprintf("%s/%s/%s/%s/%s.json", outdir, task.origin, task.destination, task.fecha, flight.FlightNumber)

			_ = os.MkdirAll(filepath.Dir(fname), os.ModePerm)

			err := saveToFile(fname, flight)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			err = processFlight(emailservice, notificationSettings, savedFlights, task.fecha, task.origin, task.destination, flight)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}, maxWorkers)

	getPriceTaskChan := make(chan getPriceTask, maxWorkers)
	threadsG := workgroup(func() {
		retrieveServices(client, getPriceTaskChan, archiveTaskChan)
	}, maxWorkers)

	for _, task := range getPriceTasks {
		getPriceTaskChan <- task
	}
	close(getPriceTaskChan)
	threadsG.Wait()

	fmt.Println("------------------------------")
	fmt.Println(" Finished retrieving services ")
	fmt.Println("------------------------------")

	close(archiveTaskChan)
	wgArchive.Wait()

	fmt.Println("----------------------------------")
	fmt.Println(" Finished saving flights archives ")
	fmt.Println("----------------------------------")

	err = processUnavailableFlights(emailservice, notificationSettings, savedFlights, actualFlights)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
