# Use an argument to toggle platforms (default to arm64)
ARG BUILD_PLATFORM=linux/arm64
FROM --platform=${BUILD_PLATFORM} golang:1.25.1

# Install Ebitengine Linux dependencies
RUN apt-get update && apt-get install -y \
    libgl1-mesa-dev \
    xorg-dev \
    libasound2-dev \
    pkg-config && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Use ARG to define the output name and architecture
ARG TARGET_ARCH=arm64
ARG BINARY_OUT=playstation41_pi64

RUN GOOS=linux GOARCH=${TARGET_ARCH} go build -o ${BINARY_OUT} .