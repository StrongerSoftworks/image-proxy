# image-proxy

# Build Container For Release

```
docker buildx build -t image-proxy -f Dockerfile .
```

# Run Container

## Development

```
docker run \
  -e GO_ENV=development \
  -p 8080:8080 \
  --name image-proxy.development \
  -d image-proxy
```

## Production

```
docker run \
  -e GO_ENV=production \
  -p 8080:8080 \
  --name image-proxy.production \
  -d image-proxy
```

# Tag and Push Container

```
docker tag image-proxy:latest image-proxy:0.0.0
aws lightsail push-container-image --profile calculator --region ca-central-1 --service-name image-proxy --label image-proxy --image image-proxy:latest
aws lightsail push-container-image --profile calculator --region ca-central-1 --service-name image-proxy --label image-proxy --image image-proxy:0.0.0
```
