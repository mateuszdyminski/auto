package model

import "time"

// FlightCrash holds info about the historical flight crash.
type FlightCrash struct {
	ID           string    `json:"id,omitempty"`
	Date         time.Time `json:"date,omitempty"`
	Location     string    `json:"location,omitempty"`
	Operator     string    `json:"operator,omitempty"`
	FlightNo     string    `json:"flightNo,omitempty"`
	Route        string    `json:"route,omitempty"`
	AircraftType string    `json:"aircraftType,omitempty"`
	Registration string    `json:"registration,omitempty"`
	SerialNumber string    `json:"serialNumber,omitempty"`
	Aboard       Aboard    `json:"aboard,omitempty"`
	Fatalities   Aboard    `json:"fatalities,omitempty"`
	Ground       int       `json:"ground,omitempty"`
	Summary      string    `json:"summary,omitempty"`
	LocationGPS  *Location `json:"locationGPS,omitempty"`
	Score        *float64  `json:"score,omitempty"`
}

// Aboard holds information about the people on the plane.
type Aboard struct {
	Total      int `json:"total,omitempty"`
	Crew       int `json:"crew,omitempty"`
	Passengers int `json:"passengers,omitempty"`
}

// Location holds inforamtion
type Location struct {
	Longitude float64 `json:"lon,omitempty"`
	Latitude  float64 `json:"lat,omitempty"`
}
