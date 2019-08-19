package show

import (
	"testing"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func TestGetCoords(t *testing.T) {
	var coords, err = GetCoords("./testdata/db.sqlite3")
	check(err)
	var expectedCoords = Coords{60.9888, 30.5187, 0.0828, -0.0839}
	if coords != expectedCoords {
		t.Errorf("expected %v, got %v", expectedCoords, coords)
	}
}
