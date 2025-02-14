package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/StrongerSoftworks/image-proxy/internal/imghttp"
	"github.com/StrongerSoftworks/image-proxy/internal/imgs3"
	"github.com/StrongerSoftworks/image-proxy/internal/transformations"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// Set up AWS connections
	bucket := getBucketName()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)

	// Extract query parameters
	imgPath := request.QueryStringParameters["img"]
	widthQuery := request.QueryStringParameters["width"]
	heightQuery := request.QueryStringParameters["height"]
	aspectRatioQuery := request.QueryStringParameters["ratio"]
	modeQuery := request.QueryStringParameters["mode"]
	formatQuery := request.QueryStringParameters["format"]
	qualityQuery := request.QueryStringParameters["quality"]

	format, err := transformations.FormatFromPath(imgPath)
	if err != nil {
		log.Printf("Error getting format from file URL: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
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
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
	}

	// Check if file exists
	s3FileKey := imgs3.MakeBucketFileKey(imgPath, &options)

	if exists, existsErr := imgs3.ImageExists(ctx, client, bucket, s3FileKey); existsErr != nil {
		log.Printf("Error getting format from file URL: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
	} else if exists {
		// Retrieve image from bucket
		imgData, err := imgs3.GetImage(ctx, client, bucket, s3FileKey)
		if err != nil {
			log.Printf("Error retrieving file: %v", err)
			return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
		}
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Headers:    imghttp.ImageHeaders(options.Format, imgData),
			Body:       string(imgData),
			// IsBase64Encoded: true,
		}, nil
	}

	// Get the image from source
	img, _, err := imghttp.GetImage(imgPath)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
	}

	// Apply transformations
	imgData, err := transformations.TransformImage(img, &options)
	if err != nil {
		log.Printf("Could not apply transformations to image: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
	}

	// Upload transformed image to S3
	err = imgs3.UploadImage(ctx, uploader, bucket, s3FileKey, imgData.Bytes())
	if err != nil {
		log.Printf("Error uploading file: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
	}

	// Return the transformed image
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    imghttp.ImageHeaders(options.Format, imgData.Bytes()),
		Body:       imgData.String(),
		// IsBase64Encoded: true,
	}, nil

}

// retrieves the bucket name from environment variables
func getBucketName() string {
	bucket := os.Getenv("S3_BUCKET_NAME")
	if bucket == "" {
		log.Fatal("S3_BUCKET_NAME environment variable is not set")
	}
	return bucket
}
