package lambda

import (
	"image"
	"log"
	"net/http"
	"strconv"

	"github.com/StrongerSoftworks/image-resizer/transformations"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// Extract query parameters
	imgPath := request.QueryStringParameters["img"]
	widthQuery := request.QueryStringParameters["width"]
	heightQuery := request.QueryStringParameters["height"]
	aspectRatioQuery := request.QueryStringParameters["ratio"]
	modeQuery := request.QueryStringParameters["mode"]
	formatQuery := request.QueryStringParameters["format"]
	qualityQuery := request.QueryStringParameters["quality"]

	width := 0
	height := 0
	quality := 100
	mode := "fit"

	// Validate options
	var aspectRatio float32 = 0.0
	if aspectRatioQuery != "" {
		ratio, found := transformations.AspectRatioToFloat(aspectRatioQuery)
		if found {
			aspectRatio = ratio
		}
	}

	if widthQuery != "" {
		var err error
		width, err = strconv.Atoi(widthQuery)
		if err != nil {
			log.Printf("Invalid width: %v", width)
			return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest}, nil
		}
	}

	if heightQuery != "" {
		var err error
		height, err = strconv.Atoi(heightQuery)
		if err != nil {
			log.Printf("Invalid height: %v", height)
			return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest}, nil
		}
	}

	if modeQuery != "" {
		mode = modeQuery
	}

	if qualityQuery != "" {
		var err error
		quality, err = strconv.Atoi(qualityQuery)
		if err != nil || quality < 0 || quality > 100 {
			log.Printf("Invalid quality: %v", quality)
			return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest}, nil
		}
	}

	// get the image
	resp, err := http.Get(imgPath)
	if err != nil {
		log.Printf("Failed to fetch image: %v\n", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error fetching image: HTTP %d\n", resp.StatusCode)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, nil
	}

	// Decode the image
	img, format, err := image.Decode(resp.Body)
	if err != nil {
		log.Printf("Failed to decode image: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest}, nil
	}

	if formatQuery != "" {
		format = formatQuery
	}

	// Apply transformations
	imgData, err := transformations.TransformImage(img, &transformations.Options{Width: width, Height: height, AspectRatio: aspectRatio, Mode: mode, Quality: quality, Format: format})
	if err != nil {
		log.Printf("Could not apply transformations to image: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, nil
	}

	// Return the transformed image
	headers := map[string]string{
		"Content-Type":  http.DetectContentType(imgData.Bytes()),
		"Cache-Control": "public, max-age=604800", // Cache for 7 days
	}

	return events.APIGatewayProxyResponse{
		StatusCode:      http.StatusOK,
		Headers:         headers,
		Body:            imgData.String(),
		IsBase64Encoded: true,
	}, nil

}

func main() {
	lambda.Start(handler)
}
