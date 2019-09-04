package show

import (
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/andbar-ru/average_color"
	"github.com/andbar-ru/distrowatch"
)

// GetLastDistrImageAverageColor returns average color of the last distr image from distrsDir.
func GetLastDistrImageAverageColor() (color.NRGBA, error) {
	files, err := ioutil.ReadDir(distrowatch.DistrsDir)
	if err != nil {
		return color.NRGBA{}, err
	}
	// find last image
	var lastImage string
	var lastModTime time.Time
	for _, file := range files {
		ext := strings.ToLower(path.Ext(file.Name()))
		// interested only in images
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".gif" {
			continue
		}
		modTime := file.ModTime()
		if modTime.After(lastModTime) {
			lastModTime = modTime
			lastImage = path.Join(distrowatch.DistrsDir, file.Name())
		}
	}
	if lastImage == "" {
		return color.NRGBA{}, fmt.Errorf("could not find images in directory %s", distrowatch.DistrsDir)
	}
	f, err := os.Open(lastImage)
	if err != nil {
		return color.NRGBA{}, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return color.NRGBA{}, err
	}
	averageColor := average_color.AverageColor(img)

	return averageColor, nil
}
