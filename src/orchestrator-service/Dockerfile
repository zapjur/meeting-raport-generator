FROM golang:1.23.1

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

RUN apt-get update && apt-get install -y ffmpeg && apt-get clean

COPY . ./

RUN go build -o orchestrator-service

CMD ["./orchestrator-service"]
