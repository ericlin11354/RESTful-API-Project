package timeSeries

import (
	"time"
)

type TimeSeries struct {
	UID        string `json:"id"`
	Admin2     string `json:"Admin2"`
	Address1   string `json:"Province/State"`
	Address2   string `json:"Country/Region"`
	Population int    `json:"Population"`

	Confirmed map[time.Time]int `json:"Confirmed"`
	Death     map[time.Time]int `json:"Death"`
	Recovered map[time.Time]int `json:"Recovered"`
}
