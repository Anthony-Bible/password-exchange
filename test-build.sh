#!/bin/bash
set -e

echo "Generating Go code from protobuf definitions..."
# Check for grpc_python_plugin
#if ! command -v protoc-gen-grpc_python &> /dev/null
#then
#    echo "protoc-gen-grpc_python not found. Please install grpcio-tools:"
#    echo "pip install grpcio-tools"
#    exit 1
#fi

# Generate Go protobuf files to pkg/pb directories
protoc --proto_path=protos \
       --go_out=./app --go_opt=module=github.com/Anthony-Bible/password-exchange/app \
       --go-grpc_out=./app --go-grpc_opt=module=github.com/Anthony-Bible/password-exchange/app \
       protos/database.proto protos/encryption.proto protos/message.proto

# Move generated files to correct pkg/pb locations
mkdir -p ./app/pkg/pb/{database,encryption,message}
mv ./app/databasepb/* ./app/pkg/pb/database/ 2>/dev/null || true
mv ./app/encryptionpb/* ./app/pkg/pb/encryption/ 2>/dev/null || true  
mv ./app/messagepb/* ./app/pkg/pb/message/ 2>/dev/null || true
rmdir ./app/databasepb ./app/encryptionpb ./app/messagepb 2>/dev/null || true

# Generate Python protobuf files for slackbot
protoc --proto_path=protos \
       --python_out=./python_protos \
       --python_grpc_out=./python_protos \
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

echo "Generating Swagger documentation..."
# Check if swag is installed, install if needed
if ! command -v swag &> /dev/null; then
  echo "Installing swag CLI tool..."
  go install github.com/swaggo/swag/cmd/swag@latest
fi

# Generate swagger docs
swag init -g internal/domains/message/adapters/primary/api/docs.go -o docs --parseDependency
if [ $? -eq 0 ]; then
  echo "Swagger documentation generation successful!"
  
  # Fix swagger generation issue with LeftDelim/RightDelim fields
  if grep -q "LeftDelim\|RightDelim" docs/docs.go; then
    echo "Fixing swagger generation compatibility issue..."
    sed -i '/LeftDelim:/d; /RightDelim:/d' docs/docs.go
    # Fix any trailing comma issues
    sed -i 's/SwaggerTemplate:  docTemplate,$/SwaggerTemplate:  docTemplate,/' docs/docs.go
  fi
  
  # Verify generated files contain expected content
  if [ -f "docs/swagger.json" ] && [ -f "docs/swagger.yaml" ] && [ -f "docs/docs.go" ]; then
    if grep -q "Password Exchange API" docs/swagger.json; then
      echo "✅ Swagger files generated and validated successfully!"
    else
      echo "❌ Generated swagger files appear to be incomplete"
      exit 1
    fi
  else
    echo "❌ Required swagger files were not generated"
    exit 1
  fi
  
  echo "Documentation available at /api/v1/docs when server is running"
else
  echo "Swagger documentation generation failed!"
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

docker build -t slackbot-test -f slackbot/Dockerfile .
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
