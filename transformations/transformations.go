package transformations

import (
	"bytes"
	"fmt"
	"image"
	"strings"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
)

// Aspect ratio mappings
var aspectRatios = map[string]float64{
	"16x9": 16.0 / 9.0,
	"9x16": 9.0 / 16.0,
	"1x1":  1.0,
	"4x3":  4.0 / 3.0,
	"3x4":  3.0 / 4.0,
}

func ValidateFormat(extension string) bool {
	validExtensions := map[string]bool{
		"jpg":  true,
		"jpeg": true,
		"png":  true,
		"webp": true,
		"avif": true,
	}
	return validExtensions[extension]
}

func AspectRatio(aspectRatio string) (float64, bool) {
	ratio, exists := aspectRatios[aspectRatio]
	return ratio, exists
}

func TransformImage(img image.Image, width int, height int, aspectRatio float64, quality int, format string) (*bytes.Buffer, error) {
	// Apply transformations
	if aspectRatio != 0 {
		if width == 0 && height == 0 {
			width = img.Bounds().Dx()
			height = int(float64(width) / aspectRatio)
		} else if width == 0 {
			width = int(float64(height) * aspectRatio)
		} else if height == 0 {
			height = int(float64(width) / aspectRatio)
		}
	}

	if width > 0 || height > 0 {
		if height == 0 {
			height = img.Bounds().Dy()
		}

		if width == 0 {
			width = img.Bounds().Dx()
		}

		img = imaging.Fit(img, width, height, imaging.Lanczos)
		// TODO implement imaging.Crop as an option
	}

	// Set output quality
	qualityPercent := 100
	if quality > 0 {
		qualityPercent = quality
	}

	// Convert and encode the image
	var buf bytes.Buffer
	var err error
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		err = imaging.Encode(&buf, img, imaging.JPEG, imaging.JPEGQuality(qualityPercent))
	case "png":
		err = imaging.Encode(&buf, img, imaging.PNG)
	case "webp":
		err = webp.Encode(&buf, img, &webp.Options{Lossless: true})
	case "avif":
		// AVIF is not natively supported; additional library needed
		err = fmt.Errorf("AVIF format not supported in this example")
	default:
		err = imaging.Encode(&buf, img, imaging.JPEG, imaging.JPEGQuality(qualityPercent))
	}

	return &buf, err
}
