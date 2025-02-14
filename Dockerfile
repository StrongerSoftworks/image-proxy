FROM amazonlinux:2023 AS builder

# Install dependencies
RUN yum update -y && yum install -y \
    gcc \
    gcc-c++ \
    make \
    tar \
    gzip \
    git \
    wget \
    && yum clean all
RUN wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
RUN tar -xvf go1.24.0.linux-amd64.tar.gz
RUN mv go /usr/local

ENV GOROOT=/usr/local/go
ENV PATH=$PATH:$GOROOT/bin

# Set the working directory
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum  /app/
RUN go mod download

# Copy the rest of the application source code
COPY ./cmd/api /app/cmd/api
COPY ./internal /app/internal
COPY ./cmd/api/.env.development ./cmd/api/.env.production /app/

# Build the Go application
RUN env GOOS=linux GOARCH=amd64 go build -o main cmd/api/main.go

# Stage 2: Create a minimal runtime image
FROM amazonlinux:2023

# Set the working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/.env.development .
COPY --from=builder /app/.env.production .

# Expose port 8080 for incoming requests
EXPOSE 8080

# Command to run the application
CMD ["./main"]
