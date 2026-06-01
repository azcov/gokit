package nominatim

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/azcov/gokit/geo"
)

var _ geo.Geocoder = (*Nominatim)(nil)

const (
	defaultBaseURL = "https://nominatim.openstreetmap.org"
	defaultTimeout = 10 * time.Second
	// Nominatim usage policy: include a valid User-Agent.
	defaultUserAgent = "gokit/1.0"
)

// Nominatim is a free geocoder backed by OpenStreetMap data.
// No API key required. Respect the usage policy: max 1 req/sec.
type Nominatim struct {
	cfg    geo.Config
	client *http.Client
}

func New(cfg geo.Config) *Nominatim {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	if cfg.Language == "" {
		cfg.Language = "en"
	}
	return &Nominatim{
		cfg:    cfg,
		client: &http.Client{Timeout: defaultTimeout},
	}
}

type searchResult struct {
	PlaceID     int    `json:"place_id"`
	DisplayName string `json:"display_name"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	Address     struct {
		Road        string `json:"road"`
		City        string `json:"city"`
		Town        string `json:"town"`
		Village     string `json:"village"`
		State       string `json:"state"`
		Country     string `json:"country"`
		PostalCode  string `json:"postcode"`
	} `json:"address"`
}

func (n *Nominatim) Geocode(ctx context.Context, address string) ([]geo.Result, error) {
	params := url.Values{
		"q":              {address},
		"format":         {"json"},
		"addressdetails": {"1"},
		"accept-language": {n.cfg.Language},
		"limit":          {"5"},
	}
	var raw []searchResult
	if err := n.get(ctx, "/search", params, &raw); err != nil {
		return nil, err
	}
	results := make([]geo.Result, 0, len(raw))
	for _, r := range raw {
		results = append(results, toResult(r))
	}
	return results, nil
}

func (n *Nominatim) ReverseGeocode(ctx context.Context, loc geo.Location) (*geo.Result, error) {
	params := url.Values{
		"lat":             {strconv.FormatFloat(loc.Lat, 'f', 7, 64)},
		"lon":             {strconv.FormatFloat(loc.Lon, 'f', 7, 64)},
		"format":          {"json"},
		"addressdetails":  {"1"},
		"accept-language": {n.cfg.Language},
	}
	var raw searchResult
	if err := n.get(ctx, "/reverse", params, &raw); err != nil {
		return nil, err
	}
	r := toResult(raw)
	return &r, nil
}

func (n *Nominatim) Close() error { return nil }

func (n *Nominatim) get(ctx context.Context, path string, params url.Values, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		n.cfg.BaseURL+path+"?"+params.Encode(), nil)
	if err != nil {
		return fmt.Errorf("nominatim: build request: %w", err)
	}
	req.Header.Set("User-Agent", defaultUserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("nominatim: http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("nominatim: unexpected status %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("nominatim: decode: %w", err)
	}
	return nil
}

func toResult(r searchResult) geo.Result {
	lat, _ := strconv.ParseFloat(r.Lat, 64)
	lon, _ := strconv.ParseFloat(r.Lon, 64)
	city := r.Address.City
	if city == "" {
		city = r.Address.Town
	}
	if city == "" {
		city = r.Address.Village
	}
	return geo.Result{
		Name: r.DisplayName,
		Location: geo.Location{Lat: lat, Lon: lon},
		Address: geo.Address{
			Street:     r.Address.Road,
			City:       city,
			State:      r.Address.State,
			Country:    r.Address.Country,
			PostalCode: r.Address.PostalCode,
			Formatted:  r.DisplayName,
		},
	}
}
