package imgpath

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/StrongerSoftworks/image-proxy/internal/transformations"
)

func MakeFilePath(imgPath string, options *transformations.Options) string {
	transformedFileName := fmt.Sprintf("%s.%s", strings.TrimSuffix(filepath.Base(imgPath), filepath.Ext(imgPath)), options.Format)
	return filepath.Join(sanitizePath(url.PathEscape(imgPath)), options.Mode,
		strconv.Itoa(options.Width), strconv.Itoa(options.Height),
		strconv.FormatFloat(float64(options.AspectRatio), 'f', -1, 32), strconv.Itoa(options.Quality), transformedFileName)
}

func sanitizePath(path string) string {
	// Replace potentially problematic characters with underscores
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
	)
	return replacer.Replace(path)
}
