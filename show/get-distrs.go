package show

import (
	"github.com/andbar-ru/distrowatch"
)

// Distr describes one distribution.
type Distr struct {
	Name       string
	Count      int
	LastUpdate int
}

// GetDistrs returns distribution list from database.
func GetDistrs() ([]*Distr, error) {
	var db, err = distrowatch.GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Query
	query := "SELECT name, count, last_update FROM distrs ORDER BY count DESC, last_update ASC"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Collect distrs.
	distrs := make([]*Distr, 0)
	for rows.Next() {
		d := new(Distr)
		err := rows.Scan(&d.Name, &d.Count, &d.LastUpdate)
		if err != nil {
			return nil, err
		}
		distrs = append(distrs, d)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return distrs, nil
}
