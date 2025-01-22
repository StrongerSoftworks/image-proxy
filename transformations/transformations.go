package transformations

import (
	"bytes"
	"fmt"
	"image"
	"strings"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/gen2brain/avif"
)

const (
	Crop = "crop"
	Fit  = "fit"
)

// Aspect ratio mappings
var aspectRatios = map[string]float32{
	"16x9": 16.0 / 9.0,
	"9x16": 9.0 / 16.0,
	"1x1":  1.0,
	"4x3":  4.0 / 3.0,
	"3x4":  3.0 / 4.0,
}

func validateFormat(extension string) bool {
	validExtensions := map[string]bool{
		"jpg":  true,
		"jpeg": true,
		"png":  true,
		"webp": true,
		"avif": true,
	}
	return validExtensions[extension]
}

func validateMode(mode string) bool {
	validModes := map[string]bool{
		"fit":  true,
		"crop": true,
	}
	return validModes[mode]
}

func AspectRatio(aspectRatio string) (float32, bool) {
	ratio, exists := aspectRatios[aspectRatio]
	return ratio, exists
}

type Options struct {
	Width       int
	Height      int
	AspectRatio float32
	Mode        string
	Quality     int
	Format      string
}

func TransformImage(img image.Image, options *Options) (*bytes.Buffer, error) {
	// Validate options
	if options.Format != "" {
		if !validateFormat(options.Format) {
			return nil, fmt.Errorf("invalid extension: %s", options.Format)
		}
	}

	if options.Mode != "" {
		if !validateMode(options.Mode) {
			return nil, fmt.Errorf("invalid mode: %s", options.Mode)
		}
	}

	// Apply transformations
	if options.AspectRatio != 0 {
		if options.Width == 0 && options.Height == 0 {
			options.Width = img.Bounds().Dx()
			options.Height = int(float32(options.Width) / options.AspectRatio)
		} else if options.Width == 0 {
			options.Width = int(float32(options.Height) * options.AspectRatio)
		} else if options.Height == 0 {
			options.Height = int(float32(options.Width) / options.AspectRatio)
		}
	}

	if options.Width > 0 || options.Height > 0 {
		if options.Height == 0 {
			options.Height = img.Bounds().Dy()
		}

		if options.Width == 0 {
			options.Width = img.Bounds().Dx()
		}

		if options.Mode == Crop {
			img = imaging.CropCenter(img, options.Width, options.Height)
		} else {
			img = imaging.Fit(img, options.Width, options.Height, imaging.Lanczos)
		}
	}

	// Set output quality
	qualityPercent := 100
	if options.Quality > 0 {
		qualityPercent = options.Quality
	}

	// Convert and encode the image
	var buf bytes.Buffer
	var err error
	switch strings.ToLower(options.Format) {
	case "jpeg", "jpg":
		err = imaging.Encode(&buf, img, imaging.JPEG, imaging.JPEGQuality(qualityPercent))
	case "png":
		err = imaging.Encode(&buf, img, imaging.PNG) // TODO quality on PNG?
	case "webp":
		err = webp.Encode(&buf, img, &webp.Options{Lossless: true, Quality: float32(qualityPercent), Exact: true})
	case "avif":
		err = avif.Encode(&buf, img, avif.Options{Quality: qualityPercent})
	default:
		err = imaging.Encode(&buf, img, imaging.JPEG, imaging.JPEGQuality(qualityPercent))
	}

	return &buf, err
}
