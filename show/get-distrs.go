package show

// Distr describes one distribution.
type Distr struct {
	Name       string
	Count      int
	LastUpdate int
}

// GetDistrs returns distribution list from database.
func GetDistrs(database string) ([]Distr, error) {
	var db, err = getDB(database)
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
	distrs := []Distr{}
	for rows.Next() {
		var d Distr
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
