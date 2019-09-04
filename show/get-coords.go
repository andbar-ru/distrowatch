package show

import (
	"database/sql"

	"github.com/andbar-ru/distrowatch"
)

// Coords desribes the current point and vector from the previous point.
type Coords struct {
	Latitude       float64
	Longitude      float64
	LatitudeDelta  float64
	LongitudeDelta float64
}

// GetCoords returns current coordinates from database.
func GetCoords() (*Coords, error) {
	var db, err = distrowatch.GetDB()
	if err != nil {
		return &Coords{}, err
	}
	defer db.Close()

	var latitude, longitude float64
	var latitudeDiff, longitudeDiff float64
	var latitudeTrend, longitudeTrend int
	var latitudeDelta, longitudeDelta float64
	err = db.QueryRow("SELECT latitude, longitude, latitude_diff, longitude_diff, latitude_trend, longitude_trend FROM coords ORDER BY date DESC LIMIT 1").Scan(&latitude, &longitude, &latitudeDiff, &longitudeDiff, &latitudeTrend, &longitudeTrend)
	if err != nil {
		if err == sql.ErrNoRows {
			latitude = 60.0
			latitudeDelta = 0.0
			longitude = 30.0
			longitudeDelta = 0.0
		} else {
			return &Coords{}, err
		}
	} else {
		latitudeDelta = latitudeDiff * float64(latitudeTrend)
		longitudeDelta = longitudeDiff * float64(longitudeTrend)
	}

	return &Coords{latitude, longitude, latitudeDelta, longitudeDelta}, nil
}
