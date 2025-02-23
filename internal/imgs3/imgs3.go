package imgs3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/StrongerSoftworks/image-proxy/internal/imghttp"
	"github.com/StrongerSoftworks/image-proxy/internal/transformations"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// retrieves the bucket name from environment variables
func GetBucketName() string {
	bucket := os.Getenv("S3_BUCKET")
	if bucket == "" {
		log.Fatal("S3_BUCKET environment variable is not set")
	}
	return bucket
}

func MakeBucketFileKey(imgPath string, options *transformations.Options) string {
	transformedFileName := fmt.Sprintf("%s.%s", strings.TrimSuffix(filepath.Base(imgPath), filepath.Ext(imgPath)), options.Format)
	return fmt.Sprintf("%s/%s/%d/%d/%f/%d/%s", url.PathEscape(imgPath), options.Mode, options.Width, options.Height, options.AspectRatio, options.Quality, transformedFileName)
}

// checks if a file exists in the S3 bucket
func ImageExists(ctx context.Context, client *s3.Client, bucket, key string) (bool, error) {
	_, err := client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		// Check if error is a "NotFound" or "NoSuchKey"
		var apiErr *types.NoSuchKey
		if errors.As(err, &apiErr) || strings.Contains(err.Error(), "NotFound") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// retrieves a file from the S3 bucket
func GetImage(ctx context.Context, client *s3.Client, bucket, key string) ([]byte, error) {
	output, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer output.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(output.Body)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// uploads a file to the S3 bucket
func UploadImage(ctx context.Context, uploader *manager.Uploader, bucket, key string, imgData []byte) error {
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(imgData),
		ContentType: aws.String(imghttp.ContentType(filepath.Ext(key), imgData)),
	})
	return err
}

func InitAWS(ctx context.Context) *s3.Client {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		log.Fatalf("unable to load AWS SDK config, %v", err)
	}
	s3Client := s3.NewFromConfig(cfg)
	return s3Client
}
