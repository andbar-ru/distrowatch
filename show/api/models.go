package main

import (
	"encoding/json"
	"strconv"
	"time"
)

type dateInt int

// Distr represents db table "distrs".
type Distr struct {
	Name       string  `json:"name"`
	Count      int     `json:"count"`
	LastUpdate dateInt `db:"last_update" json:"lastUpdate"`
}

func (d dateInt) MarshalJSON() ([]byte, error) {
	date, err := time.Parse("20060102", strconv.Itoa(int(d)))
	if err != nil {
		return nil, err
	}
	return json.Marshal(date.Format("2006-01-02"))
}
