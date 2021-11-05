From golang:1.17.0-alpine3.13

#COPY go-rewrite /app
#WORKDIR /app
#RUN go mod download 
#RUN go build /app
# RUN echo $GODEBUG

WORKDIR /go/src/app
COPY ./ .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["slackbot"]
