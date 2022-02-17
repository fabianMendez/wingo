package wingo

import (
	"bytes"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/fabianMendez/wingo/pkg/date"
	"golang.org/x/net/proxy"
)

const (
	AdminFareCode    = "BFEE"
	OriginalPlanName = "BASIC"
)

type Client struct {
	httpClient        *http.Client
	log               *log.Logger
	aditionalHeaders  map[string]string
	RequestCount      int
	requestCountMutex *sync.Mutex
}

func NewClient(logger *log.Logger) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = proxy.Dial
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	httpClient := &http.Client{Transport: transport, Timeout: time.Second * 15}

	return &Client{
		httpClient:        httpClient,
		requestCountMutex: new(sync.Mutex),
		log:               logger,
		aditionalHeaders: map[string]string{
			"User-Agent":      "Mozilla/5.0 (X11; Linux x86_64; rv:90.0) Gecko/20100101 Firefox/90.0",
			"Origin":          "https://booking.wingo.com",
			"Referer":         "https://booking.wingo.com/",
			"Accept-Language": "en-US,en;q=0.5",
		},
	}
}

func (c *Client) request(method, u string, body io.Reader, headers map[string]string) (*http.Response, error) {
	c.requestCountMutex.Lock()
	c.RequestCount++
	c.requestCountMutex.Unlock()

	c.log.Println(method, u)

	maxRetries := 10
	boff := backoff.NewExponentialBackOff()
	boff.InitialInterval = 5 * time.Second
	boff.Reset()

	var resp *http.Response

	for i := 1; i <= maxRetries; i++ {
		if i > 1 {
			boffDuration := boff.NextBackOff()
			c.log.Println("*** Backoff retry", i, ":", boffDuration)
			time.Sleep(boffDuration)
		}

		req, err := http.NewRequest(method, u, body)
		if err != nil {
			return nil, fmt.Errorf("could not create request: %w", err)
		}

		for headerKey, headerValue := range c.aditionalHeaders {
			req.Header.Add(headerKey, headerValue)
		}

		for headerKey, headerValue := range headers {
			req.Header.Add(headerKey, headerValue)
		}

		resp, err = c.httpClient.Do(req)
		if err != nil {
			if i != maxRetries {
				fmt.Fprintf(os.Stderr, "retry %d: %v\n", i, err)
				continue
			}
			return nil, fmt.Errorf("could not send request: %w", err)
		}

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotModified {
			if i != maxRetries {
				fmt.Fprintf(os.Stderr, "retry %d: %v\n", i, resp.StatusCode)
				continue
			}
			return nil, fmt.Errorf("request failed: %s %s - %s", method, u, resp.Status)
		}

		break
	}

	return resp, nil
}

func (c *Client) requestJSON(method, u string, body io.Reader, v interface{}, headers map[string]string) error {
	resp, err := c.request(method, u, body, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return fmt.Errorf("could not decode response: %w", err)
	}

	return nil
}

func (c *Client) requestJSONWithCache(method, u string, body io.Reader, v interface{}, path string, headers map[string]string) error {
	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if err == nil {
		if headers == nil {
			headers = make(map[string]string)
		}
		etag := wetag(content)
		headers["If-None-Match"] = etag
	}

	resp, err := c.request(method, u, body, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		err = json.Unmarshal(content, v)
	} else {
		var f *os.File
		f, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return err
		}
		defer func() {
			_ = f.Close()
		}()

		r := io.TeeReader(resp.Body, f)
		err = json.NewDecoder(r).Decode(v)
	}

	if err != nil {
		return fmt.Errorf("could not decode response: %w", err)
	}

	return nil
}

func (c *Client) GetInformationFlightsMonthly(origin, destination, startDate string, daysAfter int) (FlightsInformation, error) {
	startTime, err := date.Parse(startDate)
	if err != nil {
		return FlightsInformation{}, err
	}

	endTime := startTime.AddDate(0, 0, daysAfter)
	originEndDate := date.Format(endTime)

	parameters := url.Values{}
	parameters.Set("origin", origin)
	parameters.Set("originStartDate", startDate)
	parameters.Set("originEndDate", originEndDate)
	parameters.Set("originDaysBefore", "0")
	parameters.Set("originDaysAfter", strconv.Itoa(daysAfter))
	parameters.Set("destination", destination)
	parameters.Set("destinationStartDate", "Fecha inválida")
	parameters.Set("destinationEndDate", "Fecha inválida")
	parameters.Set("destinationDaysBefore", "0")
	parameters.Set("destinationDaysAfter", "NaN")
	parameters.Set("currency", "COP")
	parameters.Set("adultNumber", "1")
	parameters.Set("childNumber", "0")
	parameters.Set("infantNumber", "0")
	parameters.Set("flightType", "1")
	parameters.Set("securityToken", "")
	parameters.Set("iataNumber", "")
	parameters.Set("userAgent", "IBE")
	parameters.Set("promoCode", "")
	parameters.Set("currentCurrency", "undefined")
	parameters.Set("multiCurrency", "false")
	parameters.Set("languageId", "1")

	u := "https://routes-api.wingo.com/v1/getInformationFlightsMonthly?" + parameters.Encode()

	var response struct {
		Response FlightsInformation `json:"response"`
	}

	err = c.requestJSON(http.MethodGet, u, nil, &response, nil)
	if err != nil {
		return FlightsInformation{}, err
	}

	return response.Response, nil
}

