package main

import (
	"fmt"
	"html/template"
	"math"
	"net/http"

	"github.com/andbar-ru/distrowatch/show"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Distr appends some fields to show.Distr.
type Distr struct {
	*show.Distr
	Number        int
	Leader        bool
	Newest        bool
	Oldest        bool
	LastUpdateStr string
}

func index(writer http.ResponseWriter, request *http.Request) {
	coords, err := show.GetCoords()
	check(err)

	averageColor, err := show.GetLastDistrImageAverageColor()
	check(err)
	var averageColorStr string
	if averageColor.A == 0xff {
		averageColorStr = fmt.Sprintf("#%02x%02x%02x", averageColor.R, averageColor.G, averageColor.B)
	} else {
		averageColorStr = fmt.Sprintf("#%02x%02x%02x%02x", averageColor.R, averageColor.G, averageColor.B, averageColor.A)
	}

	distrs, err := show.GetDistrs()
	check(err)
	tmplDistrs := make([]*Distr, 0, len(distrs))
	var newestLastUpdate, oldestLastUpdate int
	oldestLastUpdate = 1e8 // certainly greater than any date
	for i, distr := range distrs {
		if distr.LastUpdate > newestLastUpdate {
			newestLastUpdate = distr.LastUpdate
		}
		if distr.LastUpdate < oldestLastUpdate {
			oldestLastUpdate = distr.LastUpdate
		}
		d := &Distr{Distr: distr}
		d.Number = i + 1
		lastUpdateStr := fmt.Sprint(distr.LastUpdate)
		d.LastUpdateStr = fmt.Sprintf("%s-%s-%s", lastUpdateStr[:4], lastUpdateStr[4:6], lastUpdateStr[6:8])
		tmplDistrs = append(tmplDistrs, d)
	}

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

	for i, distr := range tmplDistrs {
		if distr.LastUpdate == newestLastUpdate {
			distr.Newest = true
		}
		if distr.LastUpdate == oldestLastUpdate {
			distr.Oldest = true
		}
		if i <= lastLeaderIndex {
			distr.Leader = true
		}
	}

	t := template.Must(template.New("index").ParseFiles("index.html"))

	type data struct {
		Coords       *show.Coords
		AverageColor string
		Distrs       []*Distr
	}
	dataI := &data{
		Coords:       coords,
		AverageColor: averageColorStr,
		Distrs:       tmplDistrs,
	}

	err = t.Execute(writer, dataI)
	check(err)
}

func main() {
	http.HandleFunc("/", index)
	http.ListenAndServe(":8080", nil)
}
