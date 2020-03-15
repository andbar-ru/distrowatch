package main

// Distr represents db table "distrs".
type Distr struct {
	Name       string `json:"name"`
	Count      int    `json:"count"`
	LastUpdate int    `db:"last_update" json:"last_update"`
}
