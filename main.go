package main

import (
	"fmt"
	"image"
	"log"
	"net/http"
	"strconv"

	"github.com/StrongerSoftworks/image-resizer/transformations"
)

func handler(w http.ResponseWriter, r *http.Request) {

	imgPath := r.URL.Query().Get("img")
	widthStr := r.URL.Query().Get("width")
	heightStr := r.URL.Query().Get("height")
	aspectRatioQuery := r.URL.Query().Get("aspect-ratio")
	modeQuery := r.URL.Query().Get("mode")
	formatQuery := r.URL.Query().Get("format")
	qualityStr := r.URL.Query().Get("quality")

	width := 0
	height := 0
	quality := 100
	mode := "fit"

	// get the image
	resp, err := http.Get(imgPath)
	if err != nil {
		fmt.Printf("Failed to fetch image: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error fetching image: HTTP %d\n", resp.StatusCode)
		return
	}

	// Decode the image
	img, format, err := image.Decode(resp.Body)
	if err != nil {
		fmt.Printf("Failed to decode image: %v\n", err)
		return
	}

	if widthStr != "" {
		var err error
		width, err = strconv.Atoi(widthStr)
		if err != nil {
			http.Error(w, "Invalid width", http.StatusBadRequest)
			return
		}
	}

	if heightStr != "" {
		var err error
		height, err = strconv.Atoi(heightStr)
		if err != nil {
			http.Error(w, "Invalid height", http.StatusBadRequest)
			return
		}
	}

	if formatQuery != "" {
		format = formatQuery
	}

	if modeQuery != "" {
		mode = modeQuery
	}

	if qualityStr != "" {
		var err error
		quality, err = strconv.Atoi(qualityStr)
		if err != nil || quality < 0 || quality > 100 {
			http.Error(w, "Invalid quality", http.StatusBadRequest)
			return
		}
	}

	var aspectRatio float32 = 0.0
	if aspectRatioQuery != "" {
		ratio, found := transformations.AspectRatio(aspectRatioQuery)
		if found {
			aspectRatio = ratio
		}
	}

	// Apply transformations
	imgData, err := transformations.TransformImage(img, &transformations.Options{Width: width, Height: height, AspectRatio: aspectRatio, Mode: mode, Quality: quality, Format: format})
	if err != nil {
		log.Printf("Could not apply transformations to image: %v", err)
		http.Error(w, fmt.Sprintf("Could not apply transformations to image: %v", err), http.StatusBadRequest)
		return
	}

	contentType := http.DetectContentType(imgData.Bytes())
	if format == "avif" {
		contentType = "image/avif"
	}

	headers := map[string]string{
		"Content-Type": contentType,
		// "Cache-Control": "public, max-age=604800", // Cache for 7 days
	}
	for key, value := range headers {
		w.Header().Set(key, value)
	}

	w.WriteHeader(http.StatusOK) // Optional, as 200 is the default status code
	if _, err := w.Write(imgData.Bytes()); err != nil {
		fmt.Printf("Failed to write image data to response: %v\n", err)
	}
}

func main() {
	http.HandleFunc("/process", handler)
	fmt.Println("Server is running on port 8080...")
	http.ListenAndServe(":8080", nil)
}
