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
else
  echo "Docker build for main application failed!"
  exit 1
fi

echo "Testing Docker build for slackbot..."
cd slackbot
docker build -t slackbot-test .
if [ $? -eq 0 ]; then
  echo "Docker build for slackbot successful!"
else
  echo "Docker build for slackbot failed!"
  exit 1
fi

echo "Testing Kubernetes manifest generation..."
cd ..
source ./tools/bazel_stamp_vars.sh
cat k8s/*.yaml > combined.yaml
sed -i "s/%{VERSION}/${VERSION}/g" combined.yaml
sed -i "s/%{PHASE}/${PHASE}/g" combined.yaml
if [ -f combined.yaml ]; then
  echo "Kubernetes manifest generation successful!"
else
  echo "Kubernetes manifest generation failed!"
  exit 1
fi

echo "All tests passed!"

