package transformations

import (
	"bytes"
	"image"
	"reflect"
	"testing"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/gen2brain/avif"
)

func TestTransformImage(t *testing.T) {
	type args struct {
		img     image.Image
		options *Options
	}
	tests := []struct {
		name    string
		args    args
		want    *bytes.Buffer
		wantErr bool
	}{
		{
			name: "Valid transformation with fit mode",
			args: args{
				img: image.NewRGBA(image.Rect(0, 0, 200, 200)),
				options: &Options{
					Width:   100,
					Height:  100,
					Mode:    Fit,
					Format:  "jpeg",
					Quality: 80,
				},
			},
			want: func() *bytes.Buffer {
				var buf bytes.Buffer
				imaging.Encode(&buf, imaging.Resize(image.NewRGBA(image.Rect(0, 0, 200, 200)), 100, 100, imaging.Lanczos), imaging.JPEG, imaging.JPEGQuality(80))
				return &buf
			}(),
			wantErr: false,
		},
		{
			name: "Valid transformation with crop mode",
			args: args{
				img: image.NewRGBA(image.Rect(0, 0, 200, 200)),
				options: &Options{
					Width:   100,
					Height:  100,
					Mode:    Crop,
					Format:  "png",
					Quality: 90,
				},
			},
			want: func() *bytes.Buffer {
				var buf bytes.Buffer
				croppedImg := imaging.CropCenter(image.NewRGBA(image.Rect(0, 0, 200, 200)), 100, 100)
				imaging.Encode(&buf, croppedImg, imaging.PNG)
				return &buf
			}(),
			wantErr: false,
		},
		{
			name: "Transformation with aspect ratio and width only",
			args: args{
				img: image.NewRGBA(image.Rect(0, 0, 160, 90)),
				options: &Options{
					Width:       80,
					AspectRatio: 16.0 / 9.0,
					Mode:        Fit,
					Format:      "webp",
					Quality:     75,
				},
			},
			want: func() *bytes.Buffer {
				var buf bytes.Buffer
				resizedImg := imaging.Fit(image.NewRGBA(image.Rect(0, 0, 160, 90)), 80, int(80/(16.0/9.0)), imaging.Lanczos)
				webp.Encode(&buf, resizedImg, &webp.Options{Lossless: true, Quality: 75, Exact: true})
				return &buf
			}(),
			wantErr: false,
		},
		{
			name: "Invalid format",
			args: args{
				img: image.NewRGBA(image.Rect(0, 0, 100, 100)),
				options: &Options{
					Format: "bmp",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid mode",
			args: args{
				img: image.NewRGBA(image.Rect(0, 0, 100, 100)),
				options: &Options{
					Mode: "stretch",
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Transformation with aspect ratio and height only",
			args: args{
				img: image.NewRGBA(image.Rect(0, 0, 160, 90)),
				options: &Options{
					Height:      45,
					AspectRatio: 16.0 / 9.0,
					Mode:        Fit,
					Format:      "avif",
					Quality:     85,
				},
			},
			want: func() *bytes.Buffer {
				var buf bytes.Buffer
				resizedImg := imaging.Fit(image.NewRGBA(image.Rect(0, 0, 160, 90)), int(45*(16.0/9.0)), 45, imaging.Lanczos)
				avif.Encode(&buf, resizedImg, avif.Options{Quality: 85})
				return &buf
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TransformImage(tt.args.img, tt.args.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransformImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TransformImage() = %v, want %v", got, tt.want)
			}
		})
	}
}
