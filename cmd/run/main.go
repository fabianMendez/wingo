package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fabianMendez/bits/syncbits"
	"github.com/fabianMendez/wingo"
	"github.com/fabianMendez/wingo/pkg/date"
	"github.com/fabianMendez/wingo/pkg/email"
	"github.com/fabianMendez/wingo/pkg/notifications"
	"github.com/fabianMendez/wingo/pkg/whatsapp"
)

var (
	logger = log.Default()
)

const (
	outdir     = "flights"
	maxWorkers = 10
)

func formatMoney(n float64) string { return "$" + humanize.FormatFloat("#,###.##", n) }

func filtrarVuelos(vuelos []wingo.VueloIda) map[string][]wingo.Vuelo {
	filtrados := map[string][]wingo.Vuelo{}

	for _, flight := range vuelos {
		if len(flight.InfoVuelo.Vuelos) > 0 {
			for _, vuelo := range flight.InfoVuelo.Vuelos {
				price := wingo.GetBundlePrice(wingo.OriginalPlanName, vuelo, 0)
				if price != 0 {
					filtrados[flight.Fecha] = append(filtrados[flight.Fecha], vuelo)
					// log.Printf("buscando tarifas servicios del vuelo %s - %s\n", vuelo.FlightNumber, vuelo.DepartureDate)
					// []wingo.FlightService{ {Departure: flight.Fecha, From: origin, To: destination, FlightID: vuelo.LogicalFlightID}, }, flightsInformation.Token
				}

			}
		}
	}

	return filtrados
}

var serviceCache map[string][]wingo.Service
var mx sync.Mutex

func printInformation(client *wingo.Client, fecha string, vuelo wingo.Vuelo, origin, destination, token string) ([]wingo.Service, error) {
	log.Printf("buscando tarifas servicios del vuelo %s-%s (%s): %s - %s\n", origin, destination, vuelo.DepartureDate, vuelo.FlightNumber, vuelo.DepartureDate)
	now := time.Now()

	mx.Lock()
	if serviceCache == nil {
		serviceCache = make(map[string][]wingo.Service)
	}
	cachekey := origin + "-" + destination
	if services, found := serviceCache[cachekey]; found {
		return services, nil
	}
	mx.Unlock()

	serviceQuotes, err := client.RetrieveServiceQuotes([]wingo.FlightService{
		{
			Departure:              fecha,
			AnticipationDateFlight: now.Format(time.RFC3339),
			From:                   origin,
			To:                     destination,
			FlightID:               vuelo.LogicalFlightID,
		},
	}, token)
	if err != nil {
		return nil, err
	}
	log.Printf("tarifas encontradas del vuelo %s-%s (%s): %s - %s\n", origin, destination, vuelo.DepartureDate, vuelo.FlightNumber, vuelo.DepartureDate)

	// log.Println("calculando precio")
	// adminFares := wingo.GetAdminFares(serviceQuotes[0])
	// precio := wingo.GetBundlePrice(wingo.OriginalPlanName, vuelo, adminFares)

	// log.Printf("precio del vuelo %s-%s (%s - %s): %s\n", origin, destination, vuelo.FlightNumber, vuelo.DepartureDate, formatMoney(precio))

	services := serviceQuotes[0].Services
	mx.Lock()
	serviceCache[cachekey] = services
	mx.Unlock()
	return services, nil
}

func sendNotificationEmail(notificationSettings []notifications.Setting, origin, destination, date, flightNumber string, message, body string) error {
	subs := notifications.GroupByRoute(notificationSettings)[origin][destination]
	heading := fmt.Sprintf("✈️ %s-%s/%s", origin, destination, date)
	baseURL := os.Getenv("BASE_URL")

	link := fmt.Sprintf("https://booking.wingo.com/es/search/%s/%s/%s/1/0/0/1/COP/0/0", origin, destination, date)
	linkHistory := fmt.Sprintf("%s/history?origin=%s&destination=%s&date=%s&flightNumber=%s", baseURL,
		url.QueryEscape(origin), url.QueryEscape(destination), url.QueryEscape(date), url.QueryEscape(flightNumber))

	for _, sub := range subs {
		if sub.Date != date {
			continue
		}
		cancelSubscriptionLink := fmt.Sprintf("%s/.netlify/functions/cancel_subscription?uid=%s", baseURL, sub.UID)

		fmt.Println("["+sub.Email+"]:", heading, message)
		data := map[string]interface{}{
			"Message":                message,
			"Link":                   link,
			"LinkHistory":            linkHistory,
			"CancelSubscriptionLink": cancelSubscriptionLink,
		}
		err := email.SendMessageWithText(context.Background(), heading, message, email.TplPriceChange, data, strings.Split(sub.Email, ",")...)
		if err != nil {
			return err
		}

		if len(sub.PhoneNumber) != 0 {
			err = whatsapp.SendMessage(sub.PhoneNumber, heading, message)
			if err != nil {
				log.Print(err)
			}
		}
	}

	return nil
}

