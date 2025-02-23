package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/StrongerSoftworks/image-proxy/internal/imghttp"
	"github.com/StrongerSoftworks/image-proxy/internal/imgpath"
	"github.com/StrongerSoftworks/image-proxy/internal/transformations"
)

type LocalRequestHandler struct{}

func NewLocalRequestHandler() *LocalRequestHandler {
	handler := LocalRequestHandler()
	return &handler
}

func (*LocalRequestHandler) Init() {
	log.Println("Images will be saved to " + imageBasePath())
}

func (*LocalRequestHandler) Handler(w http.ResponseWriter, r *http.Request) {

	// Extract query parameters
	imgPath := r.URL.Query().Get("img")
	widthQuery := r.URL.Query().Get("width")
	heightQuery := r.URL.Query().Get("height")
	aspectRatioQuery := r.URL.Query().Get("ratio")
	modeQuery := r.URL.Query().Get("mode")
	formatQuery := r.URL.Query().Get("format")
	qualityQuery := r.URL.Query().Get("quality")

	format, err := transformations.FormatFromPath(imgPath)
	if err != nil {
		log.Printf("Error getting format from file URL: %v", err)
		http.Error(w, fmt.Sprintf("Error getting format from file URL: %v", err), http.StatusInternalServerError)
		return
	}

	// Parse and validate options
	var options = transformations.Options{
		Quality: 100,
		Mode:    "fit",
		Format:  format,
	}
	err = transformations.ParseOptions(widthQuery, heightQuery, formatQuery, modeQuery, qualityQuery, aspectRatioQuery, &options)
	if err != nil {
		log.Printf("Issue parsing options: %v", err)
		http.Error(w, fmt.Sprintf("Issue parsing options: %v", err), http.StatusBadRequest)
		return
	}

	// Check if file exists
	filePath := filepath.Join(imageBasePath(), imgpath.MakeFilePath(imgPath, &options))
	if _, err := os.Stat(filePath); err == nil {
		imgData, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Error getting image: %v", err)
			http.Error(w, fmt.Sprintf("Error getting image: %v", err), http.StatusInternalServerError)
			return
		}
		writeResponse(w, options, imgData)
		return
	}

	// Get the image from source
	img, _, err := imghttp.GetImage(imgPath)
	if err != nil {
		log.Printf("Error downloading image: %v", err)
		http.Error(w, fmt.Sprintf("Issue getting image: %v", err), http.StatusInternalServerError)
		return
	}

	// Apply transformations
	imgData, err := transformations.TransformImage(img, &options)
	if err != nil {
		log.Printf("Could not apply transformations to image: %v", err)
		http.Error(w, fmt.Sprintf("Could not apply transformations to image: %v", err), http.StatusInternalServerError)
		return
	}

	// Save transformed image
	err = writeFileWithDirs(filePath, imgData.Bytes(), 0644)
	if err != nil {
		log.Printf("Error saving image: %v", err)
		http.Error(w, fmt.Sprintf("Error saving image: %v", err), http.StatusInternalServerError)
		return
	}

	// Return the transformed image
	writeResponse(w, options, imgData.Bytes())
}

func writeResponse(w http.ResponseWriter, options transformations.Options, imgData []byte) {
	headers := imghttp.ImageHeaders(options.Format, imgData)
	for key, value := range headers {
		w.Header().Set(key, value)
	}

	w.WriteHeader(http.StatusOK) // Optional, as 200 is the default status code
	if _, err := w.Write(imgData); err != nil {
		log.Printf("Failed to write image data to response: %v\n", err)
	}
}

func writeFileWithDirs(imgPath string, data []byte, perm os.FileMode) error {
	// Create all parent directories with 0755 permissions
	dir := filepath.Dir(imgPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Write the file
	if err := os.WriteFile(imgPath, data, perm); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func imageBasePath() string {
	return path.Join(os.TempDir(), "image-proxy", "images")
}
