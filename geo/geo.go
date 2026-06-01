package geo

import "context"

type Location struct {
	Lat float64
	Lon float64
}

type Address struct {
	Street     string
	City       string
	State      string
	Country    string
	PostalCode string
	Formatted  string
}

type Result struct {
	Location Location
	Address  Address
	Name     string
}

type Config struct {
	APIKey   string `config:"api_key"`
	BaseURL  string `config:"base_url"`
	Language string `config:"language"`
}

type Geocoder interface {
	Geocode(ctx context.Context, address string) ([]Result, error)
	ReverseGeocode(ctx context.Context, loc Location) (*Result, error)
	Close() error
}
