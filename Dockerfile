# Use the official Golang image
FROM --platform=linux/arm64 golang:1.25.1

# Install Ebitengine Linux dependencies
# These are cached and won't re-run unless this block changes
RUN apt-get update && apt-get install -y \
    libgl1-mesa-dev \
    xorg-dev \
    libasound2-dev \
    pkg-config && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# 1. Copy only the dependency files first
COPY go.mod go.sum ./

# 2. Download dependencies
# This layer is cached unless go.mod or go.sum changes
RUN go mod download

# 3. Now copy the rest of the source code
COPY . .

# 4. Build the binary
# This is the only part that will run frequently
RUN GOOS=linux GOARCH=arm64 go build -o playstation41_pi .