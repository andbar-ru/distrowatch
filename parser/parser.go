package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/andbar-ru/distrowatch"
)

const (
	baseURL          = "https://distrowatch.com/"
	timeLayout       = "20060102"
	userAgent        = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.87 Safari/537.36 OPR/54.0.2952.46" // Opera 54
	initialLatitude  = 60.0
	initialLongitude = 30.0
	divider          = 10000
)

var (
	now         = time.Now()
	today       = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	todayYYMMDD = today.Format(timeLayout)
	client      = &http.Client{}
	distrCount  = 100
)

// Outcome stores outcome from baseURL.
type Outcome struct {
	distrName  string
	distrURL   string
	hpd        int
	next1HPD   int
	next1Trend int
	next2HPD   int
	next2Trend int
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func closeCheck(c io.Closer) {
	err := c.Close()
	check(err)
}

func checkResponse(response *http.Response) {
	if response.StatusCode != 200 {
		log.Fatalf("Status code error: %d %s", response.StatusCode, response.Status)
	}
}

func getResponse(url string) *http.Response {
	request, err := http.NewRequest("GET", url, nil)
	check(err)
	request.Header.Add("User-Agent", userAgent)
	response, err := client.Do(request)
	check(err)
	return response
}

func getOutcome() Outcome {
	// Get main page
	response := getResponse(baseURL)
	defer closeCheck(response.Body)
	checkResponse(response)

	// Parse the page and fetch first distribution, hits per day of which didn't change since yesterday.
	root, err := goquery.NewDocumentFromReader(response.Body)
	check(err)

	hpdTds := root.Find("td.phr3") // HPD: Hits Per Day (Column header)
	if hpdTds.Length() == 0 {
		log.Fatal("There is no tds with class phr3.")
	} else if hpdTds.Length() != distrCount {
		log.Printf("WARNING: number of tds with HPD is not %d, just %d", distrCount, hpdTds.Length())
		distrCount = hpdTds.Length()
	}

	// Find first td which has img with alt="=" and fill outcome.
	outcome := Outcome{}
	var equalFound, next1Filled bool
	hpdTds.EachWithBreak(func(index int, hpdTd *goquery.Selection) bool {
		img := hpdTd.ChildrenFiltered("img").First()
		// Every hpdTd must contain just one img.
		if img.Length() == 0 {
			log.Fatalf("td.phr3 with index %d has not an img.", index)
		}
		// Image must have the attribute 'alt'.
		alt, ok := img.Attr("alt")
		if !ok {
			log.Fatalf("img in td.phr3 with index %d has not attribute 'alt'", index)
		}

		if equalFound {
			hpd, err := strconv.Atoi(hpdTd.Text())
			check(err)
			var trend int
			switch alt {
			case "<":
				trend = -1
			case ">":
				trend = 1
			case "=":
				trend = 0
			default:
				log.Fatalf("unexpected alt %s: td.phr3 with index %d", alt, index)
			}
			if !next1Filled {
				outcome.next1HPD = hpd
				outcome.next1Trend = trend
				next1Filled = true
			} else {
				outcome.next2HPD = hpd
				outcome.next2Trend = trend
				return false
			}
			return true
		}

		if alt == "=" {
			equalFound = true
			hpd, err := strconv.Atoi(hpdTd.Text())
			check(err)
			outcome.hpd = hpd

			distributionTd := hpdTd.Prev()
			if !distributionTd.HasClass("phr2") {
				log.Fatalf("td.phr3 with index %d has previous sibling (distributionTd) with class name != 'phr2'.", index)
			}
			a := distributionTd.ChildrenFiltered("a").First()
			if a.Length() == 0 {
				log.Fatalf("td.phr3 with index %d has not an 'a' in previous sibling.", index)
			}
			outcome.distrName = a.Text()
			url, ok := a.Attr("href")
			if !ok {
				log.Fatalf("a in td.phr2 with index %d has not attribute 'href'", index)
			}
			if !strings.HasPrefix(url, "http") {
				url = baseURL + url
			}
			outcome.distrURL = url
		}
		return true
	})

	if outcome.distrName == "" {
		log.Fatal("Could not find distribution with img.alt == '='.")
	}

	return outcome
}

// updateDb updates or inserts count of distribution name in database.
func updateDb(db *sql.DB, outcome Outcome) {
	tx, err := db.Begin()
	check(err)
	_, err = tx.Exec("INSERT OR IGNORE INTO distrs (name, count, last_update) VALUES (?, 0, ?)", outcome.distrName, todayYYMMDD)
	check(err)
	_, err = tx.Exec("UPDATE distrs SET count = count + 1, last_update = ? WHERE name = ?", todayYYMMDD, outcome.distrName)
	check(err)
	_, err = tx.Exec("INSERT INTO distrs_daily (date, name, hpd) VALUES (?, ?, ?)", todayYYMMDD, outcome.distrName, outcome.hpd)
	check(err)

	// Move distrs that have been updated over year ago to the table `dropout`.
	todayYYMMDDint, err := strconv.Atoi(todayYYMMDD)
	check(err)
	yearAgoYYMMDDint := todayYYMMDDint - 10000
	_, err = tx.Exec("INSERT INTO dropout (name, count, last_update, drop_date) SELECT name, count, last_update, CAST(strftime('%Y%m%d', 'now') AS int) FROM distrs WHERE last_update < ?", yearAgoYYMMDDint)
	check(err)
	_, err = tx.Exec("DELETE FROM distrs WHERE last_update < ?", yearAgoYYMMDDint)
	check(err)

	var latitude, longitude float64
	err = db.QueryRow("SELECT latitude, longitude FROM coords ORDER BY date DESC LIMIT 1").Scan(&latitude, &longitude)
	if err != nil {
		if err == sql.ErrNoRows {
			latitude = initialLatitude
			longitude = initialLongitude
		} else {
			log.Fatal(err)
		}
	}

	longitudeDiff := float64(outcome.next1HPD) / float64(divider)
	longitudeTrend := outcome.next1Trend
	latitudeDiff := float64(outcome.next2HPD) / float64(divider)
	latitudeTrend := outcome.next2Trend
	longitude += longitudeDiff * float64(longitudeTrend)
	latitude += latitudeDiff * float64(latitudeTrend)

	if latitude > 90 {
		latitude = latitude - 180
	}
	if longitude > 180.0 {
		longitude = longitude - 360.0
	}
	_, err = tx.Exec("INSERT INTO coords (date, longitude_diff, longitude_trend, latitude_diff, latitude_trend, latitude, longitude) VALUES (?, ?, ?, ?, ?, ?, ?)", todayYYMMDD, fmt.Sprintf("%.4f", longitudeDiff), longitudeTrend, fmt.Sprintf("%.4f", latitudeDiff), latitudeTrend, fmt.Sprintf("%.4f", latitude), fmt.Sprintf("%.4f", longitude))
	check(err)

	err = tx.Commit()
	check(err)
}

func downloadScreenshot(distrURL string) string {
	// Get distr page
	response := getResponse(distrURL)
	defer closeCheck(response.Body)
	checkResponse(response)

	// Parse the page and fetch full url of screenshot.
	root, err := goquery.NewDocumentFromReader(response.Body)
	check(err)

	a := root.Find("td.TablesTitle > a").First()
	if a.Length() == 0 {
		log.Fatalf("Could not find screenshot on page %s", distrURL)
	}
	url, ok := a.Attr("href")
	if !ok {
		log.Fatalf("Screenshot a has not attribute 'href' on page %s", distrURL)
	}
	if !strings.HasPrefix(url, "http") {
		url = baseURL + url
	}
	base := path.Base(url)
	screenshotPath := path.Join(distrowatch.DistrsDir, base)

	// Download screenshot
	output, err := os.Create(screenshotPath)
	if err != nil {
		log.Fatalf("Could not create file %s, err: %s", screenshotPath, err)
	}
	defer closeCheck(output)

	response = getResponse(url)
	defer closeCheck(response.Body)
	checkResponse(response)

	_, err = io.Copy(output, response.Body)
	if err != nil {
		log.Fatalf("Could not write image %s to file %s, err: %s", url, screenshotPath, err)
	}

	return screenshotPath
}

func main() {
	// Create directory if it doesn't exist.
	_, err := os.Stat(distrowatch.DistrsDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(distrowatch.DistrsDir, 0755)
		check(err)
	}

	// Open database and create tables if they don't exist.
	db, err := distrowatch.GetDB()
	check(err)
	defer closeCheck(db)
	var answer string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='distrs'").Scan(&answer)
	if err != nil {
		if err == sql.ErrNoRows {
			_, err = db.Exec("CREATE TABLE 'distrs' (`name` TEXT NOT NULL UNIQUE, `count` INTEGER NOT NULL, `last_update` INTEGER NOT NULL UNIQUE, PRIMARY KEY(`name`))")
			check(err)
		} else {
			log.Fatal(err)
		}
	}
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='coords'").Scan(&answer)
	if err != nil {
		if err == sql.ErrNoRows {
			_, err = db.Exec("CREATE TABLE 'coords' (`date` INTEGER NOT NULL UNIQUE, `longitude_diff` FLOAT, `longitude_trend` INTEGER, `latitude_diff` FLOAT, `latitude_trend` INTEGER, `latitude` FLOAT NOT NULL, `longitude` FLOAT NOT NULL, PRIMARY KEY(`date`))")
			check(err)
		} else {
			log.Fatal(err)
		}
	}
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='dropout'").Scan(&answer)
	if err != nil {
		if err == sql.ErrNoRows {
			_, err = db.Exec("CREATE TABLE 'dropout' (`name` TEXT NOT NULL, `count` INTEGER NOT NULL, `last_update` INTEGER NOT NULL UNIQUE, `drop_date` INTEGER NOT NULL, PRIMARY KEY(`last_update`))")
			check(err)
		} else {
			log.Fatal(err)
		}
	}
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='distrs_daily'").Scan(&answer)
	if err != nil {
		if err == sql.ErrNoRows {
			_, err = db.Exec("CREATE TABLE 'distrs_daily' (`date` INTEGER NOT NULL UNIQUE, `name` TEXT NOT NULL, `hpd` INTEGER NOT NULL, PRIMARY KEY(`date`))")
			check(err)
		} else {
			log.Fatal(err)
		}
	}
	_ = answer

	// If date of last_update is today, exit.
	var lastUpdate string
	err = db.QueryRow("SELECT last_update FROM distrs ORDER BY last_update DESC LIMIT 1").Scan(&lastUpdate)
	if err != nil && err != sql.ErrNoRows {
		log.Fatal(err)
	}
	if lastUpdate == todayYYMMDD {
		fmt.Println("Database is already updated today.")
		os.Exit(0)
	}

	outcome := getOutcome()
	updateDb(db, outcome)
	_ = downloadScreenshot(outcome.distrURL)
}