func (c *Client) RetrieveServiceQuotes(flights []FlightService, token string) ([]ServiceQuote, error) {
	u := "https://ancillaries-api.wingo.com/v1/retrieveServiceQuotes"

	rb := struct {
		Currency      string          `json:"currency"`
		Flights       []FlightService `json:"flights"`
		Token         string          `json:"token"`
		Module        string          `json:"module"`
		Multicurrency bool            `json:"multicurrency"`
		LanguageID    int64           `json:"languageId"`
	}{
		Currency:      "COP",
		Token:         token,
		Module:        "2",
		Multicurrency: false,
		LanguageID:    1,
		Flights:       flights,
	}

	var response struct {
		Response []ServiceQuote `json:"response"`
	}

	bodyBytes, err := json.Marshal(rb)
	if err != nil {
		return nil, fmt.Errorf("could not encode json: %w", err)
	}

	headers := map[string]string{"Content-Type": "application/json"}
	fmt.Println(string(bodyBytes))
	err = c.requestJSON(http.MethodPost, u, bytes.NewReader(bodyBytes), &response, headers)
	if err != nil {
		return nil, err
	}

	return response.Response, nil
}

func SumarPrecioCalendario(flight Vuelo) float64 {
	if len(flight.InfoFares) > 0 {
		var taxes float64
		fare := flight.InfoFares[0].FareAdult
		for _, tax := range fare.ApplicableTaxes {
			taxes += tax.TaxAmount
		}
		return fare.FareAmount + taxes
	}
	return 0
}

func GetBundlePrice(bundle string, flight Vuelo, adminFares float64) float64 {
	var bundlePrice float64
	// if "BASIC" != bundle {
	// 	panic("bundlePrice = getBundlePriceByTitle(bundle)")
	// }
	return SumarPrecioCalendario(flight) + adminFares + bundlePrice
}

func GetAdminFares(serviceQuote ServiceQuote) float64 {
	for _, service := range serviceQuote.Services {
		if service.CodeType == AdminFareCode {
			var taxes float64

			for _, tax := range service.Taxes {
				taxes += tax.TaxAmount
			}

			return service.Amount + taxes
		}
	}
	return 0
}

func (c *Client) GetRoutes() ([]Route, error) {
	u := "https://routes-api.wingo.com/v1/completeroute/es"

	var response struct {
		Response []Route `json:"response"`
	}

	err := c.requestJSON(http.MethodGet, u, nil, &response, nil)
	if err != nil {
		return nil, err
	}

	return response.Response, nil
}

func wetag(buf []byte) string {
	if len(buf) == 0 {
		// fast-path empty body
		return `W/"0-0"`
	}

	hashBuf := sha1.Sum(buf)
	hash := base64.StdEncoding.EncodeToString(hashBuf[:])

	return fmt.Sprintf(`W/"%x-%s"`, len(buf), hash[0:27])
}

func (c *Client) GetRoutesWithCache(path string) ([]Route, error) {
	u := "https://routes-api.wingo.com/v1/completeroute/es"

	var response struct {
		Response []Route `json:"response"`
	}

	err := c.requestJSONWithCache(http.MethodGet, u, nil, &response, path, nil)
	if err != nil {
		return nil, err
	}

	return response.Response, nil
}

// now
// now + 10 months
func (c *Client) GetFlightScheduleInformation(origin, destination, startDate, endDate string) (FlightScheduleInformation, error) {
	parameters := url.Values{}
	parameters.Set("carrierCode", "P5")
	parameters.Set("searchType", "")
	parameters.Set("origin", origin)
	parameters.Set("destination", destination)
	parameters.Set("startDate", startDate)
	parameters.Set("endDate", endDate)
	parameters.Set("flightNumber", "0")
	parameters.Set("includedCancelled", "false")
	u := "https://routes-api.wingo.com/v1/scheduleinformation?" + parameters.Encode()

	var response struct {
		Response FlightScheduleInformation `json:"response"`
	}

	err := c.requestJSON(http.MethodGet, u, nil, &response, nil)
	return response.Response, err
}
