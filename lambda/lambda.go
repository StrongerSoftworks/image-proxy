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
	width, _ := strconv.Atoi(request.QueryStringParameters["width"])
	height, _ := strconv.Atoi(request.QueryStringParameters["height"])
	aspectRatioQuery := request.QueryStringParameters["aspect-ratio"]
	mode := request.QueryStringParameters["mode"]
	formatQuery := request.QueryStringParameters["format"]
	quality, _ := strconv.Atoi(request.QueryStringParameters["quality"])

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

	// Validate options
	var aspectRatio float32 = 0.0
	if aspectRatioQuery != "" {
		ratio, found := transformations.AspectRatio(aspectRatioQuery)
		if found {
			aspectRatio = ratio
		}
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