func sendNewFlightNotification(notificationSettings []notifications.Setting, origin, destination, date, flightNumber string, price float64) error {
	subject := fmt.Sprintf("Precio actual: %s.", formatMoney(price))

	return sendNotificationEmail(notificationSettings, origin, destination, date, flightNumber, subject, "")
}

func sendPriceChangedNotification(notificationSettings []notifications.Setting, origin, destination, date, flightNumber string, oldPrice, newPrice float64) error {
	emoji := "↗️"
	accion := "SUBIÓ"
	if oldPrice > newPrice {
		emoji = "↘️"
		accion = "BAJÓ"
	}

	subject := fmt.Sprintf("%s El precio %s a %s (desde %s).", emoji, accion, formatMoney(newPrice), formatMoney(oldPrice))

	return sendNotificationEmail(notificationSettings, origin, destination, date, flightNumber, subject, "")
}

func sendNotAvailableNotification(notificationSettings []notifications.Setting, origin, destination, date, flightNumber string, lastPrice float64) error {
	subject := "El vuelo ya NO está disponible."
	return sendNotificationEmail(notificationSettings, origin, destination, date, flightNumber, subject, "")
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

func calculatePrice(vuelo wingo.Vuelo, services []wingo.Service) float64 {
	adminFares := wingo.GetAdminFares(wingo.ServiceQuote{
		Services: services,
	})

	return wingo.GetBundlePrice(wingo.OriginalPlanName, vuelo, adminFares)
}

func findFlight(savedFlights flightsMap, origin, destination, date, flightNumber string) (vueloArchivado, bool) {
	for savedDate, flights := range savedFlights[origin][destination] {
		if savedDate != date {
			continue
		}

		for _, flight := range flights {
			if flight.FlightNumber == flightNumber {
				return flight, true
			}
		}
	}

	return vueloArchivado{}, false
}

func processFlight(notificationSettings []notifications.Setting, savedFlights flightsMap, date, origin, destination string, flight vueloArchivado) error {
	previous, previousFound := findFlight(savedFlights, origin, destination, date, flight.FlightNumber)

	price := calculatePrice(flight.Vuelo, flight.Services)
	// 1. Antes NO disponible y ahora disponible?
	if !previousFound {
		err := sendNewFlightNotification(notificationSettings, origin, destination, date, flight.FlightNumber, price)
		if err != nil {
			return err
		}
	} else {
		savedPrice := calculatePrice(previous.Vuelo, previous.Services)
		// 2. Antes disponible y ahora diferente precio?
		if price != savedPrice {
			err := sendPriceChangedNotification(notificationSettings, origin, destination, date, flight.FlightNumber, savedPrice, price)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func processUnavailableFlights(notificationSettings []notifications.Setting, savedFlights flightsMap, actualFlights flightsMap) error {
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
						err := sendNotAvailableNotification(notificationSettings, origin, destination, date, savedFlight.FlightNumber, savedPrice)
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

func processUnavailableFlightsForSubs(subs []notifications.Setting, savedFlights flightsMap, actualFlights flightsMap) error {
	subsByRoute := notifications.GroupByRoute(subs)

	// 3. Antes disponible y ahora NO disponible?
	for origin, originSubs := range subsByRoute {
		for destination, destinationSubs := range originSubs {
			for _, sub := range destinationSubs {
				date := sub.Date
				savedFlights := savedFlights[origin][destination][date]
				for _, savedFlight := range savedFlights {
					_, actualFound := findFlight(actualFlights, origin, destination, date, savedFlight.FlightNumber)
					if !actualFound {
						flightpath := filepath.Join(outdir, origin, destination, date, savedFlight.FlightNumber+".json")
						_ = os.Remove(flightpath)

						savedPrice := calculatePrice(savedFlight.Vuelo, savedFlight.Services)
						err := sendNotAvailableNotification(subs, origin, destination, date, savedFlight.FlightNumber, savedPrice)
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

func processSchedule(notificationSettings []notifications.Setting,
	savedFlights flightsMap, date, origin, destination string, flight vueloArchivado) error {
	previous, previousFound := findFlight(savedFlights, origin, destination, date, flight.FlightNumber)

	price := calculatePrice(flight.Vuelo, flight.Services)
	// 1. Antes NO disponible y ahora disponible?
	if !previousFound {
		err := sendNewFlightNotification(notificationSettings, origin, destination, date, flight.FlightNumber, price)
		if err != nil {
			return err
		}
	} else {
		savedPrice := calculatePrice(previous.Vuelo, previous.Services)
		// 2. Antes disponible y ahora diferente precio?
		if price != savedPrice {
			err := sendPriceChangedNotification(notificationSettings, origin, destination, date, flight.FlightNumber, savedPrice, price)
			if err != nil {
				return err
			}
		}
	}

	return nil
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
	// fmt.Println(startDate, endDate, endDate.Sub(startDate).Hours())
	daysAfter := int(endDate.Sub(startDate).Hours() / 24)
	// fmt.Printf("Start date %s-%s: %s", origin, destination, date.Format(startDate))
	// fmt.Printf(" | Days %s-%s: %d\n", origin, destination, daysAfter)
	flightsInformation, err := client.GetInformationFlightsMonthly(origin, destination, date.Format(startDate), daysAfter)
	if err != nil {
		log.Fatal(err)
	}

	tasks := convertToTasks(flightsInformation, origin, destination)
	fmt.Printf("Got flightsInformation %s/%s (%s - %s): %d\n", origin, destination, startDate, endDate, len(tasks))
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

func containsString(elms []string, s string) bool {
	for _, elm := range elms {
		if elm == s {
			return true
		}
	}
	return false
}

func processNotificationSettings(client *wingo.Client, settings []notifications.Setting,
	savedFlights flightsMap, startDate, stopDate time.Time) error {

	type getFlightScheduleTask struct{ origin, destination string }
	getFlightsScheduleTasksChan := make(chan getFlightScheduleTask, maxWorkers)

	var flightsInformation []wingo.FlightInformation

	wg := syncbits.Workgroup(func() {
		for t := range getFlightsScheduleTasksChan {
			information, err := client.GetFlightScheduleInformation(t.origin, t.destination, date.Format(startDate), date.Format(stopDate))
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

		if !containsString(routes[setting.Origin], setting.Destination) {
			routes[setting.Origin] = append(routes[setting.Origin], setting.Destination)
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

		date := date.Format(fecha)
		addFlightToMap(actualFlights, flightInf.Origin, flightInf.Destination, date, flight)
		processSchedule(settings, savedFlights, date, flightInf.Origin, flightInf.Destination, flight)
	}

	// return processUnavailableSchedules(emailservice, settings, savedFlights, actualFlights)
	return nil
}

type getInformationFlightsTask struct {
	origin, destination string
	startDate, endDate  time.Time
	subs                []notifications.Setting
}

func sendRoutesPerDate(origin, destination string, startDate, stopDate time.Time, getInformationFlightsChan chan<- getInformationFlightsTask, subs []notifications.Setting) {
	for startDate.Before(stopDate) || startDate.Equal(stopDate) {
		endDate := startDate.AddDate(0, 1, 0)
		if endDate.After(stopDate) {
			endDate = stopDate.AddDate(0, 0, 1)
		}
		getInformationFlightsChan <- getInformationFlightsTask{
			origin:      origin,
			destination: destination,
			startDate:   startDate,
			endDate:     endDate,
			subs:        subs,
		}
		startDate = endDate
	}
}

func loadRoutes(client *wingo.Client, path string) ([]wingo.Route, []wingo.Route, error) {
	var savedRoutes struct {
		Response []wingo.Route `json:"response"`
	}
	err := loadFromFile(path, &savedRoutes)
	if err != nil {
		log.Println(err)
	}

	routes, err := client.GetRoutesWithCache(path)
	return savedRoutes.Response, routes, err
}

func main() {
	starttime := time.Now()
	defer func() {
		fmt.Fprintln(os.Stderr, "Duración:", time.Since(starttime))
	}()

	months, err := strconv.Atoi(os.Getenv("WINGO_MONTHS"))
	if err != nil {
		months = 6
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

	subs, err := notifications.LoadAllSettings()
	if err != nil {
		log.Fatal(err)
	}
	subs = notifications.FilterConfirmed(subs)
	subs = notifications.FilterBetweenDates(subs, startDate, stopDate)
	if len(subs) == 0 {
		fmt.Println("we just got nothing to do")
		return
	}

	fast := len(os.Args) >= 2 && os.Args[1] == "fast"
	runSubs := len(os.Args) < 2 || (len(os.Args) >= 2 && os.Args[1] == "subs")

	logger.Println("Cargando rutas guardadas")
	routesDir := os.Getenv("ROUTES_DIR")
	if routesDir == "" {
		routesDir = "./"
	}
	savedRoutes, routes, err := loadRoutes(client, routesDir+"routes.json")
	if err != nil {
		log.Fatal(err)
	}

	logger.Println("Cargando vuelos guardados")
	savedFlights, err := loadSavedFlights(routesDir, savedRoutes, startDate, stopDate)
	if err != nil {
		log.Fatal(err)
	}

	logger.Println("Subscriptions count:", len(subs))
	if fast {
		processNotificationSettings(client, subs, savedFlights, startDate, stopDate)
		return
	}

	logger.Println("Rutas guardadas:", len(savedRoutes))
	logger.Println("Routes from API:", len(routes))
	logger.Println("Vuelos guardados:", len(savedFlights))

	if len(routes) == 0 {
		logger.Println("Routes not found")
		return
	}

	var getPriceTasks []getPriceTask

	routesCount := 0
	getInformationFlightsChan := make(chan getInformationFlightsTask, maxWorkers)

	wg := syncbits.Workgroup(func() {
		for t := range getInformationFlightsChan {
			tasks := getInformationFlightsMonthly(client, t.origin, t.destination, t.startDate, t.endDate)
			for _, t2 := range tasks {
				if err := checkDate(t.startDate, t.endDate, t2.fecha); err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
			}
			if t.subs != nil {
				for _, pt := range tasks {
					found := false
					for _, sub := range t.subs {
						if sub.Date == pt.fecha {
							found = true
							break
						}
					}
					if found {
						getPriceTasks = append(getPriceTasks, pt)
					}
				}
			} else {
				getPriceTasks = append(getPriceTasks, tasks...)
			}
		}
	}, maxWorkers)

	if runSubs {
		subsByRoute := notifications.GroupByRoute(subs)

		for origin, originSubs := range subsByRoute {
			for destination, subs := range originSubs {
				var routeStartDate, routeStopDate *time.Time
				for _, sub := range subs {
					d := date.MustParse(sub.Date)

					if routeStartDate == nil || d.Before(*routeStartDate) {
						routeStartDate = &d
					}

					if routeStopDate == nil || d.After(*routeStopDate) {
						routeStopDate = &d
					}
				}

				if routeStartDate != nil && routeStopDate != nil {
					fmt.Println(origin, "=>", destination)
					routesCount++
					sendRoutesPerDate(origin, destination, *routeStartDate, *routeStopDate, getInformationFlightsChan, subs)
				} else {
					fmt.Println("both dates are nil - should not happen")
				}
			}
		}
	} else {
		for _, origin := range routes {
			for _, destination := range origin.Routes {
				fmt.Println(origin.Name, "=>", destination.Name)
				routesCount++
				sendRoutesPerDate(origin.Code, destination.Code, startDate, stopDate, getInformationFlightsChan, nil)
			}
		}
	}
	close(getInformationFlightsChan)

	fmt.Println("--------------------------------------")
	fmt.Println("Waiting to load", routesCount, "routes information")
	fmt.Println("--------------------------------------")
	wg.Wait()

	fmt.Println("------------------------------------")
	fmt.Println("Finished loading routes information")
	fmt.Println("------------------------------------")

	actualFlights := flightsMap{}
	actualFlightsMutex := new(sync.Mutex)
	archiveTaskChan := make(chan archiveTask, maxWorkers)
	wgArchive := syncbits.Workgroup(func() {
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

			err = processFlight(subs, savedFlights, task.fecha, task.origin, task.destination, flight)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}, maxWorkers)

	getPriceTaskChan := make(chan getPriceTask, maxWorkers)
	threadsG := syncbits.Workgroup(func() {
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

	if runSubs {
		err = processUnavailableFlightsForSubs(subs, savedFlights, actualFlights)
	} else {
		err = processUnavailableFlights(subs, savedFlights, actualFlights)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
