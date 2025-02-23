package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/StrongerSoftworks/image-proxy/internal/imghttp"
	"github.com/StrongerSoftworks/image-proxy/internal/imgs3"
	"github.com/StrongerSoftworks/image-proxy/internal/transformations"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3RequestHanlder struct {
	ImageProxyRequestHandler
	s3Client   *s3.Client
	uploader   *manager.Uploader
	bucketName string
}

func NewS3RequestHanlder() *S3RequestHanlder {
	handler := S3RequestHanlder{}
	return &handler
}

func (handler *S3RequestHanlder) Init() {
	bucketName := imgs3.GetBucketName()
	log.Println("Images will be saved to " + bucketName)
	handler.s3Client = imgs3.InitAWS(context.Background())
	handler.uploader = manager.NewUploader(handler.s3Client)
	handler.bucketName = bucketName
}

func (handler *S3RequestHanlder) Handler(w http.ResponseWriter, r *http.Request) {
	imgPath := r.URL.Query().Get("img")
	widthQuery := r.URL.Query().Get("width")
	heightQuery := r.URL.Query().Get("height")
	aspectRatioQuery := r.URL.Query().Get("ratio")
	modeQuery := r.URL.Query().Get("mode")
	formatQuery := r.URL.Query().Get("format")
	qualityQuery := r.URL.Query().Get("quality")

	format, err := transformations.FormatFromPath(imgPath)
	if err != nil {
		http.Error(w, "Invalid image format", http.StatusInternalServerError)
		return
	}

	options := transformations.Options{
		Quality: 100,
		Mode:    "fit",
		Format:  format,
	}
	err = transformations.ParseOptions(widthQuery, heightQuery, formatQuery, modeQuery, qualityQuery, aspectRatioQuery, &options)
	if err != nil {
		http.Error(w, "Invalid transformation options", http.StatusBadRequest)
		return
	}

	s3Key := imgs3.MakeBucketFileKey(imgPath, &options)

	// Check if image exists in S3
	_, err = handler.s3Client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: &handler.bucketName,
		Key:    &s3Key,
	})
	if err == nil {
		redirectURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", handler.bucketName, s3Key)
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	img, _, err := imghttp.GetImage(imgPath)
	if err != nil {
		http.Error(w, "Error downloading image", http.StatusInternalServerError)
		return
	}

	imgData, err := transformations.TransformImage(img, &options)
	if err != nil {
		http.Error(w, "Error transforming image", http.StatusInternalServerError)
		return
	}

	err = imgs3.UploadImage(context.TODO(), handler.uploader, handler.bucketName, s3Key, imgData.Bytes())
	if err != nil {
		http.Error(w, "Error uploading to S3", http.StatusInternalServerError)
		return
	}

	redirectURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", handler.bucketName, s3Key)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}
