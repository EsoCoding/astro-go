package geocode

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const userAgent = "astro-go/0.1 (desktop astrology application)"

type Result struct {
	DisplayName string
	Latitude    float64
	Longitude   float64
}

type nominatimResult struct {
	DisplayName string `json:"display_name"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
}

type NominatimClient struct {
	baseURL string
	client  *http.Client
}

func NewNominatimClient() NominatimClient {
	return NominatimClient{
		baseURL: "https://nominatim.openstreetmap.org",
		client: &http.Client{
			Timeout: 8 * time.Second,
		},
	}
}

func (c NominatimClient) Lookup(query string) (Result, error) {
	if strings.TrimSpace(query) == "" {
		return Result{}, fmt.Errorf("enter a location name first")
	}
	endpoint := fmt.Sprintf("%s/search?format=jsonv2&limit=1&q=%s", c.baseURL, url.QueryEscape(query))
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("geocode lookup failed: %s", resp.Status)
	}

	var rows []nominatimResult
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return Result{}, err
	}
	if len(rows) == 0 {
		return Result{}, fmt.Errorf("no result found for %q", query)
	}

	latitude, err := strconv.ParseFloat(rows[0].Lat, 64)
	if err != nil {
		return Result{}, err
	}
	longitude, err := strconv.ParseFloat(rows[0].Lon, 64)
	if err != nil {
		return Result{}, err
	}

	displayName := rows[0].DisplayName
	parts := strings.Split(displayName, ",")
	if len(parts) >= 3 {
		displayName = strings.TrimSpace(parts[0]) + ", " + strings.TrimSpace(parts[len(parts)-1])
	}

	return Result{
		DisplayName: displayName,
		Latitude:    latitude,
		Longitude:   longitude,
	}, nil
}
