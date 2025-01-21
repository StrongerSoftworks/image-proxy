package lambda

import (
	"image"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/StrongerSoftworks/image-resizer/transformations"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Parse the S3 bucket and key from the URL path
	bucket := os.Getenv("S3_BUCKET")
	key := strings.TrimPrefix(request.Path, "/")

	// Extract query parameters
	width, _ := strconv.Atoi(request.QueryStringParameters["width"])
	height, _ := strconv.Atoi(request.QueryStringParameters["height"])
	aspectRatioQuery := request.QueryStringParameters["aspect-ratio"]
	format := request.QueryStringParameters["format"]
	quality, _ := strconv.Atoi(request.QueryStringParameters["quality"])

	// Initialize AWS S3 client
	sess := session.Must(session.NewSession())
	s3Client := s3.New(sess)

	// Get the image from S3
	s3Object, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Printf("Failed to get object from S3: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusNotFound}, nil
	}
	defer s3Object.Body.Close()

	// Decode the image
	img, _, err := image.Decode(s3Object.Body)
	if err != nil {
		log.Printf("Failed to decode image: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest}, nil
	}

	var aspectRatio = 0.0
	if aspectRatioQuery != "" {
		ratio, found := transformations.AspectRatio(aspectRatioQuery)
		if found {
			aspectRatio = ratio
		}
	}

	if format != "" {
		if !transformations.ValidateFormat(format) {
			log.Printf("Invalid or missing extension: %v", format)
			return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest}, nil
		}
	}

	// Apply transformations
	imgData, err := transformations.TransformImage(img, width, height, aspectRatio, quality, format)
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
		Body:            string(imgData.Bytes()),
		IsBase64Encoded: true,
	}, nil

}

func main() {
	lambda.Start(handler)
}
