FROM golang:1.23.1

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o summary-service

CMD ["./summary-service"]
