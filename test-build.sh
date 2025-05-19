#!/bin/bash
set -e

echo "Generating Go code from protobuf definitions..."
protoc --proto_path=protos \
       --go_out=./app --go_opt=module=github.com/Anthony-Bible/password-exchange/app \
       --go-grpc_out=./app --go-grpc_opt=module=github.com/Anthony-Bible/password-exchange/app \
       protos/database.proto protos/encryption.proto protos/message.proto

if [ $? -ne 0 ]; then
  echo "Protobuf generation failed!"
  exit 1
fi
echo "Protobuf generation successful!"

echo "Testing Go build..."
cd app
go mod tidy
go build -o app
if [ $? -eq 0 ]; then
  echo "Go build successful!"
else
  echo "Go build failed!"
  exit 1
fi

echo "Testing Docker build for main application..."
cd ..
docker build -t passwordexchange-test .
if [ $? -eq 0 ]; then
  echo "Docker build for main application successful!"
  MAIN_IMAGE_SHA=$(docker inspect -f "{{.Id}}" passwordexchange-test)
else
  echo "Docker build for main application failed!"
  exit 1
fi

echo "Testing Docker build for slackbot..."
cd slackbot
docker build -t slackbot-test .
if [ $? -eq 0 ]; then
  echo "Docker build for slackbot successful!"
  SLACKBOT_IMAGE_SHA=$(docker inspect -f "{{.Id}}" slackbot-test)
else
  echo "Docker build for slackbot failed!"
  exit 1
fi

echo "Testing Kubernetes manifest generation..."
cd ..

# Set VERSION and PHASE
if [[ "${GITHUB_REF_TYPE}" =~ "tag" ]]; then
  VERSION="${GITHUB_REF##*/}"
  PHASE="prod"
else
  VERSION=$(git rev-parse HEAD)
  PHASE="dev"
fi

# Combine manifests with dashes between each file
rm -f combined.yaml
first=1
for f in k8s/*.yaml; do
  if [ "$first" -eq 0 ]; then
    printf "\n---\n" >> combined.yaml
  fi
  cat "$f" >> combined.yaml
  first=0
done

sed -i \
  -e "s/%{VERSION}/${VERSION}/g" \
  -e "s/%{PHASE}/${PHASE}/g" \
  -e "s/%{MAIN_IMAGE_SHA}/${MAIN_IMAGE_SHA}/g" \
  -e "s/%{SLACKBOT_IMAGE_SHA}/${SLACKBOT_IMAGE_SHA}/g" \
  combined.yaml
if [ -f combined.yaml ]; then
  echo "Kubernetes manifest generation successful!"
else
  echo "Kubernetes manifest generation failed!"
  exit 1
fi

echo "All tests passed!"
