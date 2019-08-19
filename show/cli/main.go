package main

import (
	"fmt"
	"math"
	"os"
	"path"

	"github.com/andbar-ru/distrowatch/show"
)

var (
	database = path.Join(os.Getenv("HOME"), "Images/distrs/db.sqlite3")
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	// Print coordinates in one line.
	coords, err := show.GetCoords(database)
	check(err)
	fmt.Printf("Coordinates: %.4f (%+.4f) %.4f (%+.4f)\n\n",
		coords.Latitude, coords.LatitudeDelta, coords.Longitude, coords.LongitudeDelta)

	// Print distr stats in a table.
	distrs, err := show.GetDistrs(database)
	check(err)
	// Figure out table parameters.
	var nameFieldSize, countFieldSize int
	var newestLastUpdate, oldestLastUpdate int
	oldestLastUpdate = 1e8 // certainly greater than any date
	for _, distr := range distrs {
		if len(distr.Name) > nameFieldSize {
			nameFieldSize = len(distr.Name)
		}
		if distr.LastUpdate > newestLastUpdate {
			newestLastUpdate = distr.LastUpdate
		}
		if distr.LastUpdate < oldestLastUpdate {
			oldestLastUpdate = distr.LastUpdate
		}
	}
	countFieldSize = len(fmt.Sprint(distrs[0].Count)) // number of digits

	// Difference between adjacent distrs to define leaders.
	diff := int(math.Ceil(float64(distrs[0].Count) / 10))

	// Define leaders from the end.
	lastLeaderIndex := -1 // no leaders by default
	for i := len(distrs) - 1; i > 0; i-- {
		if distrs[i-1].Count-distrs[i].Count >= diff {
			lastLeaderIndex = i - 1
			break
		}
	}

	// Print table.
	numberFieldSize := len(fmt.Sprint(len(distrs)))
	for i, distr := range distrs {
		boldness := 0
		color := 37
		if i <= lastLeaderIndex {
			boldness = 1
		}
		fmt.Printf("\x1b[%d;%dm", boldness, color)
		fmt.Printf("%*d. %-*s %*d %d", numberFieldSize, i+1, nameFieldSize, distr.Name, countFieldSize, distr.Count, distr.LastUpdate)
		if distr.LastUpdate == newestLastUpdate {
			fmt.Print(" newest")
		}
		if distr.LastUpdate == oldestLastUpdate {
			fmt.Print(" oldest")
		}
		fmt.Print("\x1b[0m")
		fmt.Println()
	}
}
