package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

/*
	HTTP GET /api/v1/production
	{
		"wattHoursToday": 5585,
		"wattHoursSevenDays": 118960,
		"wattHoursLifetime": 7021969,
		"wattsNow": 3860
	}
*/

type stats struct {
	WattHoursToday int `json:"wattHoursToday"`
	WattsNow       int `json:"wattsNow"`
	// TODO: wattHoursSevenDays
	WattHoursLifetime int `json:"wattHoursLifetime"`
}

func fetchStats(addr string) (stats, error) {
	resp, err := http.Get("http://" + addr + "/api/v1/production")
	if err != nil {
		return stats{}, fmt.Errorf("fetching Enphase data: %v", err)
	}
	raw, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return stats{}, fmt.Errorf("reading Enphase data: %v", err)
	}
	if resp.StatusCode != 200 {
		return stats{}, fmt.Errorf("got non-200 status code %d from Enphase", resp.StatusCode)
	}

	var st stats
	if err := json.Unmarshal(raw, &st); err != nil {
		return stats{}, fmt.Errorf("parsing Enphase data: %v", err)
	}
	return st, nil
}
