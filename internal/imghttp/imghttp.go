package imghttp

import (
	"fmt"
	"image"
	"net/http"
)

func GetImage(imgPath string) (image.Image, string, error) {
	resp, err := http.Get(imgPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch image: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("error fetching image: HTTP %d", resp.StatusCode)
	}

	// Decode the image
	img, format, err := image.Decode(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode image: %v", err)
	}
	return img, format, err
}

func ContentType(extension string, imgData []byte) string {
	contentType := http.DetectContentType(imgData)
	if extension == "avif" {
		contentType = "image/avif"
	}

	return contentType
}

func ImageHeaders(imgFormat string, imgData []byte) map[string]string {
	return map[string]string{
		"Content-Type":  ContentType(imgFormat, imgData),
		"Cache-Control": "public, max-age=604800", // Cache for 7 days
	}
}
