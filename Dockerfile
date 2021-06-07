From golang:1.16.5-alpine3.13
#COPY go-rewrite /app
#WORKDIR /app
#RUN go mod download 
#RUN go build /app


WORKDIR /go/src/app
COPY go-rewrite/ .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["password.exchange"]
