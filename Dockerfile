# Use an argument to toggle platforms (default to arm64)
ARG BUILD_PLATFORM=linux/arm64
# PIN TO BOOKWORM for Raspberry Pi OS compatibility
FROM --platform=${BUILD_PLATFORM} golang:1.25.1-bookworm

# Install Ebitengine Linux dependencies
RUN apt-get update && apt-get install -y \
    libgl1-mesa-dev \
    xorg-dev \
    libasound2-dev \
    pkg-config && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Pre-copy go.mod and go.sum for better caching
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# These ARGs are passed in by your Makefile during 'docker build'
ARG TARGET_ARCH=arm64
ARG BINARY_OUT=playstation41_pi

# Build the binary using the name passed from the Makefile
RUN GOOS=linux GOARCH=${TARGET_ARCH} go build -o ${BINARY_OUT} .