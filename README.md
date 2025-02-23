# image-proxy

# Build Container

## For Release

```
docker buildx build -t image-proxy -f Dockerfile .
```

## For Debug

```
docker buildx build -t image-proxy-debug -f DockerfileDebug .
```

# Run Container

## Development

```
docker run \
  -e GO_ENV=development \
  -p 8080:8080 \
  --name image-proxy-development \
  -d image-proxy
```

## Production

```
docker run \
  -e GO_ENV=production \
  -p 8080:8080 \
  --name image-proxy-production \
  -d image-proxy
```

## Development Debug

```
docker run \
  -e GO_ENV=development \
  -p 40000:40000 \
  -p 8080:8080 \
  --name image-proxy-debug-development \
  -d image-proxy-debug
```

## Production Debug

```
docker run \
  -e GO_ENV=production \
  -p 40000:40000 \
  -p 8080:8080 \
  --name image-proxy-debug-production \
  -d image-proxy-debug
```

# Tag and Push Container

## Production

```
aws lightsail push-container-image --profile calculator --region ca-central-1 --service-name image-proxy --label image-proxy --image image-proxy:latest
```
