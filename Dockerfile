#From golang:1.17.0-alpine3.13
#
##COPY go-rewrite /app
##WORKDIR /app
##RUN go mod download 
##RUN go build /app
## RUN echo $GODEBUG
#
#WORKDIR /go/src/app
#COPY app/ .
#
#RUN go get -d -v ./...
#RUN go install -v ./...
#
#CMD ["app"]
From python:3.8.10-slim-buster
RUN apt update && apt install zstd libssl-dev build-essential curl wget gcc mariadb-client libmariadb-dev clang -y
#RUN apt update && apt install -y curl libssl-dev python3-dev wget zstd gcc build-essential
RUN wget -O /usr/local/bin/bazel https://github.com/bazelbuild/bazelisk/releases/download/v1.11.0/bazelisk-linux-amd64 && chmod +x /usr/local/bin/bazel 
RUN   curl -fsSL https://get.docker.com -o get-docker.sh && DRY_RUN=1 sh ./get-docker.sh  &&  echo -ne "DONE WITH DRY-RUN \n-----------------------\n-----------------------\n-----------------------\n-----------------------\n" && sh ./get-docker.sh
ADD ./ /app
WORKDIR /app
RUN bazel build //k8s:slackbot
#FOR DOCKER LOGIN
