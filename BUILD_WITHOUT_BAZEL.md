# Building Without Bazel

This document provides instructions for building the Password Exchange project without using Bazel.

## Prerequisites

- Go 1.21 or later
- Docker
- Python 3.11 (for slackbot)
- libmariadb-dev (for database connectivity)

## Building the Go Application

```bash
# Navigate to the app directory
cd app

# Download dependencies
go mod tidy

# Build the application
go build -o app
```

## Building Docker Images

### Main Application

```bash
# From the repository root
docker build -t ghcr.io/anthony-bible/passwordexchange-container-dev:latest .
```

### Slackbot

```bash
# From the repository root
docker build -t ghcr.io/anthony-bible/passwordexchange-slackbot:latest -f slackbot/Dockerfile slackbot/
```

## Pushing Docker Images

```bash
# Login to GitHub Container Registry
docker login ghcr.io -u USERNAME -p TOKEN

# Push images
docker push ghcr.io/anthony-bible/passwordexchange-container-dev:latest
docker push ghcr.io/anthony-bible/passwordexchange-slackbot:latest
```

## Generating Kubernetes Manifests

```bash
# From the repository root
# Set version and phase
export VERSION=$(git rev-parse HEAD)
export PHASE=dev

# If building from a tag
if [[ -n "$TAG" ]]; then
  export VERSION=$TAG
  export PHASE=prod
fi

# Combine manifests and replace variables
cat k8s/*.yaml > combined.yaml
sed -i "s/%{VERSION}/${VERSION}/g" combined.yaml
sed -i "s/%{PHASE}/${PHASE}/g" combined.yaml
```

## Running the Application

### Main Application

```bash
# Run the web component
./app/app web

# Run the email component
./app/app email

# Run the encryption component
./app/app encryption

# Run the database component
./app/app database
```

### Slackbot

```bash
# Navigate to the slackbot directory
cd slackbot

# Install dependencies
pip install -r requirements.txt

# Run the slackbot
python program.py
```

## Testing

You can use the provided test script to verify that everything builds correctly:

```bash
./test-build.sh
```