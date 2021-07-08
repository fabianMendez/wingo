package wingo

type FlightsInformation struct {
	ExchangeRate int64       `json:"exchangeRate"`
	AirportInfo  AirportInfo `json:"airportInfo"`
	VueloIda     []VueloIda  `json:"vueloIda"`
	Token        string      `json:"token"`
}

type AirportInfo struct {
	ToTerminal   string `json:"toTerminal"`
	FromTerminal string `json:"fromTerminal"`
}

type VueloIda struct {
	Fecha           string                  `json:"fecha"`
	InfoVuelo       InfoVuelo               `json:"infoVuelo"`
	ApplicableTaxes []VueloIdaApplicableTax `json:"applicableTaxes"`
}

type VueloIdaApplicableTax struct {
	TaxID           int64  `json:"taxId"`
	TaxDesc         string `json:"taxDesc"`
	TaxCurr         string `json:"taxCurr"`
	TaxOriginalDesc string `json:"taxOriginalDesc"`
}

type InfoVuelo struct {
	Fecha  string  `json:"fecha"`
	Vuelos []Vuelo `json:"vuelos"`
}

type Vuelo struct {
	DurationHours       int64      `json:"durationHours"`
	LogicalFlightID     int64      `json:"logicalFlightID"`
	AircraftDescription string     `json:"aircraftDescription"`
	CarrierCode         string     `json:"carrierCode"`
	InfoFares           []InfoFare `json:"infoFares"`
	DurationMins        int64      `json:"durationMins"`
	DepartureDate       string     `json:"departureDate"`
	ArrivalDate         string     `json:"arrivalDate"`
	FlightNumber        string     `json:"flightNumber"`
}

type InfoFare struct {
	IsInternational bool        `json:"isInternational"`
	FareAdult       Fare        `json:"fareAdult"`
	FareInfant      Fare        `json:"fareInfant"`
	LegDetail       []LegDetail `json:"legDetail"`
	FareChild       Fare        `json:"fareChild"`
}

type Fare struct {
	FareAmountOriginal     int64                `json:"fareAmountOriginal"`
	PassengerType          int64                `json:"passengerType"`
	FareID                 int64                `json:"fareID"`
	BaseFareAmount         int64                `json:"baseFareAmount"`
	BaseFareAmountOriginal int64                `json:"baseFareAmountOriginal"`
	ApplicableTaxes        []FaretApplicableTax `json:"applicableTaxes"`
	FareAmount             float64              `json:"fareAmount"`
	PromotionAmount        int64                `json:"promotionAmount"`
	SeatsAvailable         int64                `json:"seatsAvailable"`
}

type FaretApplicableTax struct {
	TaxID             int64   `json:"taxId"`
	TaxAmount         float64 `json:"taxAmount"`
	TaxAmountOriginal int64   `json:"taxAmountOriginal"`
}

type LegDetail struct {
	Pfid          int64  `json:"PFID"`
	DepartureDate string `json:"departureDate"`
}

type FlightService struct {
	Departure string `json:"departure"`
	From      string `json:"from"`
	To        string `json:"to"`
	FlightID  int64  `json:"flightId"`
}

type ServiceQuote struct {
	From          string    `json:"from"`
	To            string    `json:"to"`
	DepartureDate string    `json:"departureDate"`
	FlightID      int64     `json:"flightId"`
	Services      []Service `json:"services"`
}

type Service struct {
	CodeType          string  `json:"codeType"`
	Amount            float64 `json:"amount"`
	AmountOriginal    float64 `json:"amountOriginal"`
	Description       string  `json:"description"`
	AvalaibleQuantity int64   `json:"avalaibleQuantity"`
	Taxes             []Tax   `json:"taxes"`
	ServiceIDRadixx   int64   `json:"serviceIDRadixx"`
	SsrCategory       int64   `json:"ssrCategory"`
}

type Tax struct {
	TaxCode                string  `json:"taxCode"`
	TaxDescription         string  `json:"taxDescription"`
	TaxAmount              float64 `json:"taxAmount"`
	TaxAmountOriginal      float64 `json:"taxAmountOriginal"`
	TaxCurrencyCode        string  `json:"taxCurrencyCode"`
	TaxOriginalDescription string  `json:"taxOriginalDescription"`
}
