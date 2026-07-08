FROM mcr.microsoft.com/devcontainers/go:1.25-bookworm

# Install Ebitengine native Linux dependencies for ARM64 compilation
RUN apt-get update && apt-get install -y \
    libasound2-dev \
    libgl1-mesa-dev \
    libglu1-mesa-dev \
    libxcursor-dev \
    libxi-dev \
    libxinerama-dev \
    libxrandr-dev \
    libxxf86vm-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*